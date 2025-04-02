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
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var (
	policy   = assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}, Ancestors: []string{"projects/1", "organizations/1"}}
	resource = assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}, Ancestors: []string{"projects/1", "organizations/1"}}
)

type ProviderTestSuite struct {
	suite.Suite
	logger                 *clog.Logger
	mockedInventory        *AssetsInventoryWrapper
	mockedResourceIterator *MockIterator
	mockedPoliciesIterator *MockIterator
	mockedCrm              *ResourceManagerWrapper
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {
	s.logger = testhelper.NewObserverLogger(s.T())
	s.mockedResourceIterator = new(MockIterator)
	s.mockedPoliciesIterator = new(MockIterator)
	s.mockedInventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.ContentType {
			case assetpb.ContentType_RESOURCE:
				return s.mockedResourceIterator
			case assetpb.ContentType_IAM_POLICY:
				return s.mockedPoliciesIterator
			default:
				return nil
			}
		},
	}
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
	return &Provider{
		log:       s.logger,
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

	initMock.EXPECT().Init(mock.Anything, s.logger, gcpConfig).Return(&Provider{}, nil).Once()
	provider, err := initMock.Init(context.Background(), s.logger, gcpConfig)
	s.Require().NoError(err)
	s.NotNil(provider)
}

func (s *ProviderTestSuite) TestListAssetTypes_IteratorsError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()

	s.mockedResourceIterator.On("Next").Return(nil, errors.New("test")).Once()
	s.mockedPoliciesIterator.On("Next").Return(nil, errors.New("test")).Once()

	go provider.ListAssetTypes(context.Background(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Empty(results)
	s.mockedResourceIterator.AssertExpectations(s.T())
	s.mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_PolicyIteratorError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()

	s.mockedResourceIterator.On("Next").Return(&resource, nil).Once()
	s.mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedPoliciesIterator.On("Next").Return(nil, errors.New("test")).Once()

	go provider.ListAssetTypes(context.Background(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	logs := logp.ObserverLogs().FilterMessageSnippet("Error fetching GCP IAM_POLICY: test").TakeAll()
	s.Len(logs, 1)
	s.Equal(zapcore.ErrorLevel, logs[0].Level)

	// verify we send the asset we have (resource)
	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	s.Nil(results[0].IamPolicy)
	s.NotNil(results[0].Resource)
	s.mockedResourceIterator.AssertExpectations(s.T())
	s.mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_ResourceIteratorError() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()

	s.mockedResourceIterator.On("Next").Return(nil, errors.New("test")).Once()
	s.mockedPoliciesIterator.On("Next").Return(&policy, nil).Once()
	s.mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	go provider.ListAssetTypes(context.Background(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	logs := logp.ObserverLogs().FilterMessageSnippet("Error fetching GCP RESOURCE: test").TakeAll()
	s.Len(logs, 1)
	s.Equal(zapcore.ErrorLevel, logs[0].Level)

	// verify we send the asset we have (policy)
	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	s.NotNil(results[0].IamPolicy)
	s.Nil(results[0].Resource)
	s.mockedResourceIterator.AssertExpectations(s.T())
	s.mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListAssetTypes_Success() {
	outCh := make(chan *ExtendedGcpAsset)
	provider := s.NewMockProvider()
	provider.crm.config.Parent = "projects/1"
	s.mockedResourceIterator.On("Next").Return(&resource, nil).Once()
	s.mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	s.mockedPoliciesIterator.On("Next").Return(&policy, nil).Once()
	s.mockedPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	go provider.ListAssetTypes(context.Background(), []string{"someAssetType"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 1)
	s.Equal("AssetName1", results[0].Name)
	// verify merged assets
	s.NotNil(results[0].IamPolicy)
	s.NotNil(results[0].Resource)
	// verify cloud account metadata
	s.Equal("ProjectName", results[0].CloudAccount.AccountName)
	s.Empty(results[0].CloudAccount.OrganizationName) // when config.parent is project, orgName is empty
	s.mockedResourceIterator.AssertExpectations(s.T())
	s.mockedPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListMonitoringAssets_Success() {
	provider := s.NewMockProvider()
	logMetricsIterator := new(MockIterator)
	alertPoliciesIterator := new(MockIterator)

	provider.inventory = &AssetsInventoryWrapper{
		Close: func() error { return nil },
		ListAssets: func(_ context.Context, req *assetpb.ListAssetsRequest, _ ...gax.CallOption) Iterator {
			switch req.AssetTypes[0] {
			case MonitoringLogMetricAssetType:
				return logMetricsIterator
			case MonitoringAlertPolicyAssetType:
				return alertPoliciesIterator
			default:
				return nil
			}
		},
	}

	logMetricsIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetricName1", Ancestors: []string{"projects/1"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	logMetricsIterator.On("Next").Return(&assetpb.Asset{Name: "LogMetricName1", Ancestors: []string{"projects/2"}, AssetType: MonitoringLogMetricAssetType}, nil).Once()
	logMetricsIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicyName1", Ancestors: []string{"projects/1"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{Name: "AlertPolicyName1", Ancestors: []string{"projects/2"}, AssetType: MonitoringAlertPolicyAssetType}, nil).Once()
	alertPoliciesIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	outCh := make(chan *MonitoringAsset)
	go provider.ListMonitoringAssets(context.Background(), outCh)

	results := testhelper.CollectResourcesBlocking(outCh)

	// grouped by project id + enriched with cloud account metadata
	s.Len(results, 2)
	s.Len(results[0].LogMetrics, 1)
	s.Len(results[0].Alerts, 1)
	s.Len(results[1].LogMetrics, 1)
	s.Len(results[1].Alerts, 1)
	ids1 := []string{results[0].LogMetrics[0].CloudAccount.AccountId, results[0].Alerts[0].CloudAccount.AccountId}
	ids2 := []string{results[1].LogMetrics[0].CloudAccount.AccountId, results[1].Alerts[0].CloudAccount.AccountId}
	s.ElementsMatch(append(ids1, ids2...), []string{"1", "1", "2", "2"})

	defer logMetricsIterator.AssertExpectations(s.T())
	defer alertPoliciesIterator.AssertExpectations(s.T())
}

func (s *ProviderTestSuite) TestListProjectAssets() {
	outCh := make(chan *ProjectAssets)
	provider := s.NewMockProvider()

	s.mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "asset1", Ancestors: []string{"projects/1"}}, nil).Once()
	s.mockedResourceIterator.On("Next").Return(&assetpb.Asset{Name: "asset2", Ancestors: []string{"projects/2"}}, nil).Once()
	s.mockedResourceIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	go provider.ListProjectAssets(context.Background(), []string{"assettype"}, outCh)
	results := testhelper.CollectResourcesBlocking(outCh)

	s.Len(results, 2)
	s.Equal("1", results[0].CloudAccount.AccountId)
	s.Len(results[0].Assets, 1)
	s.Equal("2", results[1].CloudAccount.AccountId)
	s.Len(results[1].Assets, 1)

	s.mockedResourceIterator.AssertExpectations(s.T())
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

	go provider.ListNetworkAssets(context.Background(), outCh)
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

	go provider.ListProjectsAncestorsPolicies(context.Background(), outCh)
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
