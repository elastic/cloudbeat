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
	Init(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig) (ServiceAPI, error)
}

func newAssetsInventoryWrapper(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig) (*AssetsInventoryWrapper, error) {
	limiter := NewAssetsInventoryRateLimiter(log)
	client, err := asset.NewClient(ctx, append(gcpConfig.ClientOpts, option.WithGRPCDialOption(limiter.GetInterceptorDialOption()))...)
	if err != nil {
		return nil, err
	}

	return &AssetsInventoryWrapper{
		Close: client.Close,
		ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
			return client.ListAssets(ctx, req, append(opts, RetryOnResourceExhausted)...)
		},
	}, nil
}

func (p *ProviderInitializer) Init(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig) (ServiceAPI, error) {
	assetsInventory, err := newAssetsInventoryWrapper(ctx, log, gcpConfig)
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

	resourceAssetsCh := make(chan *assetpb.Asset) // *assetpb.Asset with Resource
	policiesAssetsCh := make(chan *assetpb.Asset) // *assetpb.Asset with IamPolicy
	assetsCh := make(chan *assetpb.Asset)         // *assetpb.Asset with Resource and IamPolicy
	extendedAssetsCh := make(chan *ExtendedGcpAsset)

	go p.fetchAssets(ctx, assetpb.ContentType_RESOURCE, assetTypes, resourceAssetsCh)
	go p.fetchAssets(ctx, assetpb.ContentType_IAM_POLICY, assetTypes, policiesAssetsCh)
	go p.mergeAssets(ctx, resourceAssetsCh, policiesAssetsCh, assetsCh)
	go p.extendAssets(ctx, assetsCh, extendedAssetsCh)

	for asset := range extendedAssetsCh {
		out <- asset
	}
}

func (p *Provider) ListMonitoringAssets(ctx context.Context, out chan<- *MonitoringAsset) {
	defer close(out)

	projectsCh := make(chan *assetpb.Asset)
	extendedAssetCh := make(chan *ExtendedGcpAsset)
	go p.getAllAssets(ctx, p.config.Parent, assetpb.ContentType_RESOURCE, []string{CrmProjectAssetType}, projectsCh)
	go p.extendAssets(ctx, projectsCh, extendedAssetCh)

	for project := range extendedAssetCh {
		logsResourceCh := make(chan *assetpb.Asset)
		alertsResourceCh := make(chan *assetpb.Asset)
		logsAssetCh := make(chan *ExtendedGcpAsset)
		alertsAssetCh := make(chan *ExtendedGcpAsset)

		go p.getAllAssets(ctx, project.Ancestors[0], assetpb.ContentType_RESOURCE, []string{MonitoringLogMetricAssetType}, logsResourceCh)
		go p.getAllAssets(ctx, project.Ancestors[0], assetpb.ContentType_RESOURCE, []string{MonitoringAlertPolicyAssetType}, alertsResourceCh)
		go p.extendAssets(ctx, logsResourceCh, logsAssetCh)
		go p.extendAssets(ctx, alertsResourceCh, alertsAssetCh)

		var logAssets, alertAssets []*ExtendedGcpAsset
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			logAssets = collect(logsAssetCh)
			p.log.Debugf("Listed %d log metrics for %v\n", len(logAssets), project.Name)
		}()
		go func() {
			defer wg.Done()
			alertAssets = collect(alertsAssetCh)
			p.log.Debugf("Listed %d alert policies for %v\n", len(alertAssets), project.Name)
		}()
		wg.Wait()

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

	projectPoliciesCh := make(chan *assetpb.Asset)
	extendedAssetsCh := make(chan *ExtendedGcpAsset)
	policiesCache := &sync.Map{}

	go p.getAllAssets(ctx, p.config.Parent, assetpb.ContentType_IAM_POLICY, []string{CrmProjectAssetType}, projectPoliciesCh)
	go p.extendAssets(ctx, projectPoliciesCh, extendedAssetsCh)

	for asset := range extendedAssetsCh {
		ancestorPolicies := p.getAssetAncestorsPolicies(ctx, asset.Ancestors[1:], policiesCache)

		out <- &ProjectPoliciesAsset{
			CloudAccount: asset.CloudAccount,
			Policies:     append([]*ExtendedGcpAsset{asset}, ancestorPolicies...),
		}
	}
}

