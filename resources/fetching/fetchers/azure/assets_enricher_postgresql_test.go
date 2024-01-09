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

package fetchers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestPostgresqlEnricher_Enrich(t *testing.T) {
	tcs := map[string]struct {
		input       []inventory.AzureAsset
		expected    []inventory.AzureAsset
		expectError bool
		configRes   map[string]enricherResponse
	}{
		"Error on enriching configs": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockPostgresAsset("id1", "psql-a"),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockPostgresAsset("id1", "psql-a"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []map[string]any{
						psqlConfigProps("log_checkpoints", "on"),
						psqlConfigProps("log_connections", "off"),
					},
				}),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			configRes: map[string]enricherResponse{
				"psql-a": assetRes(
					mockPostgresConfigAsset("conf1", psqlConfigProps("log_checkpoints", "on")),
					mockPostgresConfigAsset("conf2", psqlConfigProps("log_connections", "off")),
				),
				"psql-b": errorRes(errors.New("error")),
			},
		},
		"Enrich configs": {
			expectError: false,
			input: []inventory.AzureAsset{
				mockPostgresAsset("id1", "psql-a"),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockPostgresAsset("id1", "psql-a"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []map[string]any{
						psqlConfigProps("log_checkpoints", "on"),
						psqlConfigProps("log_connections", "off"),
					},
				}),
				mockOther("id2"),
				addExtension(mockPostgresAsset("id3", "psql-b"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []map[string]any{
						psqlConfigProps("log_disconnections", "on"),
						psqlConfigProps("connection_throttling", "off"),
					},
				}),
			},
			configRes: map[string]enricherResponse{
				"psql-a": assetRes(
					mockPostgresConfigAsset("conf1", psqlConfigProps("log_checkpoints", "on")),
					mockPostgresConfigAsset("conf2", psqlConfigProps("log_connections", "off")),
				),
				"psql-b": assetRes(
					mockPostgresConfigAsset("conf1", psqlConfigProps("log_disconnections", "on")),
					mockPostgresConfigAsset("conf2", psqlConfigProps("connection_throttling", "off")),
				),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			for serverName, r := range tc.configRes {
				provider.EXPECT().ListPostgresConfigurations(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			e := postgresqlEnricher{provider: provider}

			err := e.Enrich(context.Background(), cmd, tc.input)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func mockPostgresAsset(id, name string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           name,
		Type:           inventory.PostgreSQLDBAssetType,
	}
}

func mockPostgresConfigAsset(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Type:           inventory.PostgreSQLDBAssetType + "/configuration",
		Properties:     props,
	}
}

func psqlConfigProps(name string, value string) map[string]any {
	return map[string]any{
		"name":         name,
		"source":       "system-default",
		"value":        value,
		"dataType":     "Boolean",
		"defaultValue": "on",
	}
}
