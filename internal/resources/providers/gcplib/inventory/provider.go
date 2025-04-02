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

	// wrap the assets inventory client for mocking
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

	resourceCh := make(chan *assetpb.Asset) // *assetpb.Asset with Resource
	policyCh := make(chan *assetpb.Asset)   // *assetpb.Asset with IamPolicy
	mergeCh := make(chan *assetpb.Asset)    // *assetpb.Asset with Resource and IamPolicy
	enrichCh := make(chan *ExtendedGcpAsset)

	go p.getResources(ctx, assetTypes, resourceCh)
	go p.getPolicies(ctx, assetTypes, policyCh)
	go p.mergeAssets(ctx, mergeCh, resourceCh, policyCh)
	go p.enrichAssets(ctx, mergeCh, enrichCh)

	for asset := range enrichCh {
		out <- asset
	}
}

func (p *Provider) ListMonitoringAssets(ctx context.Context, out chan<- *MonitoringAsset) {
	defer close(out)

	logsResourceCh := make(chan *assetpb.Asset)
	alertsResourceCh := make(chan *assetpb.Asset)
	logsAssetCh := make(chan *ExtendedGcpAsset)
	alertsAssetCh := make(chan *ExtendedGcpAsset)

	go p.getResources(ctx, []string{MonitoringLogMetricAssetType}, logsResourceCh)
	go p.getResources(ctx, []string{MonitoringAlertPolicyAssetType}, alertsResourceCh)
	go p.enrichAssets(ctx, logsResourceCh, logsAssetCh)
	go p.enrichAssets(ctx, alertsResourceCh, alertsAssetCh)

	var logAssets, alertAssets []*ExtendedGcpAsset
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		logAssets = collect(logsAssetCh)
		p.log.Debugf("Listed %d log metrics\n", len(logAssets))
	}()
	go func() {
		defer wg.Done()
		alertAssets = collect(alertsAssetCh)
		p.log.Debugf("Listed %d alert policies\n", len(alertAssets))
	}()
	wg.Wait()

	assetsByProject := lo.GroupBy(append(logAssets, alertAssets...), func(asset *ExtendedGcpAsset) string { return asset.CloudAccount.AccountId })
	for _, projectAssets := range assetsByProject {
		out <- &MonitoringAsset{
			LogMetrics:   getAssetsByType(projectAssets, MonitoringLogMetricAssetType),
			Alerts:       getAssetsByType(projectAssets, MonitoringAlertPolicyAssetType),
			CloudAccount: projectAssets[0].CloudAccount,
		}
	}
}

func (p *Provider) ListProjectsAncestorsPolicies(ctx context.Context, out chan<- *ProjectPoliciesAsset) {
	defer close(out)

	prjCh := make(chan *assetpb.Asset)
	prjMetadataCh := make(chan *ExtendedGcpAsset)
	policiesCache := &sync.Map{}

	go p.getPolicies(ctx, []string{CrmProjectAssetType}, prjCh)
	go p.enrichAssets(ctx, prjCh, prjMetadataCh)

	for asset := range prjMetadataCh {
		out <- &ProjectPoliciesAsset{
			CloudAccount: asset.CloudAccount,
			Policies:     append([]*ExtendedGcpAsset{asset}, p.getAssetAncestorsPolicies(ctx, asset.Ancestors[1:], policiesCache)...),
		}
	}
}

func (p *Provider) ListProjectAssets(ctx context.Context, assetTypes []string, out chan<- *ProjectAssets) {
	defer close(out)

	resourcesCh := make(chan *assetpb.Asset)
	assetsCh := make(chan *ExtendedGcpAsset)
	go p.getResources(ctx, assetTypes, resourcesCh)
	go p.enrichAssets(ctx, resourcesCh, assetsCh)

	assets := collect(assetsCh)
	p.log.Debugf("Listed %d resources for %v\n", len(assets), assetTypes)

	assetsByProject := lo.GroupBy(assets, func(asset *ExtendedGcpAsset) string { return asset.CloudAccount.AccountId })
	for _, projectAssets := range assetsByProject {
		out <- &ProjectAssets{
			Assets:       projectAssets,
			CloudAccount: projectAssets[0].CloudAccount,
		}
	}
}

