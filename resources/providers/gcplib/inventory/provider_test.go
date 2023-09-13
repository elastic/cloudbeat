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
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

type ProviderTestSuite struct {
	suite.Suite
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) TestProviderInit() {
	initMock := new(MockProviderInitializerAPI)

	ctx := context.Background()
	log := logp.NewLogger("test")
	gcpConfig := auth.GcpFactoryConfig{
		Parent:     "projects/1",
		ClientOpts: []option.ClientOption{},
	}

	initMock.On("Init", ctx, log, gcpConfig).Return(&Provider{}, nil).Once()
	provider, err := initMock.Init(ctx, log, gcpConfig)
	s.Assert().NoError(err)
	s.Assert().NotNil(provider)
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	ctx := context.Background()
	mockIterator := new(MockIterator)
	provider := &Provider{
		log: logp.NewLogger("test"),
		inventory: &AssetsInventoryWrapper{
			Close: func() error { return nil },
			ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
				return mockIterator
			},
		},
		ctx: ctx,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(ctx context.Context, parent string) string {
				return "ProjectName"
			},
			getOrganizationDisplayName: func(ctx context.Context, parent string) string {
				return "OrganizationName"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName2", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	value, err := provider.ListAllAssetTypesByName([]string{"test"})
	s.Assert().NoError(err)

	// test merging assets with same name:
	assetNames := lo.Map(value, func(asset *ExtendedGcpAsset, _ int) string { return asset.Name })
	resourceAssets := lo.Filter(value, func(asset *ExtendedGcpAsset, _ int) bool { return asset.Resource != nil })
	policyAssets := lo.Filter(value, func(asset *ExtendedGcpAsset, _ int) bool { return asset.IamPolicy != nil })
	s.Assert().Equal(lo.Contains(assetNames, "AssetName1"), true)
	s.Assert().Equal(len(resourceAssets), 2) // 2 assets with resources (assetName1, assetName2)
	s.Assert().Equal(len(policyAssets), 1)   // 1 assets with policy 	(assetName1)
	s.Assert().Equal(len(value), 2)          // 2 assets in total 		(assetName1 merged resource/policy, assetName2)

	// tests extending assets with display names for org/prj:
	projectNames := lo.UniqBy(value, func(asset *ExtendedGcpAsset) string { return asset.Ecs.ProjectName })
	orgNames := lo.UniqBy(value, func(asset *ExtendedGcpAsset) string { return asset.Ecs.OrganizationName })
	s.Assert().Equal(len(projectNames), 1)
	s.Assert().Equal(projectNames[0].Ecs.ProjectName, "ProjectName")
	s.Assert().Equal(len(orgNames), 1)
	s.Assert().Equal(orgNames[0].Ecs.OrganizationName, "OrganizationName")
}

func (s *ProviderTestSuite) TestListMonitoringAssets() {
	ctx := context.Background()
	mockIterator := new(MockIterator)
	provider := &Provider{
		log: logp.NewLogger("test"),
		inventory: &AssetsInventoryWrapper{
			Close: func() error { return nil },
			ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
				return mockIterator
			},
		},
		ctx: ctx,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(ctx context.Context, parent string) string {
				if parent == "projects/1" {
					return "ProjectName1"
				}
				return "ProjectName2"
			},
			getOrganizationDisplayName: func(ctx context.Context, parent string) string {
				return "OrganizationName1"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	//  AssetType: "logging.googleapis.com/LogMetric"}
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetric1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogMetric"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetric1", IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogMetric"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	//  AssetType: "monitoring.googleapis.com/AlertPolicy"}
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicy1", Resource: nil, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "monitoring.googleapis.com/AlertPolicy"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicy1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "monitoring.googleapis.com/AlertPolicy"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	var monitoringAssetTypes = map[string][]string{
		"LogMetric":   {"logging.googleapis.com/LogMetric"},
		"AlertPolicy": {"monitoring.googleapis.com/AlertPolicy"},
	}
	value, err := provider.ListMonitoringAssets(monitoringAssetTypes)
	s.Assert().NoError(err)

	// 2 assets, 1 for each project
	s.Assert().Equal(len(value), 2)
	// project1 has 1 logMetric
	s.Assert().Equal(len(value[0].Alerts), 0)
	s.Assert().Equal(len(value[0].LogMetrics), 1)
	s.Assert().Equal(value[0].LogMetrics[0].Name, "LogMetric1")
	s.Assert().Equal(value[0].Ecs.ProjectId, "1")
	s.Assert().Equal(value[0].Ecs.ProjectName, "ProjectName1")
	s.Assert().Equal(value[0].Ecs.OrganizationId, "1")
	s.Assert().Equal(value[0].Ecs.OrganizationName, "OrganizationName1")

	// project2 has 1 alertPolicy
	s.Assert().Equal(len(value[1].LogMetrics), 0)
	s.Assert().Equal(len(value[1].Alerts), 1)
	s.Assert().Equal(value[1].Alerts[0].Name, "AlertPolicy1")
	s.Assert().Equal(value[1].Ecs.ProjectId, "2")
	s.Assert().Equal(value[1].Ecs.ProjectName, "ProjectName2")
	s.Assert().Equal(value[1].Ecs.OrganizationId, "1")
	s.Assert().Equal(value[1].Ecs.OrganizationName, "OrganizationName1")
}

func (s *ProviderTestSuite) TestListServiceUsageAssets() {
	ctx := context.Background()
	mockIterator := new(MockIterator)
	provider := &Provider{
		log: logp.NewLogger("test"),
		inventory: &AssetsInventoryWrapper{
			Close: func() error { return nil },
			ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
				return mockIterator
			},
		},
		ctx: ctx,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: &ResourceManagerWrapper{
			getProjectDisplayName: func(ctx context.Context, parent string) string {
				if parent == "projects/1" {
					return "ProjectName1"
				}
				return "ProjectName2"
			},
			getOrganizationDisplayName: func(ctx context.Context, parent string) string {
				return "OrganizationName1"
			},
		},
		crmCache: make(map[string]*fetching.EcsGcp),
	}

	//  AssetType: "logging.googleapis.com/LogMetric"}
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogMetric"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage1", IamPolicy: nil, Ancestors: []string{"projects/1", "organizations/1"}, AssetType: "logging.googleapis.com/LogMetric"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	//  AssetType: "monitoring.googleapis.com/AlertPolicy"}
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage2", Resource: nil, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "monitoring.googleapis.com/AlertPolicy"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "ServiceUsage2", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/2", "organizations/1"}, AssetType: "monitoring.googleapis.com/AlertPolicy"}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	var monitoringAssetTypes = map[string][]string{
		"LogMetric":   {"logging.googleapis.com/LogMetric"},
		"AlertPolicy": {"monitoring.googleapis.com/AlertPolicy"},
	}
	value, err := provider.ListMonitoringAssets(monitoringAssetTypes)
	s.Assert().NoError(err)

	// 2 assets, 1 for each project
	s.Assert().Equal(len(value), 2)
	// project1 has 1 logMetric
	s.Assert().Equal(len(value[0].Alerts), 0)
	s.Assert().Equal(len(value[0].LogMetrics), 1)
	s.Assert().Equal(value[0].LogMetrics[0].Name, "ServiceUsage1")
	s.Assert().Equal(value[0].Ecs.ProjectId, "1")
	s.Assert().Equal(value[0].Ecs.ProjectName, "ProjectName1")
	s.Assert().Equal(value[0].Ecs.OrganizationId, "1")
	s.Assert().Equal(value[0].Ecs.OrganizationName, "OrganizationName1")

	// project2 has 1 alertPolicy
	s.Assert().Equal(len(value[1].LogMetrics), 0)
	s.Assert().Equal(len(value[1].Alerts), 1)
	s.Assert().Equal(value[1].Alerts[0].Name, "ServiceUsage2")
	s.Assert().Equal(value[1].Ecs.ProjectId, "2")
	s.Assert().Equal(value[1].Ecs.ProjectName, "ProjectName2")
	s.Assert().Equal(value[1].Ecs.OrganizationId, "1")
	s.Assert().Equal(value[1].Ecs.OrganizationName, "OrganizationName1")
}
