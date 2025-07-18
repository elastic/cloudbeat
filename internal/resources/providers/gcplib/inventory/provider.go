// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package inventory

import (
	"context"
	"strings"
	"sync"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
)

type Provider struct {
	log       *clog.Logger
	config    auth.GcpFactoryConfig
	inventory *AssetsInventoryWrapper
	crm       *ResourceManagerWrapper
}

type AssetsInventoryWrapper struct {
	Close      func() error
	ListAssets func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator
}

type MonitoringAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	LogMetrics   []*ExtendedGcpAsset `json:"log_metrics,omitempty"`
	Alerts       []*ExtendedGcpAsset `json:"alerts,omitempty"`
}

type ProjectPoliciesAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	Policies     []*ExtendedGcpAsset `json:"policies,omitempty"`
}

type ProjectAssets struct {
	CloudAccount *fetching.CloudAccountMetadata
	Assets       []*ExtendedGcpAsset
}

type ExtendedGcpAsset struct {
	*assetpb.Asset
	CloudAccount *fetching.CloudAccountMetadata
}

type ProviderInitializer struct{}

type dnsPolicyFields struct {
	networks      []string
	enableLogging bool
}

type Iterator interface {
	Next() (*assetpb.Asset, error)
}

type ServiceAPI interface {
	ListAssetTypes(ctx context.Context, assetTypes []string, out chan<- *ExtendedGcpAsset)
	ListMonitoringAssets(ctx context.Context, out chan<- *MonitoringAsset)
	ListProjectsAncestorsPolicies(ctx context.Context, out chan<- *ProjectPoliciesAsset)
	ListProjectAssets(ctx context.Context, assetTypes []string, out chan<- *ProjectAssets)
	ListNetworkAssets(ctx context.Context, out chan<- *ExtendedGcpAsset)
	Clear()
	Close() error
}

type ProviderInitializerAPI interface {
	Init(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig, cfg config.GcpConfig) (ServiceAPI, error)
}

func newAssetsInventoryWrapper(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig, cfg config.GcpConfig) (*AssetsInventoryWrapper, error) {
	limiter := NewAssetsInventoryRateLimiter(log)
	client, err := asset.NewClient(ctx, append(gcpConfig.ClientOpts, option.WithGRPCDialOption(limiter.GetInterceptorDialOption()))...)
	if err != nil {
		return nil, err
	}

	return &AssetsInventoryWrapper{
		Close: client.Close,
		ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
			if req.PageSize == 0 {
				req.PageSize = cfg.GcpCallOpt.ListAssetsPageSize
			}
			return client.ListAssets(ctx, req, append(opts, GAXCallOptionRetrier(log), gax.WithTimeout(cfg.GcpCallOpt.ListAssetsTimeout))...)
		},
	}, nil
}

func (p *ProviderInitializer) Init(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig, cfg config.GcpConfig) (ServiceAPI, error) {
	assetsInventory, err := newAssetsInventoryWrapper(ctx, log, gcpConfig, cfg)
	if err != nil {
		return nil, err
	}

	cloudResourceManager, err := NewResourceManagerWrapper(ctx, log, gcpConfig)
	if err != nil {
		return nil, err
	}

	return &Provider{
		config:    gcpConfig,
		log:       log,
		inventory: assetsInventory,
		crm:       cloudResourceManager,
	}, nil
}

func (p *Provider) ListAssetTypes(ctx context.Context, assetTypes []string, out chan<- *ExtendedGcpAsset) {
	defer close(out)

	resourceAssetsCh := p.fetchAssets(ctx, assetpb.ContentType_RESOURCE, assetTypes)
	policiesAssetsCh := p.fetchAssets(ctx, assetpb.ContentType_IAM_POLICY, assetTypes)
	assetsCh := p.mergeAssets(ctx, resourceAssetsCh, policiesAssetsCh)

	for gcpAsset := range assetsCh {
		select {
		case <-ctx.Done():
			return
		case out <- gcpAsset:
		}
	}
}

