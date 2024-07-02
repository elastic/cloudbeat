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
	"fmt"
	"strings"
	"sync"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
)

type Provider struct {
	log                       *logp.Logger
	config                    auth.GcpFactoryConfig
	inventory                 *AssetsInventoryWrapper
	crm                       *ResourceManagerWrapper
	cloudAccountMetadataCache *MapCache[*fetching.CloudAccountMetadata]
}

type AssetsInventoryWrapper struct {
	Close      func() error
	ListAssets func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator
}

type ResourceManagerWrapper struct {
	// returns project display name or an empty string
	getProjectDisplayName func(ctx context.Context, parent string) string

	// returns org display name or an empty string
	getOrganizationDisplayName func(ctx context.Context, parent string) string
}

type MonitoringAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	LogMetrics   []*ExtendedGcpAsset `json:"log_metrics,omitempty"`
	Alerts       []*ExtendedGcpAsset `json:"alerts,omitempty"`
}

type LoggingAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	LogSinks     []*ExtendedGcpAsset `json:"log_sinks,omitempty"`
}

type ProjectPoliciesAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	Policies     []*ExtendedGcpAsset `json:"policies,omitempty"`
}

type ServiceUsageAsset struct {
	CloudAccount *fetching.CloudAccountMetadata
	Services     []*ExtendedGcpAsset `json:"services,omitempty"`
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

type TypeGenerator[T any] func(assets []*ExtendedGcpAsset, projectId, projectName, orgId, orgName string) *T

type Iterator interface {
	Next() (*assetpb.Asset, error)
}

type ServiceAPI interface {
	// ListAllAssetTypesByName List all content types of the given assets types
	ListAllAssetTypesByName(ctx context.Context, assets []string) ([]*ExtendedGcpAsset, error)

	// ListMonitoringAssets List all monitoring assets by project id
	ListMonitoringAssets(ctx context.Context, monitoringAssetTypes map[string][]string) ([]*MonitoringAsset, error)

	// ListLoggingAssets returns a list of logging assets grouped by project id, extended with folder and org level log sinks
	ListLoggingAssets(ctx context.Context) ([]*LoggingAsset, error)

	// ListServiceUsageAssets returns a list of service usage assets grouped by project id
	ListServiceUsageAssets(ctx context.Context) ([]*ServiceUsageAsset, error)

	// returns a project policies for all its ancestors
	ListProjectsAncestorsPolicies(ctx context.Context) ([]*ProjectPoliciesAsset, error)

	// Close the GCP asset client
	Close() error
}

type ProviderInitializerAPI interface {
	// Init initializes the GCP asset client
	Init(ctx context.Context, log *logp.Logger, gcpConfig auth.GcpFactoryConfig) (ServiceAPI, error)
}

func (p *ProviderInitializer) Init(ctx context.Context, log *logp.Logger, gcpConfig auth.GcpFactoryConfig) (ServiceAPI, error) {
	limiter := NewAssetsInventoryRateLimiter(log)
	// initialize GCP assets inventory client
	client, err := asset.NewClient(ctx, append(gcpConfig.ClientOpts, option.WithGRPCDialOption(limiter.GetInterceptorDialOption()))...)
	if err != nil {
		return nil, err
	}
	// wrap the assets inventory client for mocking
	assetsInventoryWrapper := &AssetsInventoryWrapper{
		Close: client.Close,
		ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
			return client.ListAssets(ctx, req, append(opts, RetryOnResourceExhausted)...)
		},
	}

	// initialize GCP resource manager client
	var gcpClientOpt []option.ClientOption
	gcpClientOpt = append(append(gcpClientOpt, option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope)), gcpConfig.ClientOpts...)
	crmService, err := cloudresourcemanager.NewService(ctx, gcpClientOpt...)
	if err != nil {
		return nil, err
	}

	displayNamesCache := NewMapCache[string]()
	// wrap the resource manager client for mocking
	crmServiceWrapper := &ResourceManagerWrapper{
		getProjectDisplayName: func(ctx context.Context, parent string) string {
			return displayNamesCache.Get(func() string {
				prj, err := crmService.Projects.Get(parent).Context(ctx).Do()
				if err != nil {
					log.Errorf("error fetching GCP Project: %s, error: %s", parent, err)
					return ""
				}
				return prj.DisplayName
			}, parent)
		},
		getOrganizationDisplayName: func(ctx context.Context, parent string) string {
			return displayNamesCache.Get(func() string {
				org, err := crmService.Organizations.Get(parent).Context(ctx).Do()
				if err != nil {
					log.Errorf("error fetching GCP Org: %s, error: %s", parent, err)
					return ""
				}
				return org.DisplayName
			}, parent)
		},
	}

	return &Provider{
		config:                    gcpConfig,
		log:                       log,
		inventory:                 assetsInventoryWrapper,
		crm:                       crmServiceWrapper,
		cloudAccountMetadataCache: NewMapCache[*fetching.CloudAccountMetadata](),
	}, nil
}

