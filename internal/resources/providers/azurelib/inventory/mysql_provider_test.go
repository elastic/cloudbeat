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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/stretchr/testify/require"
)

type flexibleConfigFn func(configName string) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error)

func mockAssetFlexibleConfigurationMysqlProvider(f flexibleConfigFn) MysqlProviderAPI {
	wrapper := mysqlAzureClientWrapper{
		AssetFlexibleConfiguration: func(_ context.Context, _, _, _, configName string, _ *arm.ClientOptions, _ *armmysqlflexibleservers.ConfigurationsClientGetOptions) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error) {
			return f(configName)
		},
	}

	return &mysqlProvider{client: wrapper, log: clog.NewLogger("mock_asset_flexible_config")}
}

func TestMysqlProvider_GetFlexibleTLSVersionConfiguration(t *testing.T) {
	cases := map[string]struct {
		expectError bool
		configMock  flexibleConfigFn
		expected    []AzureAsset
	}{
		"Returns error": {
			expectError: true,
			configMock: func(_ string) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error) {
				return armmysqlflexibleservers.ConfigurationsClientGetResponse{}, errors.New("error")
			},
			expected: nil,
		},

		"No configuration found": {
			expectError: false,
			configMock: func(_ string) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error) {
				return armmysqlflexibleservers.ConfigurationsClientGetResponse{}, nil
			},
			expected: nil,
		},

		"Returns TLS version configurtion": {
			expectError: false,
			configMock: func(configName string) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error) {
				require.Equal(t, mysqlConfigurationTLSVersion, configName)
				return armmysqlflexibleservers.ConfigurationsClientGetResponse{
					Configuration: armmysqlflexibleservers.Configuration{
						ID:   to.Ptr("config1"),
						Name: to.Ptr(mysqlConfigurationTLSVersion),
						Properties: &armmysqlflexibleservers.ConfigurationProperties{
							Source:       to.Ptr(armmysqlflexibleservers.ConfigurationSourceSystemDefault),
							Value:        to.Ptr("TLSV1.2"),
							DataType:     to.Ptr("string"),
							DefaultValue: to.Ptr(""),
						},
					},
				}, nil
			},

			expected: []AzureAsset{
				{
					Id:             "config1",
					Name:           "tls_version",
					Location:       assetLocationGlobal,
					SubscriptionId: "subscription",
					ResourceGroup:  "resource",
					Properties: map[string]any{
						"source":       "system-default",
						"value":        "tlsv1.2",
						"dataType":     "string",
						"defaultValue": "",
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			provider := mockAssetFlexibleConfigurationMysqlProvider(tc.configMock)

			assets, err := provider.GetFlexibleTLSVersionConfiguration(context.Background(), "subscription", "resource", "server")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, assets)
		})
	}
}
