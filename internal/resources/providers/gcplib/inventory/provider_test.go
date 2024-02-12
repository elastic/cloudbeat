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
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
)

type ProviderTestSuite struct {
	suite.Suite
	logger          *logp.Logger
	mockedInventory *AssetsInventoryWrapper
	mockedIterator  *MockIterator
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {
	s.logger = logp.NewLogger("test")
	s.mockedIterator = new(MockIterator)
	s.mockedInventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, _ *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			return s.mockedIterator
		},
	}
}

func (s *ProviderTestSuite) TestProviderInit() {
	initMock := new(MockProviderInitializerAPI)
	gcpConfig := auth.GcpFactoryConfig{
		Parent:     "projects/1",
		ClientOpts: []option.ClientOption{},
	}

	initMock.EXPECT().Init(mock.Anything, s.logger, gcpConfig).Return(&Provider{}, nil).Once()
	provider, err := initMock.Init(context.Background(), s.logger, gcpConfig)
	s.Require().NoError(err)
	s.NotNil(provider)
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	provider := &Provider{
		log:       s.logger,
		inventory: s.mockedInventory,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, _ string) string {
				return "ProjectName"
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName2", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	value, err := provider.ListAllAssetTypesByName(context.Background(), []string{"test"})
	s.Require().NoError(err)

	// test merging assets with same name:
	assetNames := lo.Map(value, func(asset *ExtendedGcpAsset, _ int) string { return asset.Name })
	resourceAssets := lo.Filter(value, func(asset *ExtendedGcpAsset, _ int) bool { return asset.Resource != nil })
	policyAssets := lo.Filter(value, func(asset *ExtendedGcpAsset, _ int) bool { return asset.IamPolicy != nil })
	s.True(lo.Contains(assetNames, "AssetName1"))
	s.Len(resourceAssets, 2) // 2 assets with resources (assetName1, assetName2)
	s.Len(policyAssets, 1)   // 1 asset with policy 	(assetName1)
	s.Len(value, 2)          // 2 assets in total 		(assetName1 merged resource/policy, assetName2)

	// tests extending assets with display names for org/prj:
	projectNames := lo.UniqBy(value, func(asset *ExtendedGcpAsset) string { return asset.Ecs.ProjectName })
	orgNames := lo.UniqBy(value, func(asset *ExtendedGcpAsset) string { return asset.Ecs.OrganizationName })
	s.Len(projectNames, 1)
	s.Equal("ProjectName", projectNames[0].Ecs.ProjectName)
	s.Len(orgNames, 1)
	s.Equal("OrganizationName", orgNames[0].Ecs.OrganizationName)
}

func (s *ProviderTestSuite) TestListMonitoringAssets() {
	provider := &Provider{
		log:       s.logger,
		inventory: s.mockedInventory,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, parent string) string {
				if parent == "projects/1" {
					return "ProjectName1"
				}
				return "ProjectName2"
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName1"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	expected := []*MonitoringAsset{
		{
			LogMetrics: []*ExtendedGcpAsset{
				{
					Asset: &assetpb.Asset{Name: "LogMetric1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: MonitoringLogMetricAssetType},
					Ecs:   &fetching.EcsGcp{ProjectId: "1", ProjectName: "ProjectName1", OrganizationId: "1", OrganizationName: "OrganizationName1"},
				},
			},
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "1",
				ProjectName:      "ProjectName1",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			Alerts: make([]*ExtendedGcpAsset, 0, 1),
		},
		{
			LogMetrics: make([]*ExtendedGcpAsset, 0, 1),
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "2",
				ProjectName:      "ProjectName2",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			Alerts: []*ExtendedGcpAsset{
				{
					Asset: &assetpb.Asset{Name: "AlertPolicy1", Resource: nil, IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: MonitoringAlertPolicyAssetType},
					Ecs: &fetching.EcsGcp{
						ProjectId:        "2",
						ProjectName:      "ProjectName2",
						OrganizationId:   "1",
						OrganizationName: "OrganizationName1",
					},
				},
			},
		},
	}

	//  AssetType: "logging.googleapis.com/LogMetric"}
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetric1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetric1", IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	//  AssetType: "monitoring.googleapis.com/AlertPolicy"}
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicy1", Resource: nil, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicy1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	monitoringAssetTypes := map[string][]string{
		"LogMetric":   {MonitoringLogMetricAssetType},
		"AlertPolicy": {MonitoringAlertPolicyAssetType},
	}

	values, err := provider.ListMonitoringAssets(context.Background(), monitoringAssetTypes)

	s.Require().NoError(err)
	s.ElementsMatch(expected, values)
}

func (s *ProviderTestSuite) TestEnrichNetworkAssets() {
	provider := &Provider{
		log:       s.logger,
		inventory: s.mockedInventory,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, _ string) string {
				return "ProjectName"
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	assets := []*ExtendedGcpAsset{
		{
			Asset: &assetpb.Asset{
				Name:      "//compute.googleapis.com/projects/test-project/global/networks/test-network-1",
				AssetType: ComputeNetworkAssetType,
				Resource:  &assetpb.Resource{Data: &structpb.Struct{Fields: map[string]*structpb.Value{}}},
				Ancestors: []string{"projects/1", "organizations/1"},
			},
		},
		{
			Asset: &assetpb.Asset{Name: "//compute.googleapis.com/projects/test-project/global/networks/test-network-2", AssetType: ComputeNetworkAssetType, Resource: &assetpb.Resource{
				Data: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": {Kind: &structpb.Value_StringValue{StringValue: "network2"}},
					},
				},
			}, Ancestors: []string{"projects/1", "organizations/1"}},
		},
		{
			Asset: &assetpb.Asset{Name: "//compute.googleapis.com/projects/test-project/global/networks/test-network-3", AssetType: ComputeNetworkAssetType, Resource: &assetpb.Resource{
				Data: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": {Kind: &structpb.Value_StringValue{StringValue: "network2"}},
					},
				},
			}, Ancestors: []string{"projects/1", "organizations/1"}},
		},
	}

	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "dnsPolicyAsset", Resource: &assetpb.Resource{
		Data: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"networks": {Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: []*structpb.Value{
					{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{"networkUrl": {Kind: &structpb.Value_StringValue{StringValue: "https://compute.googleapis.com/compute/v1/projects/test-project/global/networks/test-network-1"}}},
					}}},
					{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{"networkUrl": {Kind: &structpb.Value_StringValue{StringValue: "https://compute.googleapis.com/compute/v1/projects/test-project/global/networks/test-network-2"}}},
					}}},
				}}}},
				"enableLogging": {Kind: &structpb.Value_BoolValue{BoolValue: true}},
			},
		},
	}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	provider.enrichNetworkAssets(context.Background(), assets)

	enrichedAssets := lo.Filter(assets, func(asset *ExtendedGcpAsset, _ int) bool {
		return asset.GetResource().GetData().GetFields()["enabledDnsLogging"] != nil
	})
	assetNames := lo.Map(enrichedAssets, func(asset *ExtendedGcpAsset, _ int) string { return asset.Name })

	s.True(lo.Contains(assetNames, "//compute.googleapis.com/projects/test-project/global/networks/test-network-1"))
	s.True(lo.Contains(assetNames, "//compute.googleapis.com/projects/test-project/global/networks/test-network-2"))

	s.Len(assets, 3)         // 3 network assets in total
	s.Len(enrichedAssets, 2) // 2 assets was enriched
}