func (p *Provider) ListMonitoringAssets(ctx context.Context, out chan<- *MonitoringAsset) {
	defer close(out)

	projectsCh := p.getParentResources(ctx, p.config.Parent, []string{CrmProjectAssetType})

	for project := range projectsCh {
		if ctx.Err() != nil {
			return
		}
		logsAssetCh := p.getParentResources(ctx, project.Ancestors[0], []string{MonitoringLogMetricAssetType})
		alertsAssetCh := p.getParentResources(ctx, project.Ancestors[0], []string{MonitoringAlertPolicyAssetType})

		var logAssets, alertAssets []*ExtendedGcpAsset
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			logAssets = collect(logsAssetCh)
		}()
		go func() {
			defer wg.Done()
			alertAssets = collect(alertsAssetCh)
		}()
		wg.Wait()

		p.log.Debugf("Listed %d log metrics and %d alert policies for %v\n", len(logAssets), len(alertAssets), project.Name)
		if len(logAssets) == 0 && len(alertAssets) == 0 {
			continue
		}
		out <- &MonitoringAsset{
			LogMetrics:   logAssets,
			Alerts:       alertAssets,
			CloudAccount: project.CloudAccount,
		}
	}
}

func (p *Provider) ListProjectsAncestorsPolicies(ctx context.Context, out chan<- *ProjectPoliciesAsset) {
	defer close(out)

	projectsCh := make(chan *ExtendedGcpAsset)
	policiesCache := &sync.Map{}

	go p.fetchAllAssets(ctx, projectsCh, &assetpb.ListAssetsRequest{
		Parent:      p.config.Parent,
		AssetTypes:  []string{CrmProjectAssetType},
		ContentType: assetpb.ContentType_IAM_POLICY,
	})

	for projectAsset := range projectsCh {
		select {
		case <-ctx.Done():
			return
		case out <- &ProjectPoliciesAsset{
			CloudAccount: projectAsset.CloudAccount,
			Policies:     append([]*ExtendedGcpAsset{projectAsset}, p.getAssetAncestorsPolicies(ctx, projectAsset.Ancestors[1:], policiesCache)...),
		}:
		}
	}
}

func (p *Provider) ListProjectAssets(ctx context.Context, assetTypes []string, out chan<- *ProjectAssets) {
	defer close(out)

	projectsCh := p.getParentResources(ctx, p.config.Parent, []string{CrmProjectAssetType})

	for project := range projectsCh {
		if ctx.Err() != nil {
			return
		}
		assets := collect(p.getParentResources(ctx, project.Ancestors[0], assetTypes))
		p.log.Debugf("Listed %d resources of type: %v in %v\n", len(assets), assetTypes, project.Name)
		if len(assets) == 0 {
			continue
		}
		out <- &ProjectAssets{
			Assets:       assets,
			CloudAccount: project.CloudAccount,
		}
	}
}

func (p *Provider) ListNetworkAssets(ctx context.Context, out chan<- *ExtendedGcpAsset) {
	defer close(out)

	dnsPolicyAssetCh := p.getParentResources(ctx, p.config.Parent, []string{DnsPolicyAssetType})
	networkAssetCh := p.getParentResources(ctx, p.config.Parent, []string{ComputeNetworkAssetType})

	dnsPoliciesFields := decodeDnsPolicies(lo.Map(collect(dnsPolicyAssetCh), func(policyAsset *ExtendedGcpAsset, _ int) *assetpb.Asset { return policyAsset.Asset }))
	for networkAsset := range networkAssetCh {
		select {
		case <-ctx.Done():
			return
		case out <- enrichNetworkAsset(networkAsset, dnsPoliciesFields):
		}
	}
}

func (p *Provider) Close() error {
	return p.inventory.Close()
}

func (p *Provider) Clear() {
	p.crm.Clear()
}

