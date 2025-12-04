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
	"errors"
	"fmt"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
	mockedInventory *AssetsInventoryWrapper
	mockedCrm       *ResourceManagerWrapper
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}
func NewMockInventoryContentIterators() (inventory *AssetsInventoryWrapper, resourceIter *MockIterator, policiesIter *MockIterator) {
	mockedResourceIterator := new(MockIterator)
	mockedPoliciesIterator := new(MockIterator)

	mockedInventory := &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.ContentType {
			case assetpb.ContentType_RESOURCE:
				return mockedResourceIterator
			case assetpb.ContentType_IAM_POLICY:
				return mockedPoliciesIterator
			default:
				return nil
			}
		},
	}
	return mockedInventory, mockedResourceIterator, mockedPoliciesIterator
}

func (s *ProviderTestSuite) SetupTest() {
	s.mockedCrm = &ResourceManagerWrapper{
		getProjectDisplayName: func(_ context.Context, _ string) string {
			return "ProjectName"
		},
		getOrganizationDisplayName: func(_ context.Context, _ string) string {
			return "OrganizationName"
		},
	}
}

func (s *ProviderTestSuite) NewMockProvider() *Provider {
	log, observer := testhelper.NewObserverLogger(s.T())
	return &Provider{
		log:       log,
		observer:  observer,
		inventory: s.mockedInventory,
		config: auth.GcpFactoryConfig{
			Parent:     "projects/1",
			ClientOpts: []option.ClientOption{},
		},
		crm: s.mockedCrm,
	}
}

func (s *ProviderTestSuite) TestProviderInit() {
	initMock := new(MockProviderInitializerAPI)
	gcpConfig := auth.GcpFactoryConfig{
		Parent:     "projects/1",
		ClientOpts: []option.ClientOption{},
	}
	gcpCnf := config.GcpConfig{}

	logger := testhelper.NewLogger(s.T())
	initMock.EXPECT().Init(mock.Anything, logger, gcpConfig, gcpCnf).Return(&Provider{}, nil).Once()
	t := s.T()
	provider, err := initMock.Init(t.Context(), logger, gcpConfig, gcpCnf)
	s.Require().NoError(err)
	s.NotNil(provider)
}