func (s *ProviderTestSuite) TestListServiceUsageAssets() {
	expected := []*ServiceUsageAsset{
		{
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "1",
				ProjectName:      "ProjectName1",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			Services: []*ExtendedGcpAsset{{
				Asset: &assetpb.Asset{Name: "ServiceUsage1", Resource: &assetpb.Resource{}, IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"},
				Ecs:   &fetching.EcsGcp{ProjectId: "1", ProjectName: "ProjectName1", OrganizationId: "1", OrganizationName: "OrganizationName1"},
			}},
		},
		{
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "2",
				ProjectName:      "ProjectName2",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			Services: []*ExtendedGcpAsset{{
				Asset: &assetpb.Asset{Name: "ServiceUsage2", Resource: nil, IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"},
				Ecs:   &fetching.EcsGcp{ProjectId: "2", ProjectName: "ProjectName2", OrganizationId: "1", OrganizationName: "OrganizationName1"},
			}},
		},
	}

	provider := &Provider{
		log: logp.NewLogger("test"),
		inventory: &AssetsInventoryWrapper{
			Close: func() error { return nil },
			ListAssets: func(_ context.Context, _ *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
				return s.mockedIterator
			},
		},
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, parent string) string {
				if parent == "projects/1" {
					return "ProjectName1"
				}
				return "ProjectName2"
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName1"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	// asset's resource
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage2", Resource: nil, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	// asset's iam policy
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage1", IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage2", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "serviceusage.googleapis.com/Service"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	values, err := provider.ListServiceUsageAssets(context.Background())
	s.Require().NoError(err)

	// 2 assets, 1 for each project
	s.Len(values, 2)
	s.ElementsMatch(expected, values)
}

func (s *ProviderTestSuite) TestListLoggingAssets() {
	expected := []*LoggingAsset{
		{
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "1",
				ProjectName:      "ProjectName1",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			LogSinks: []*ExtendedGcpAsset{
				{
					Asset: &assetpb.Asset{Name: "LogSink1", Resource: &assetpb.Resource{}, IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"},
					Ecs:   &fetching.EcsGcp{ProjectId: "1", ProjectName: "ProjectName1", OrganizationId: "1", OrganizationName: "OrganizationName1"},
				},
				{
					Asset: &assetpb.Asset{Name: "LogSink3", Resource: nil, IamPolicy: nil, Ancestors: []string{"organizations/1"}, AssetType: "logging.googleapis.com/LogSink"},
					Ecs:   &fetching.EcsGcp{ProjectId: "", ProjectName: "", OrganizationId: "1", OrganizationName: "OrganizationName1"},
				},
			},
		},
		{
			Ecs: &fetching.EcsGcp{
				Provider:         "gcp",
				ProjectId:        "2",
				ProjectName:      "ProjectName2",
				OrganizationId:   "1",
				OrganizationName: "OrganizationName1",
			},
			LogSinks: []*ExtendedGcpAsset{
				{
					Asset: &assetpb.Asset{Name: "LogSink2", Resource: nil, IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"},
					Ecs:   &fetching.EcsGcp{ProjectId: "2", ProjectName: "ProjectName2", OrganizationId: "1", OrganizationName: "OrganizationName1"},
				},
				{
					Asset: &assetpb.Asset{Name: "LogSink3", Resource: nil, IamPolicy: nil, Ancestors: []string{"organizations/1"}, AssetType: "logging.googleapis.com/LogSink"},
					Ecs:   &fetching.EcsGcp{ProjectId: "", ProjectName: "", OrganizationId: "1", OrganizationName: "OrganizationName1"},
				},
			},
		},
	}

	provider := &Provider{
		log: logp.NewLogger("test"),
		inventory: &AssetsInventoryWrapper{
			Close: func() error { return nil },
			ListAssets: func(_ context.Context, _ *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
				return s.mockedIterator
			},
		},
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, parent string) string {
				if parent == "projects/1" {
					return "ProjectName1"
				}

				if parent == "projects/2" {
					return "ProjectName2"
				}

				return ""
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName1"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	// asset's resource
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink2", Resource: nil, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink3", Resource: nil, Ancestors: []string{"organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once() // asset at the organization level
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	// asset's iam policy
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink1", IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink2", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "LogSink3", IamPolicy: nil, Ancestors: []string{"organizations/1"}, AssetType: "logging.googleapis.com/LogSink"}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	values, err := provider.ListLoggingAssets(context.Background())
	s.Require().NoError(err)

	// 2 assets, 1 for each project
	s.Len(values, 2)
	s.ElementsMatch(expected, values)
}

func (s *ProviderTestSuite) TestListProjectsAncestorsPolicies() {
	provider := &Provider{
		log:       s.logger,
		inventory: s.mockedInventory,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(_ context.Context, _ string) string {
				return "ProjectName"
			},
			getOrganizationDisplayName: func(_ context.Context, _ string) string {
				return "OrganizationName"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName2", IamPolicy: &iampb.Policy{}, Ancestors: []string{"organizations/1"}}, nil).Once()
	s.mockedIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	value, err := provider.ListProjectsAncestorsPolicies(context.Background())
	s.Require().NoError(err)

	s.Len(value, 1)             // single project
	s.Len(value[0].Policies, 2) // multiple policies - project + org
	s.Equal("ProjectName", value[0].Ecs.ProjectName)
	s.Equal("OrganizationName", value[0].Ecs.OrganizationName)
	s.Equal("AssetName1", value[0].Policies[0].Name)
	s.Equal("AssetName2", value[0].Policies[1].Name)
}
