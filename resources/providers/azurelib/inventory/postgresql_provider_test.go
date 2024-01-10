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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

type assetConfigFn func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error)

func mockAssetPSQLConfiguration(f assetConfigFn) PostgresqlProviderAPI {
	cl := &psqlAzureClientWrapper{
		AssetConfigurations: func(ctx context.Context, subID, resourceGroup, serverName string, clientOptions *arm.ClientOptions, options *armpostgresql.ConfigurationsClientListByServerOptions) ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
			return f()
		},
	}

	return &psqlProvider{
		log:    logp.NewLogger("mock_asset_sql_encryption_protector"),
		client: cl,
	}
}

func TestListPostgresConfigurations(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    assetConfigFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"No Encryption Protector Response": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return wrapPsqlConfigResponse(
					wrapPsqlConfigResult(
						psqlConfigAzure("id1", "log_checkpoints", "on"),
						psqlConfigAzure("id2", "log_connections", "off"),
					),
					wrapPsqlConfigResult(
						psqlConfigAzure("id3", "log_disconnections", "on"),
						psqlConfigAzure("id4", "connection_throttling", "off"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				psqlConfigAsset("id1", "log_checkpoints", "on"),
				psqlConfigAsset("id2", "log_connections", "off"),
				psqlConfigAsset("id3", "log_disconnections", "on"),
				psqlConfigAsset("id4", "connection_throttling", "off"),
			},
		},
		"Lower case values": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return wrapPsqlConfigResponse(
					wrapPsqlConfigResult(
						psqlConfigAzure("id1", "log_checkpoints", "ON"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				psqlConfigAsset("id1", "log_checkpoints", "on"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetPSQLConfiguration(tc.apiMockCall)
			got, err := p.ListPostgresConfigurations(context.Background(), "subId", "resourceGroup", "psqlInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func wrapPsqlConfigResponse(results ...armpostgresql.ConfigurationListResult) []armpostgresql.ConfigurationsClientListByServerResponse {
	return lo.Map(results, func(r armpostgresql.ConfigurationListResult, index int) armpostgresql.ConfigurationsClientListByServerResponse {
		return armpostgresql.ConfigurationsClientListByServerResponse{
			ConfigurationListResult: r,
		}
	})
}

func wrapPsqlConfigResult(configs ...*armpostgresql.Configuration) armpostgresql.ConfigurationListResult {
	return armpostgresql.ConfigurationListResult{
		Value: configs,
	}
}

func psqlConfigAzure(id, name, value string) *armpostgresql.Configuration {
	return &armpostgresql.Configuration{
		ID:   to.Ptr(id),
		Name: to.Ptr(name),
		Type: to.Ptr("psql/configurations"),
		Properties: &armpostgresql.ConfigurationProperties{
			Source:        to.Ptr("system-default"),
			Value:         to.Ptr(value),
			AllowedValues: to.Ptr("on,off"),
			DataType:      to.Ptr("Boolean"),
			DefaultValue:  to.Ptr("on"),
			Description:   to.Ptr("Value for config " + name),
		},
	}
}

func psqlConfigAsset(id string, name, value string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           name,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "psql/configurations",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"name":         name,
			"source":       "system-default",
			"value":        value,
			"dataType":     "Boolean",
			"defaultValue": "on",
		},
		Extension: nil,
	}
}