func (s *ProviderTestSuite) TestListAssetTypes_IteratorsError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()
	inventory, mockedResourceIterator, mockedPoliciesIterator := NewMockInventoryContentIterators()
	provider.inventory = inventory

	mockedResourceIterator.On("Next").Return(nil, errors.New("test")).Once()
	mockedPoliciesIterator.On("Next").Return(nil, errors.New("test")).Once()
	t := s.T()
	go provider.ListAssetTypes(t.Context(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Empty(results)
	mockedResourceIterator.AssertExpectations(s.T())
	mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_PolicyIteratorError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()

	inventory, mockedResourceIterator, mockedPoliciesIterator := NewMockInventoryContentIterators()
	provider.inventory = inventory

	mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockedPoliciesIterator.On("Next").Return(nil, errors.New("test")).Once()
	t := s.T()
	go provider.ListAssetTypes(t.Context(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	logs := provider.observer.FilterMessageSnippet(fmt.Sprintf("Error fetching GCP %v of types: %v for %v: %v\n", "IAM_POLICY", []string{"someAssetType"}, provider.config.Parent, "test")).TakeAll()
	s.Len(logs, 1)
	s.Equal(zapcore.ErrorLevel, logs[0].Level)

	// verify we send the asset we have (resource)
	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	s.Nil(results[0].IamPolicy)
	s.NotNil(results[0].Resource)
	mockedResourceIterator.AssertExpectations(s.T())
	mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_ResourceIteratorError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()
	inventory, mockedResourceIterator, mockedPoliciesIterator := NewMockInventoryContentIterators()
	provider.inventory = inventory

	mockedResourceIterator.On("Next").Return(nil, errors.New("test")).Once()
	mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	t := s.T()
	go provider.ListAssetTypes(t.Context(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	logs := provider.observer.FilterMessageSnippet(fmt.Sprintf("Error fetching GCP %v of types: %v for %v: %v\n", "RESOURCE", []string{"someAssetType"}, provider.config.Parent, "test")).TakeAll()
	s.Len(logs, 1)
	s.Equal(zapcore.ErrorLevel, logs[0].Level)

	// verify we send the asset we have (policy)
	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	s.NotNil(results[0].IamPolicy)
	s.Nil(results[0].Resource)
	mockedResourceIterator.AssertExpectations(s.T())
	mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_Success() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()
	provider.crm.config.Parent = "projects/1"
	inventory, mockedResourceIterator, mockedPoliciesIterator := NewMockInventoryContentIterators()
	provider.inventory = inventory

	mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	t := s.T()
	go provider.ListAssetTypes(t.Context(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	// verify merged assets
	s.NotNil(results[0].IamPolicy)
	s.NotNil(results[0].Resource)
	// verify cloud account metadata
	s.Equal("ProjectName", results[0].CloudAccount.AccountName)
	s.Empty(results[0].CloudAccount.OrganizationName) // when config.parent is project, orgName is empty
	mockedResourceIterator.AssertExpectations(s.T())
	mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListMonitoringAssets_Success() {
	provider := s.NewMockProvider()
	logMetricsIterator := new(MockIterator)
	alertPoliciesIterator := new(MockIterator)
	projectIterator := new(MockIterator)

	provider.inventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.AssetTypes[0] {
			case MonitoringLogMetricAssetType:
				return logMetricsIterator
			case MonitoringAlertPolicyAssetType:
				return alertPoliciesIterator
			case CrmProjectAssetType:
				return projectIterator
			default:
				return nil
			}
		},
	}
	projectIterator.On("Next").Return(&assetpb.Asset{Name: "projects/1", Ancestors: []string{"organizations/1/projects/1"}}, nil).Once()
	projectIterator.On("Next").Return(&assetpb.Asset{Name: "projects/2", Ancestors: []string{"organizations/1/projects/2"}}, nil).Once()
	projectIterator.On("Next").Return(nil, iterator.Done).Once()

	logMetricsIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetricName1", Ancestors: []string{"projects/1"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	logMetricsIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicyName1", Ancestors: []string{"projects/1"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	logMetricsIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetricName2", Ancestors: []string{"projects/2"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	logMetricsIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicyName2", Ancestors: []string{"projects/2"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	outCh := make(chan *MonitoringAsset)
	t := s.T()
	go provider.ListMonitoringAssets(t.Context(), outCh)

	results := testhelper.CollectResourcesBlocking(outCh)

	// grouped by project id + enriched with cloud account metadata
	s.Len(results, 2)
	s.Len(results[0].LogMetrics, 1)
	s.Len(results[0].Alerts, 1)
	s.Len(results[1].LogMetrics, 1)
	s.Len(results[1].Alerts, 1)

	s.ElementsMatch([]string{"1", "1"}, []string{results[0].LogMetrics[0].CloudAccount.AccountId, results[0].Alerts[0].CloudAccount.AccountId})
	s.ElementsMatch([]string{"2", "2"}, []string{results[1].LogMetrics[0].CloudAccount.AccountId, results[1].Alerts[0].CloudAccount.AccountId})

	projectIterator.AssertExpectations(s.T())
	logMetricsIterator.AssertExpectations(s.T())
	alertPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListProjectAssets() {
	outCh := make(chan *ProjectAssets)
	provider := s.NewMockProvider()
	mockedProjectIterator := new(MockIterator)
	mockedResourceIterator := new(MockIterator)
	provider.inventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.AssetTypes[0] {
			case CrmProjectAssetType:
				return mockedProjectIterator
			default:
				return mockedResourceIterator
			}
		},
	}

	mockedProjectIterator.On("Next").Return(&assetpb.Asset{Name: "prj1", Ancestors: []string{"projects/1"}}, nil).Once()
	mockedProjectIterator.On("Next").Return(&assetpb.Asset{Name: "prj2", Ancestors: []string{"projects/2"}}, nil).Once()
	mockedProjectIterator.On("Next").Return(nil, iterator.Done).Once()

	mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "rsrc1", Ancestors: []string{"projects/1"}}, nil).Once()
	mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "rsrc2", Ancestors: []string{"projects/2"}}, nil).Once()
	mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	t := s.T()
	go provider.ListProjectAssets(t.Context(), []string{"assettype"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 2)
	s.Len(results[0].Assets, 1)
	s.Len(results[1].Assets, 1)
	s.ElementsMatch([]string{"1", "2"}, []string{results[0].CloudAccount.AccountId, results[1].CloudAccount.AccountId})

	mockedProjectIterator.AssertExpectations(s.T())
	mockedResourceIterator.AssertExpectations(s.T())
}
func (s *ProviderTestSuite) TestListNetworkAssets() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()

	mockedDnsIterator := new(MockIterator)
	mockedNetworkIterator := new(MockIterator)
	provider.inventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.AssetTypes[0] {
			case DnsPolicyAssetType:
				return mockedDnsIterator
			case ComputeNetworkAssetType:
				return mockedNetworkIterator
			default:
				return nil
			}
		},
	}
	mockedDnsIterator.On("Next").Return(&assetpb.Asset{Name: "//compute.googleapis.com/projects/1/global/networks/network1", Ancestors: []string{"projects/1"}, AssetType: DnsPolicyAssetType, Resource: &assetpb.Resource{
		Data: NewStructMap(map[string]any{
			"enableLogging": true,
			"networks": []any{
				map[string]any{
					"networkUrl": "/projects/1/global/networks/network1",
				},
			},
		}),
	}}, nil).Once()
	mockedNetworkIterator.On("Next").Return(&assetpb.Asset{Name: "//compute.googleapis.com/projects/2/global/networks/network1", Ancestors: []string{"projects/1"}, AssetType: DnsPolicyAssetType, Resource: &assetpb.Resource{
		Data: NewStructMap(map[string]any{
			"enableLogging": true,
			"networks": []any{
				map[string]any{
					"networkUrl": "/projects/2/global/networks/network1",
				},
			},
		}),
	}}, nil).Once()
	mockedDnsIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockedNetworkIterator.On("Next").Return(&assetpb.Asset{Name: "//compute.googleapis.com/projects/1/global/networks/network1", Ancestors: []string{"projects/1"}, AssetType: ComputeNetworkAssetType, Resource: &assetpb.Resource{Data: NewStructMap(map[string]any{})}}, nil).Once()
	mockedNetworkIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	t := s.T()
	go provider.ListNetworkAssets(t.Context(), outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 2)
	enrichedValues := lo.Map(results, func(r *ExtendedGcpAsset, _ int) bool {
		return r.Resource.Data.Fields["enabledDnsLogging"].GetBoolValue()
	})
	s.ElementsMatch(enrichedValues, []bool{true, false})

	mockedNetworkIterator.AssertExpectations(s.T())
	mockedDnsIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListProjectsAncestorsPolicies() {
	outCh := make(chan *ProjectPoliciesAsset)
	provider := s.NewMockProvider()

	prjIterator := new(MockIterator)
	orgIterator := new(MockIterator)
	provider.crm.config.Parent = "organizations/1"
	provider.inventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.AssetTypes[0] {
			case CrmProjectAssetType:
				return prjIterator
			case CrmOrgAssetType:
				return orgIterator
			default:
				return nil
			}
		},
	}

	prjIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}, nil).Once()
	prjIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	orgIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"organizations/1"}}, nil).Once()
	orgIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	t := s.T()
	go provider.ListProjectsAncestorsPolicies(t.Context(), outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 1)
	s.Len(results[0].Policies, 2)
	s.Equal("ProjectName", results[0].CloudAccount.AccountName)
	s.Equal("OrganizationName", results[0].CloudAccount.OrganizationName)

	prjIterator.AssertExpectations(s.T())
	orgIterator.AssertExpectations(s.T())
}

func NewStructMap(data map[string]any) *structpb.Struct {
	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		panic(err)
	}
	return dataStruct
}