func (p *Provider) getAssetAncestorsPolicies(ctx context.Context, ancestors []string, cache *sync.Map) []*ExtendedGcpAsset {
	wg := sync.WaitGroup{}
	var assets []*ExtendedGcpAsset
	mu := sync.Mutex{}
	for _, ancestor := range ancestors {
		if value, ok := cache.Load(ancestor); ok {
			if v, ok := value.([]*ExtendedGcpAsset); ok {
				mu.Lock()
				assets = append(assets, v...)
				mu.Unlock()
			}
			continue
		}
		prjAncestorPolicyCh := make(chan *ExtendedGcpAsset)
		var assetType string
		if isFolder(ancestor) {
			assetType = CrmFolderAssetType
		}
		if isOrganization(ancestor) {
			assetType = CrmOrgAssetType
		}
		go p.fetchAllAssets(ctx, prjAncestorPolicyCh, &assetpb.ListAssetsRequest{
			Parent:      ancestor,
			AssetTypes:  []string{assetType},
			ContentType: assetpb.ContentType_IAM_POLICY,
		})

		wg.Add(1)
		go func() {
			defer wg.Done()
			ancestorPolicies := collect(prjAncestorPolicyCh)
			cache.Store(ancestor, ancestorPolicies)
			mu.Lock()
			assets = append(assets, ancestorPolicies...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	p.log.Debugf("Listed %d policies for ancestors: %v\n", len(assets), ancestors)
	return assets
}

func (p *Provider) getParentResources(ctx context.Context, parent string, assetTypes []string) <-chan *ExtendedGcpAsset {
	ch := make(chan *ExtendedGcpAsset)
	go p.fetchAllAssets(ctx, ch, &assetpb.ListAssetsRequest{
		Parent:      parent,
		AssetTypes:  assetTypes,
		ContentType: assetpb.ContentType_RESOURCE,
	})
	return ch
}

func (p *Provider) fetchAssets(ctx context.Context, contentType assetpb.ContentType, assetTypes []string) <-chan *ExtendedGcpAsset {
	out := make(chan *ExtendedGcpAsset)
	wg := sync.WaitGroup{}
	// Fetch each asset type separately to limit failures to a single type
	for _, assetType := range assetTypes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan *ExtendedGcpAsset)
			go p.fetchAllAssets(ctx, ch, &assetpb.ListAssetsRequest{
				Parent:      p.config.Parent,
				AssetTypes:  []string{assetType},
				ContentType: contentType,
			})
			for asset := range ch {
				out <- asset
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (p *Provider) fetchAllAssets(ctx context.Context, out chan<- *ExtendedGcpAsset, req *assetpb.ListAssetsRequest) {
	defer close(out)

	p.log.Infof("Listing %v assets of types: %v for %v\n", req.ContentType, req.AssetTypes, req.Parent)
	it := p.inventory.ListAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      req.Parent,
		AssetTypes:  req.AssetTypes,
		ContentType: req.ContentType,
	})
	for {
		response, err := it.Next()
		if err == iterator.Done {
			p.log.Infof("Finished fetching GCP %v of types: %v for %v\n", req.ContentType, req.AssetTypes, req.Parent)
			return
		}

		if err != nil {
			p.log.Errorf("Error fetching GCP %v of types: %v for %v: %v\n", req.ContentType, req.AssetTypes, req.Parent, err)
			return
		}

		p.log.Debugf("Fetched GCP %v of type %v: %v\n", req.ContentType, response.AssetType, response.Name)
		out <- p.newGcpExtendedAsset(ctx, response)
	}
}

// merge by asset name. send assets with both resource & policy if both channels are open.
// if one channel closes, send remaining assets from the other. finally, flush remaining assets.
//
//revive:disable-next-line
func (p *Provider) mergeAssets(ctx context.Context, resourceCh, policyCh <-chan *ExtendedGcpAsset) <-chan *ExtendedGcpAsset {
	out := make(chan *ExtendedGcpAsset)

	go func() {
		defer close(out)
		assetStore := make(map[string]*ExtendedGcpAsset)
		rch, pch := resourceCh, policyCh
		for rch != nil || pch != nil || len(assetStore) > 0 {
			select {
			case <-ctx.Done():
				return
			case asset, ok := <-rch:
				if ok {
					mergeAssetContentType(assetStore, asset)
				} else {
					rch = nil
				}
			case asset, ok := <-pch:
				if ok {
					mergeAssetContentType(assetStore, asset)
				} else {
					pch = nil
				}
			}
			for id, a := range assetStore {
				hasPolicy := a.IamPolicy != nil
				hasResource := a.Resource != nil
				hasBoth := hasPolicy && hasResource
				if hasBoth || (rch == nil && hasPolicy) || (pch == nil && hasResource) {
					out <- a
					delete(assetStore, id)
				}
			}
		}
	}()

	return out
}

func (p *Provider) newGcpExtendedAsset(ctx context.Context, gcpAsset *assetpb.Asset) *ExtendedGcpAsset {
	return &ExtendedGcpAsset{
		Asset:        gcpAsset,
		CloudAccount: p.crm.GetCloudMetadata(ctx, gcpAsset),
	}
}

func mergeAssetContentType(store map[string]*ExtendedGcpAsset, extendedAsset *ExtendedGcpAsset) {
	if existing, ok := store[extendedAsset.Name]; ok {
		if extendedAsset.Resource != nil {
			existing.Resource = extendedAsset.Resource
		}
		if extendedAsset.IamPolicy != nil {
			existing.IamPolicy = extendedAsset.IamPolicy
		}
	} else {
		store[extendedAsset.Name] = extendedAsset
	}
}

func enrichNetworkAsset(networkAsset *ExtendedGcpAsset, dnsPoliciesFields []*dnsPolicyFields) *ExtendedGcpAsset {
	networkAssetFields := networkAsset.GetResource().GetData().GetFields()
	networkIdentifier := strings.TrimPrefix(networkAsset.GetName(), "//compute.googleapis.com")
	dnsPolicy := findDnsPolicyByNetwork(dnsPoliciesFields, networkIdentifier)

	if dnsPolicy != nil {
		networkAssetFields["enabledDnsLogging"] = &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: dnsPolicy.enableLogging}}
	}
	return networkAsset
}

