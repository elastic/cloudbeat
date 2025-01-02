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

package azurefetcher

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/ecs"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	azurelib_inventory "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestAccountFetcher_Fetch_Tenants(t *testing.T) {
	azureAssets := []azurelib_inventory.AzureAsset{
		{
			Id:          "/tenants/<tenant UUID>",
			Name:        "<tenant UUID>",
			DisplayName: "Mario",
			TenantId:    "<tenant UUID>",
		},
	}
	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureTenant,
			"/tenants/<tenant UUID>",
			"Mario",
			inventory.WithRawAsset(azureAssets[0]),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "<tenant UUID>",
				ServiceName: "Azure",
			}),
		),
	}

	// setup
	logger := logp.NewLogger("azurefetcher_test")
	provider := newMockAccountProvider(t)
	provider.EXPECT().ListTenants(mock.Anything).Return(azureAssets, nil)
	provider.EXPECT().ListSubscriptions(mock.Anything).Return(nil, nil)
	fetcher := newAccountFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestAccountFetcher_Fetch_Subscriptions(t *testing.T) {
	azureAssets := []azurelib_inventory.AzureAsset{
		{
			Id:          "/subscriptions/<sub UUID>",
			Name:        "<sub UUID>",
			DisplayName: "Luigi",
			TenantId:    "<sub UUID>",
		},
	}
	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureSubscription,
			"/subscriptions/<sub UUID>",
			"Luigi",
			inventory.WithRawAsset(azureAssets[0]),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.AzureCloudProvider,
				AccountID:   "<sub UUID>",
				ServiceName: "Azure",
			}),
		),
	}

	// setup
	logger := logp.NewLogger("azurefetcher_test")
	provider := newMockAccountProvider(t)
	provider.EXPECT().ListTenants(mock.Anything).Return(nil, nil)
	provider.EXPECT().ListSubscriptions(mock.Anything).Return(azureAssets, nil)
	fetcher := newAccountFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