func (p *Provider) ListAllAssetTypesByName(ctx context.Context, assetTypes []string) ([]*ExtendedGcpAsset, error) {
	wg := sync.WaitGroup{}
	var resourceAssets []*assetpb.Asset
	var policyAssets []*assetpb.Asset

	wg.Add(1)
	go func() {
		resourceAssets = p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
			Parent:      p.config.Parent,
			AssetTypes:  assetTypes,
			ContentType: assetpb.ContentType_RESOURCE,
		})
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		policyAssets = p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
			Parent:      p.config.Parent,
			AssetTypes:  assetTypes,
			ContentType: assetpb.ContentType_IAM_POLICY,
		})
		wg.Done()
	}()

	wg.Wait()

	var assets []*assetpb.Asset
	assets = append(append(assets, resourceAssets...), policyAssets...)
	mergedAssets := mergeAssetContentType(assets)
	extendedAssets := p.extendWithCloudMetadata(ctx, mergedAssets)
	// Enrich network assets with dns policy
	p.enrichNetworkAssets(ctx, extendedAssets)

	return extendedAssets, nil
}

// ListMonitoringAssets returns a list of monitoring assets grouped by project id
func (p *Provider) ListMonitoringAssets(ctx context.Context, monitoringAssetTypes map[string][]string) ([]*MonitoringAsset, error) {
	logMetrics, err := p.ListAllAssetTypesByName(ctx, monitoringAssetTypes["LogMetric"])
	if err != nil {
		return nil, err
	}

	alertPolicies, err := p.ListAllAssetTypesByName(ctx, monitoringAssetTypes["AlertPolicy"])
	if err != nil {
		return nil, err
	}

	typeGenerator := func(assets []*ExtendedGcpAsset, projectId, projectName, orgId, orgName string) *MonitoringAsset {
		return &MonitoringAsset{
			LogMetrics: getAssetsByType(assets, MonitoringLogMetricAssetType),
			Alerts:     getAssetsByType(assets, MonitoringAlertPolicyAssetType),
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        projectId,
				AccountName:      projectName,
				OrganisationId:   orgId,
				OrganizationName: orgName,
			},
		}
	}

	var assets []*ExtendedGcpAsset
	assets = append(append(assets, logMetrics...), alertPolicies...)
	monitoringAssets := getAssetsByProject[MonitoringAsset](assets, p.log, typeGenerator)

	return monitoringAssets, nil
}

// ListLoggingAssets returns a list of logging assets grouped by project id, extended with folder and org level log sinks
func (p *Provider) ListLoggingAssets(ctx context.Context) ([]*LoggingAsset, error) {
	logSinks, err := p.ListAllAssetTypesByName(ctx, []string{LogSinkAssetType})
	if err != nil {
		return nil, err
	}

	typeGenerator := func(assets []*ExtendedGcpAsset, projectId, projectName, orgId, orgName string) *LoggingAsset {
		return &LoggingAsset{
			LogSinks: assets,
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        projectId,
				AccountName:      projectName,
				OrganisationId:   orgId,
				OrganizationName: orgName,
			},
		}
	}

	loggingAssets := getAssetsByProject[LoggingAsset](logSinks, p.log, typeGenerator)
	return loggingAssets, nil
}

// ListServiceUsageAssets returns a list of service usage assets grouped by project id
func (p *Provider) ListServiceUsageAssets(ctx context.Context) ([]*ServiceUsageAsset, error) {
	services, err := p.ListAllAssetTypesByName(ctx, []string{ServiceUsageAssetType})
	if err != nil {
		return nil, err
	}

	typeGenerator := func(assets []*ExtendedGcpAsset, projectId, projectName, orgId, orgName string) *ServiceUsageAsset {
		return &ServiceUsageAsset{
			Services: assets,
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        projectId,
				AccountName:      projectName,
				OrganisationId:   orgId,
				OrganizationName: orgName,
			},
		}
	}

	assets := getAssetsByProject[ServiceUsageAsset](services, p.log, typeGenerator)
	return assets, nil
}