func (p *Provider) ListNetworkAssets(ctx context.Context, out chan<- *ExtendedGcpAsset) {
	defer close(out)

	dnsResourceCh := make(chan *assetpb.Asset)
	networkResourceCh := make(chan *assetpb.Asset)
	extendedNetworkResourceCh := make(chan *ExtendedGcpAsset)

	go p.getResources(ctx, []string{DnsPolicyAssetType}, dnsResourceCh)
	go p.getResources(ctx, []string{ComputeNetworkAssetType}, networkResourceCh)
	go p.enrichAssets(ctx, networkResourceCh, extendedNetworkResourceCh)

	dnsPoliciesFields := decodeDnsPolicies(collect(dnsResourceCh))
	for asset := range extendedNetworkResourceCh {
		out <- enrichNetworkAsset(asset, dnsPoliciesFields)
	}
}

func (p *Provider) Close() error {
	return p.inventory.Close()
}

// TODO: when to call this method?
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
		go p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
			ContentType: assetpb.ContentType_IAM_POLICY,
			Parent:      ancestor,
			AssetTypes:  []string{assetType},
		}, prjAncestorPolicyCh)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var ancestorPolicies []*ExtendedGcpAsset
			for asset := range prjAncestorPolicyCh {
				ancestorPolicies = append(ancestorPolicies, p.extendWithCloudMetadata(ctx, asset)) // TODO: we don't need to extend
			}
			cache.Store(ancestor, ancestorPolicies)
			assets = append(assets, ancestorPolicies...)
		}()
	}
	wg.Wait()
	p.log.Debugf("Listed %d policies for ancestors: %v\n", len(assets), ancestors)
	return assets
}

func (p *Provider) getAllAssets(ctx context.Context, request *assetpb.ListAssetsRequest, out chan<- *assetpb.Asset) {
	defer close(out)

	p.log.Infof("Listing %v assets of types: %v for %v\n", request.ContentType, request.AssetTypes, request.Parent)
	it := p.inventory.ListAssets(ctx, request)
	for {
		response, err := it.Next()
		if err == iterator.Done {
			p.log.Infof("Finished fetching GCP %v for %v\n", request.ContentType, request.AssetTypes)
			return
		}

		if err != nil {
			p.log.Errorf("Error fetching GCP %v: %v\n", request.ContentType, err)
			return
		}

		p.log.Debugf("Fetched GCP %v of type %v: %v\n", request.ContentType, response.AssetType, response.Name)
		out <- response
	}
}

// merge by asset name. send assets with both resource & policy if both channels are open.
// if one channel closes, send remaining assets from the other. finally, flush remaining assets.
//
//revive:disable-next-line
func (p *Provider) mergeAssets(ctx context.Context, out chan<- *assetpb.Asset, resourceCh, policyCh <-chan *assetpb.Asset) {
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

func (p *Provider) enrichAssets(ctx context.Context, in <-chan *assetpb.Asset, out chan<- *ExtendedGcpAsset) {
	defer close(out)

	for asset := range in {
		select {
		case <-ctx.Done():
			return
		case out <- p.extendWithCloudMetadata(ctx, asset):
		}
	}
}

func (p *Provider) getResources(ctx context.Context, assetTypes []string, out chan<- *assetpb.Asset) {
	p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      p.config.Parent,
		AssetTypes:  assetTypes,
		ContentType: assetpb.ContentType_RESOURCE,
	}, out)
}

func (p *Provider) getPolicies(ctx context.Context, assetTypes []string, out chan<- *assetpb.Asset) {
	p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      p.config.Parent,
		AssetTypes:  assetTypes,
		ContentType: assetpb.ContentType_IAM_POLICY,
	}, out)
}

// extends the assets with the project and organization display name
func (p *Provider) extendWithCloudMetadata(ctx context.Context, asset *assetpb.Asset) *ExtendedGcpAsset {
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
		// TODO: avoid updating the raw asset
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

func getAssetsByType(projectAssets []*ExtendedGcpAsset, assetType string) []*ExtendedGcpAsset {
	return lo.Filter(projectAssets, func(asset *ExtendedGcpAsset, _ int) bool { return asset.AssetType == assetType })
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
