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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/google/uuid"
	"github.com/microsoft/kiota-abstractions-go/store"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestActiveDirectoryFetcher_Fetch(t *testing.T) {
	servicePrincipal := &models.ServicePrincipal{
		DirectoryObject: models.DirectoryObject{
			Entity: models.Entity{},
		},
	}
	appOwnerOrganizationId, _ := uuid.NewUUID()
	values := map[string]any{
		"id":                     pointers.Ref("id"),
		"displayName":            pointers.Ref("dn"),
		"appOwnerOrganizationId": &appOwnerOrganizationId,
	}
	store := store.NewInMemoryBackingStore()
	for k, v := range values {
		_ = store.Set(k, v)
	}
	servicePrincipal.SetBackingStore(store)

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureServicePrincipal,
			[]string{"id"},
			"dn",
			inventory.WithRawAsset(values),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: appOwnerOrganizationId.String(),
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		),
	}

	// setup
	logger := logp.NewLogger("azurefetcher_test")
	provider := newMockActivedirectoryProvider(t)

	provider.EXPECT().ListServicePrincipals(mock.Anything).Return(
		[]*models.ServicePrincipal{servicePrincipal}, nil,
	)

	fetcher := newActiveDirectoryFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