func (p *Provider) Close() error {
	return p.inventory.Close()
}

// enrichNetworkAssets enriches the network assets with dns policy if exists
func (p *Provider) enrichNetworkAssets(ctx context.Context, assets []*ExtendedGcpAsset) {
	networkAssets := getAssetsByType(assets, ComputeNetworkAssetType)
	if len(networkAssets) == 0 {
		p.log.Infof("no %s assets were listed", ComputeNetworkAssetType)
		return
	}
	dnsPolicyAssets := p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      p.config.Parent,
		AssetTypes:  []string{DnsPolicyAssetType},
		ContentType: assetpb.ContentType_RESOURCE,
	})

	if len(dnsPolicyAssets) == 0 {
		p.log.Infof("no %s assets were listed, return original assets", DnsPolicyAssetType)
		return
	}
	dnsPolicies := decodeDnsPolicies(dnsPolicyAssets)

	p.log.Infof("attempting to enrich %d %s assets with dns policy", len(assets), ComputeNetworkAssetType)
	for _, networkAsset := range networkAssets {
		networkAssetFields := networkAsset.GetResource().GetData().GetFields()
		networkIdentifier := strings.TrimPrefix(networkAsset.GetName(), "//compute.googleapis.com")

		dnsPolicy := findDnsPolicyByNetwork(dnsPolicies, networkIdentifier)
		if dnsPolicy != nil {
			p.log.Infof("enrich a %s asset with dns policy, name: %s", ComputeNetworkAssetType, networkIdentifier)
			networkAssetFields["enabledDnsLogging"] = &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: dnsPolicy.enableLogging}}
		}
	}
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

// decodeDnsPolicies gets the required fields from the dns policies assets
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

// getAssetsByProject groups assets by project, extracts metadata for each project, and adds folder and organization-level resources for each group.
func getAssetsByProject[T any](assets []*ExtendedGcpAsset, log *logp.Logger, f TypeGenerator[T]) []*T {
	assetsByProject := lo.GroupBy(assets, func(asset *ExtendedGcpAsset) string { return asset.CloudAccount.AccountId })
	enrichedAssets := make([]*T, 0, len(assetsByProject))
	for projectId, projectAssets := range assetsByProject {
		if projectId == "" {
			continue
		}

		if len(projectAssets) == 0 {
			log.Errorf("no assets were listed for project: %s", projectId)
			continue
		}

		cloudAccount := projectAssets[0].CloudAccount

		// add folder and org level log sinks for each project
		projectAssets = append(projectAssets, assetsByProject[""]...)
		enrichedAssets = append(enrichedAssets, f(
			projectAssets,
			projectId,
			cloudAccount.AccountName,
			cloudAccount.OrganisationId,
			cloudAccount.OrganizationName,
		))
	}
	return enrichedAssets
}

func (p *Provider) getAllAssets(ctx context.Context, request *assetpb.ListAssetsRequest) []*assetpb.Asset {
	p.log.Infof("Listing asset types: %v of type %v for %v", request.AssetTypes, request.ContentType, request.Parent)
	results := make([]*assetpb.Asset, 0)
	it := p.inventory.ListAssets(ctx, request)
	for {
		response, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			p.log.Errorf("Error fetching GCP Asset: %s", err)
			return results
		}

		p.log.Debugf("Fetched GCP Asset: %+v", response.Name)
		results = append(results, response)
	}
	return results
}

func mergeAssetContentType(assets []*assetpb.Asset) []*assetpb.Asset {
	resultsMap := make(map[string]*assetpb.Asset)
	for _, asset := range assets {
		assetKey := asset.Name
		if _, ok := resultsMap[assetKey]; !ok {
			resultsMap[assetKey] = asset
			continue
		}
		item := resultsMap[assetKey]
		if asset.Resource != nil {
			item.Resource = asset.Resource
		}
		if asset.IamPolicy != nil {
			item.IamPolicy = asset.IamPolicy
		}
	}
	results := make([]*assetpb.Asset, 0, len(resultsMap))
	for _, asset := range resultsMap {
		results = append(results, asset)
	}
	return results
}