func (p *Provider) ListProjectAssets(ctx context.Context, assetTypes []string, out chan<- *ProjectAssets) {
	defer close(out)

	projectResourcesCh := make(chan *assetpb.Asset)
	extendedAssetCh := make(chan *ExtendedGcpAsset)
	go p.getAllAssets(ctx, p.config.Parent, assetpb.ContentType_RESOURCE, []string{CrmProjectAssetType}, projectResourcesCh)
	go p.extendAssets(ctx, projectResourcesCh, extendedAssetCh)

	for project := range extendedAssetCh {
		resourcesCh := make(chan *assetpb.Asset)
		assetsCh := make(chan *ExtendedGcpAsset)
		go p.getAllAssets(ctx, project.Ancestors[0], assetpb.ContentType_RESOURCE, assetTypes, resourcesCh)
		go p.extendAssets(ctx, resourcesCh, assetsCh)

		assets := collect(assetsCh)
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

	dnsResourceCh := make(chan *assetpb.Asset)
	networkResourceCh := make(chan *assetpb.Asset)
	extendedNetworkResourceCh := make(chan *ExtendedGcpAsset)

	go p.getAllAssets(ctx, p.config.Parent, assetpb.ContentType_RESOURCE, []string{DnsPolicyAssetType}, dnsResourceCh)
	go p.getAllAssets(ctx, p.config.Parent, assetpb.ContentType_RESOURCE, []string{ComputeNetworkAssetType}, networkResourceCh)
	go p.extendAssets(ctx, networkResourceCh, extendedNetworkResourceCh)

	dnsPoliciesFields := decodeDnsPolicies(collect(dnsResourceCh))
	for asset := range extendedNetworkResourceCh {
		out <- enrichNetworkAsset(asset, dnsPoliciesFields)
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
	for _, ancestor := range ancestors {
		if value, ok := cache.Load(ancestor); ok {
			if v, ok := value.([]*ExtendedGcpAsset); ok {
				assets = append(assets, v...)
			}
			continue
		}
		prjAncestorPolicyCh := make(chan *assetpb.Asset)
		var assetType string
		if isFolder(ancestor) {
			assetType = CrmFolderAssetType
		}
		if isOrganization(ancestor) {
			assetType = CrmOrgAssetType
		}
		go p.getAllAssets(ctx, ancestor, assetpb.ContentType_IAM_POLICY, []string{assetType}, prjAncestorPolicyCh)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var ancestorPolicies []*ExtendedGcpAsset
			for asset := range prjAncestorPolicyCh {
				ancestorPolicies = append(ancestorPolicies, p.newGcpExtendedAsset(ctx, asset))
			}
			cache.Store(ancestor, ancestorPolicies)
			assets = append(assets, ancestorPolicies...)
		}()
	}
	wg.Wait()
	p.log.Debugf("Listed %d policies for ancestors: %v\n", len(assets), ancestors)
	return assets
}

func (p *Provider) fetchAssets(ctx context.Context, contentType assetpb.ContentType, assetTypes []string, out chan<- *assetpb.Asset) {
	defer close(out)

	wg := sync.WaitGroup{}
	// Fetch each asset type separately to limit failures to a single type
	for _, assetType := range assetTypes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan *assetpb.Asset)
			go p.getAllAssets(ctx, p.config.Parent, contentType, []string{assetType}, ch)
			for asset := range ch {
				out <- asset
			}
		}()
	}
	wg.Wait()
}

func (p *Provider) getAllAssets(ctx context.Context, parent string, contentType assetpb.ContentType, assetTypes []string, out chan<- *assetpb.Asset) {
	defer close(out)

	p.log.Infof("Listing %v assets of types: %v for %v\n", contentType, assetTypes, parent)
	it := p.inventory.ListAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      parent,
		AssetTypes:  assetTypes,
		ContentType: contentType,
	})
	for {
		response, err := it.Next()
		if err == iterator.Done {
			p.log.Infof("Finished fetching GCP %v of types: %v for %v\n", contentType, assetTypes, parent)
			return
		}

		if err != nil {
			p.log.Errorf("Error fetching GCP %v of types: %v for %v: %v\n", contentType, assetTypes, parent, err)
			return
		}

		p.log.Debugf("Fetched GCP %v of type %v: %v\n", contentType, response.AssetType, response.Name)
		out <- response
	}
}

// merge by asset name. send assets with both resource & policy if both channels are open.
// if one channel closes, send remaining assets from the other. finally, flush remaining assets.
//
//revive:disable-next-line
func (p *Provider) mergeAssets(ctx context.Context, resourceCh, policyCh <-chan *assetpb.Asset, out chan<- *assetpb.Asset) {
	defer close(out)

	assetStore := make(map[string]*assetpb.Asset)
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
}

func (p *Provider) extendAssets(ctx context.Context, in <-chan *assetpb.Asset, out chan<- *ExtendedGcpAsset) {
	defer close(out)

	for asset := range in {
		select {
		case <-ctx.Done():
			return
		case out <- p.newGcpExtendedAsset(ctx, asset):
		}
	}
}

func (p *Provider) newGcpExtendedAsset(ctx context.Context, asset *assetpb.Asset) *ExtendedGcpAsset {
	return &ExtendedGcpAsset{
		Asset:        asset,
		CloudAccount: p.crm.GetCloudMetadata(ctx, asset),
	}
}

func mergeAssetContentType(store map[string]*assetpb.Asset, asset *assetpb.Asset) {
	if existing, ok := store[asset.Name]; ok {
		if asset.Resource != nil {
			existing.Resource = asset.Resource
		}
		if asset.IamPolicy != nil {
			existing.IamPolicy = asset.IamPolicy
		}
	} else {
		store[asset.Name] = asset
	}
}

func enrichNetworkAsset(asset *ExtendedGcpAsset, dnsPoliciesFields []*dnsPolicyFields) *ExtendedGcpAsset {
	networkAssetFields := asset.GetResource().GetData().GetFields()
	networkIdentifier := strings.TrimPrefix(asset.GetName(), "//compute.googleapis.com")
	dnsPolicy := findDnsPolicyByNetwork(dnsPoliciesFields, networkIdentifier)

	if dnsPolicy != nil {
		networkAssetFields["enabledDnsLogging"] = &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: dnsPolicy.enableLogging}}
	}
	return asset
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
