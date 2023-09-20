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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite
	ctx          context.Context
	logger       *logp.Logger
	mockedClient *AzureClientWrapper
}

var nonTruncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(1)),
	Data: []any{
		map[string]any{
			"id":             "3",
			"name":           "3",
			"location":       "3",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "3",
			"subscriptionId": "3",
			"tenantId":       "3",
			"type":           "3",
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
}

var truncatedResponse = armresourcegraph.QueryResponse{
	Count: to.Ptr(int64(2)),
	Data: []any{
		map[string]any{
			"id":             "1",
			"name":           "1",
			"location":       "1",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "1",
			"subscriptionId": "1",
			"tenantId":       "1",
			"type":           "1",
		},
		map[string]any{
			"id":             "2",
			"name":           "2",
			"location":       "2",
			"properties":     map[string]any{"test": "test"},
			"resourceGroup":  "2",
			"subscriptionId": "2",
			"tenantId":       "2",
			"type":           "2",
		},
	},
	ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedTrue),
	SkipToken:       to.Ptr("token"),
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = logp.NewLogger("test")
	s.mockedClient = &AzureClientWrapper{
		AssetQuery: func(ctx context.Context, query armresourcegraph.QueryRequest, options *armresourcegraph.ClientResourcesOptions) (armresourcegraph.ClientResourcesResponse, error) {
			if query.Options.SkipToken != nil && *query.Options.SkipToken != "" {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: nonTruncatedResponse,
				}, nil
			} else {
				return armresourcegraph.ClientResourcesResponse{
					QueryResponse: truncatedResponse,
				}, nil
			}
		},
	}
}

func (s *ProviderTestSuite) TestGetString() {
	tests := []struct {
		name string
		data map[string]any
		key  string
		want string
	}{
		{
			name: "nil map",
			data: nil,
			key:  "key",
			want: "",
		},
		{
			name: "key does not exist",
			data: map[string]any{"key": "value"},
			key:  "other-key",
			want: "",
		},
		{
			name: "wrong type",
			data: map[string]any{"key": 1},
			key:  "key",
			want: "",
		},
		{
			name: "correct value",
			data: map[string]any{"key": "value", "other-key": 1},
			key:  "key",
			want: "value",
		},
	}
	for _, tt := range tests {
		s.Assert().Equal(tt.want, getString(tt.data, tt.key), "getString(%v, %s) = %s", tt.data, tt.key, tt.want)
	}
}

func (s *ProviderTestSuite) TestProviderInit() {
	initMock := new(MockProviderInitializerAPI)
	azureConfig := auth.AzureFactoryConfig{
		Credentials: &azidentity.DefaultAzureCredential{},
	}

	initMock.On("Init", s.ctx, s.logger, azureConfig).Return(&Provider{}, nil).Once()
	provider, err := initMock.Init(s.ctx, s.logger, azureConfig)
	s.Assert().NoError(err)
	s.Assert().NotNil(provider)
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	provider := &Provider{
		log:    s.logger,
		client: s.mockedClient,
		ctx:    s.ctx,
		Config: auth.AzureFactoryConfig{
			Credentials: &azidentity.DefaultAzureCredential{},
		},
	}

	values, err := provider.ListAllAssetTypesByName([]string{"test"})
	s.Assert().NoError(err)
	s.Assert().Equal(int(*nonTruncatedResponse.Count+*truncatedResponse.Count), len(values))
	lo.ForEach(values, func(r AzureAsset, index int) {
		strIndex := fmt.Sprintf("%d", index+1)
		s.Assert().Equal(r.Id, strIndex)
		s.Assert().Equal(r.Name, strIndex)
		s.Assert().Equal(r.Location, strIndex)
		s.Assert().Equal(r.ResourceGroup, strIndex)
		s.Assert().Equal(r.SubscriptionId, strIndex)
		s.Assert().Equal(r.TenantId, strIndex)
		s.Assert().Equal(r.Type, strIndex)
		s.Assert().Equal(r.Properties, map[string]any{"test": "test"})
	})
}