// extends the assets with the project and organization display name
func (p *Provider) extendWithCloudMetadata(ctx context.Context, assets []*assetpb.Asset) []*ExtendedGcpAsset {
	extendedAssets := make([]*ExtendedGcpAsset, 0, len(assets))
	for _, asset := range assets {
		orgId := getOrganizationId(asset.Ancestors)
		projectId := getProjectId(asset.Ancestors)
		cacheKey := fmt.Sprintf("%s/%s", projectId, orgId)
		cloudAccount := p.cloudAccountMetadataCache.Get(func() *fetching.CloudAccountMetadata {
			return p.getCloudAccountMetadata(ctx, projectId, orgId)
		}, cacheKey)
		extendedAssets = append(extendedAssets, &ExtendedGcpAsset{
			Asset:        asset,
			CloudAccount: cloudAccount,
		})
	}
	return extendedAssets
}

func (p *Provider) ListProjectsAncestorsPolicies(ctx context.Context) ([]*ProjectPoliciesAsset, error) {
	projects := p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
		ContentType: assetpb.ContentType_IAM_POLICY,
		Parent:      p.config.Parent,
		AssetTypes:  []string{CrmProjectAssetType},
	})
	p.log.Infof("Listed %d GCP projects", len(projects))
	ancestorsPoliciesCache := NewMapCache[[]*ExtendedGcpAsset]()
	return lo.Map(projects, func(project *assetpb.Asset, _ int) *ProjectPoliciesAsset {
		projectAsset := p.extendWithCloudMetadata(ctx, []*assetpb.Asset{project})[0]
		// Skip first ancestor it as we already got it
		policiesAssets := append([]*ExtendedGcpAsset{projectAsset}, getAncestorsAssets(ctx, ancestorsPoliciesCache, p, project.Ancestors[1:])...)
		return &ProjectPoliciesAsset{CloudAccount: projectAsset.CloudAccount, Policies: policiesAssets}
	}), nil
}

func getAncestorsAssets(ctx context.Context, ancestorsPoliciesCache *MapCache[[]*ExtendedGcpAsset], p *Provider, ancestors []string) []*ExtendedGcpAsset {
	return lo.Flatten(lo.Map(ancestors, func(parent string, _ int) []*ExtendedGcpAsset {
		return ancestorsPoliciesCache.Get(func() []*ExtendedGcpAsset {
			var assetType string
			if strings.HasPrefix(parent, "folders") {
				assetType = CrmFolderAssetType
			}
			if strings.HasPrefix(parent, "organizations") {
				assetType = CrmOrgAssetType
			}
			return p.extendWithCloudMetadata(ctx, p.getAllAssets(ctx, &assetpb.ListAssetsRequest{
				ContentType: assetpb.ContentType_IAM_POLICY,
				Parent:      parent,
				AssetTypes:  []string{assetType},
			}))
		}, parent)
	}))
}

func (p *Provider) getCloudAccountMetadata(ctx context.Context, projectId string, orgId string) *fetching.CloudAccountMetadata {
	var orgName string
	var projectName string
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		orgName = p.crm.getOrganizationDisplayName(ctx, fmt.Sprintf("organizations/%s", orgId))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		// some assets are not associated with a project
		if projectId != "" {
			projectName = p.crm.getProjectDisplayName(ctx, fmt.Sprintf("projects/%s", projectId))
		}
		wg.Done()
	}()
	wg.Wait()
	return &fetching.CloudAccountMetadata{
		AccountId:        projectId,
		AccountName:      projectName,
		OrganisationId:   orgId,
		OrganizationName: orgName,
	}
}

func getOrganizationId(ancestors []string) string {
	last := ancestors[len(ancestors)-1]
	parts := strings.Split(last, "/") // organizations/1234567890

	if parts[0] == "organizations" {
		return parts[1]
	}

	return ""
}

func getProjectId(ancestors []string) string {
	parts := strings.Split(ancestors[0], "/") // projects/1234567890

	if parts[0] == "projects" {
		return parts[1]
	}

	return ""
}

func getAssetsByType(projectAssets []*ExtendedGcpAsset, assetType string) []*ExtendedGcpAsset {
	return lo.Filter(projectAssets, func(asset *ExtendedGcpAsset, _ int) bool {
		return asset.AssetType == assetType
	})
}