// findDnsPolicyByNetwork finds DNS policy by network identifier
func findDnsPolicyByNetwork(dnsPolicies []*dnsPolicyFields, networkIdentifier string) *dnsPolicyFields {
	for _, dnsPolicy := range dnsPolicies {
		if lo.SomeBy(dnsPolicy.networks, func(networkUrl string) bool {
			return strings.HasSuffix(networkUrl, networkIdentifier)
		}) {
			return dnsPolicy
		}
	}
	return nil
}

func decodeDnsPolicies(dnsPolicyAssets []*assetpb.Asset) []*dnsPolicyFields {
	dnsPolicies := make([]*dnsPolicyFields, 0)
	for _, dnsPolicyAsset := range dnsPolicyAssets {
		fields := new(dnsPolicyFields)
		dnsPolicyData := dnsPolicyAsset.GetResource().GetData().GetFields()

		if attachedNetworks, exist := dnsPolicyData["networks"]; exist {
			networks := attachedNetworks.GetListValue().GetValues()
			for _, network := range networks {
				if networkUrl, found := network.GetStructValue().GetFields()["networkUrl"]; found {
					fields.networks = append(fields.networks, networkUrl.GetStringValue())
				}
			}
		}

		if enableLogging, exist := dnsPolicyData["enableLogging"]; exist {
			fields.enableLogging = enableLogging.GetBoolValue()
		}

		dnsPolicies = append(dnsPolicies, fields)
	}

	return dnsPolicies
}

func isFolder(parent string) bool {
	return strings.HasPrefix(parent, "folders")
}

func isOrganization(parent string) bool {
	return strings.HasPrefix(parent, "organizations")
}

func collect[T any](ch <-chan T) []T {
	res := make([]T, 0)
	for item := range ch {
		res = append(res, item)
	}
	return res
}
