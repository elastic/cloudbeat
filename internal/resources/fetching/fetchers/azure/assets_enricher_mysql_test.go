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

	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestMysqlAssetEnricher_Enrich(t *testing.T) {
	cases := map[string]struct {
		input       []inventory.AzureAsset
		expected    []inventory.AzureAsset
		expectError bool
		configRes   map[string]enricherResponse
	}{
		"error in one asset": {
			input: []inventory.AzureAsset{
				mockFlexibleMysqlAsset("mysql-1", "mysql-1"),
				mockPostgresAsset("psql-1", "psql"),
				mockFlexibleMysqlAsset("mysql-2", "mysql-2"),
			},
			expected: []inventory.AzureAsset{
				mockFlexibleMysqlAsset("mysql-1", "mysql-1"),
				mockPostgresAsset("psql-1", "psql"),
				addExtension(mockFlexibleMysqlAsset("mysql-2", "mysql-2"), map[string]any{
					inventory.ExtensionMysqlConfigurations: []inventory.AzureAsset{
						mockFlexMysqlTLSVersionConfig("mysql-2/tls_version", "tlsv1.2"),
					},
				}),
			},
			expectError: true,
			configRes: map[string]enricherResponse{
				"mysql-1": errorRes(errors.New("error")),
				"mysql-2": assetRes(mockFlexMysqlTLSVersionConfig("mysql-2/tls_version", "tlsv1.2")),
			},
		},

		"enrich tls version properly": {
			input: []inventory.AzureAsset{
				mockFlexibleMysqlAsset("mysql-1", "mysql-1"),
				mockFlexibleMysqlAsset("mysql-2", "mysql-2"),
				mockFlexibleMysqlAsset("mysql-3", "mysql-3"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockFlexibleMysqlAsset("mysql-1", "mysql-1"), map[string]any{
					inventory.ExtensionMysqlConfigurations: []inventory.AzureAsset{
						mockFlexMysqlTLSVersionConfig("mysql-1/tls_version", "tlsv1.2"),
					},
				}),
				addExtension(mockFlexibleMysqlAsset("mysql-2", "mysql-2"), map[string]any{
					inventory.ExtensionMysqlConfigurations: []inventory.AzureAsset{
						mockFlexMysqlTLSVersionConfig("mysql-2/tls_version", "tlsv1.3"),
					},
				}),
				mockFlexibleMysqlAsset("mysql-3", "mysql-3"),
			},
			expectError: false,
			configRes: map[string]enricherResponse{
				"mysql-1": assetRes(mockFlexMysqlTLSVersionConfig("mysql-1/tls_version", "tlsv1.2")),
				"mysql-2": assetRes(mockFlexMysqlTLSVersionConfig("mysql-2/tls_version", "tlsv1.3")),
				"mysql-3": assetRes(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}
			provider := azurelib.NewMockProviderAPI(t)
			ctx := context.Background()

			for serverName, mock := range tc.configRes {
				provider.EXPECT().GetFlexibleTLSVersionConfiguration(ctx, "subId", "group", serverName).Return(mock.assets, mock.err)
			}

			e := mysqlAssetEnricher{provider: provider}

			err := e.Enrich(ctx, cmd, tc.input)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, tc.input)
		})
	}
}

func mockFlexMysqlTLSVersionConfig(id, tlsVersion string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           "tls_version",
		Type:           inventory.FlexibleMySQLDBServerAssetType + "/configuration",
		Properties: map[string]any{
			"name":         "tls_version",
			"source":       "system-default",
			"value":        tlsVersion,
			"dataType":     "string",
			"defaultValue": "",
		},
	}
}
func mockFlexibleMysqlAsset(id string, name string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           name,
		Type:           inventory.FlexibleMySQLDBServerAssetType,
	}
}
