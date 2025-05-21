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

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type psqlServerType string

const (
	psqlServerTypeSingle   psqlServerType = "single"
	psqlServerTypeFlexible psqlServerType = "flexible"
)

type psqlEnricherResponse struct {
	enricherResponse
	serverType psqlServerType
}

func TestPostgresqlEnricher_Enrich(t *testing.T) {
	tcs := map[string]struct {
		input            []inventory.AzureAsset
		expected         []inventory.AzureAsset
		expectError      bool
		configRes        map[string]psqlEnricherResponse
		firewallRulesRes map[string]psqlEnricherResponse
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
					inventory.ExtensionPostgresqlConfigurations: []inventory.AzureAsset{
						mockPostgresConfigAsset("log_checkpoints", psqlConfigProps("on")),
						mockPostgresConfigAsset("log_connections", psqlConfigProps("off")),
					},
				}),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			configRes: map[string]psqlEnricherResponse{
				"psql-a": {assetRes(
					mockPostgresConfigAsset("log_checkpoints", psqlConfigProps("on")),
					mockPostgresConfigAsset("log_connections", psqlConfigProps("off")),
				), psqlServerTypeSingle},
				"psql-b": {
					errorRes(errors.New("error")), psqlServerTypeSingle,
				},
			},
			firewallRulesRes: map[string]psqlEnricherResponse{
				"psql-a": {serverType: psqlServerTypeSingle}, //nolint:exhaustruct
				"psql-b": {serverType: psqlServerTypeSingle}, //nolint:exhaustruct
			},
		},
		"Error on enriching firewall rules": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockPostgresAsset("id1", "psql-a"),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockPostgresAsset("id1", "psql-a"), map[string]any{
					inventory.ExtensionPostgresqlFirewallRules: []inventory.AzureAsset{
						mockPostgresFirewallRuleAsset("fr1", psqlFirewallRuleProps("name-fr1", "0.0.0.0", "196.198.198.256")),
					},
				}),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
			},
			configRes: map[string]psqlEnricherResponse{
				"psql-a": {serverType: psqlServerTypeSingle}, //nolint:exhaustruct
				"psql-b": {serverType: psqlServerTypeSingle}, //nolint:exhaustruct
			},
			firewallRulesRes: map[string]psqlEnricherResponse{
				"psql-a": {assetRes(mockPostgresFirewallRuleAsset("fr1", psqlFirewallRuleProps("name-fr1", "0.0.0.0", "196.198.198.256"))), psqlServerTypeSingle},
				"psql-b": {errorRes(errors.New("error")), psqlServerTypeSingle},
			},
		},
		"Enrich configs and firewall rules": {
			expectError: false,
			input: []inventory.AzureAsset{
				mockPostgresAsset("id1", "psql-a"),
				mockOther("id2"),
				mockPostgresAsset("id3", "psql-b"),
				mockFlexiblePostgresAsset("id4", "flex-psql-a"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockPostgresAsset("id1", "psql-a"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []inventory.AzureAsset{
						mockPostgresConfigAsset("log_checkpoints", psqlConfigProps("on")),
						mockPostgresConfigAsset("log_connections", psqlConfigProps("off")),
					},
					inventory.ExtensionPostgresqlFirewallRules: []inventory.AzureAsset{
						mockPostgresFirewallRuleAsset("fr1", psqlFirewallRuleProps("name-fr1", "0.0.0.0", "196.198.198.256")),
					},
				}),
				mockOther("id2"),
				addExtension(mockPostgresAsset("id3", "psql-b"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []inventory.AzureAsset{
						mockPostgresConfigAsset("log_disconnections", psqlConfigProps("on")),
						mockPostgresConfigAsset("connection_throttling", psqlConfigProps("off")),
					},
					inventory.ExtensionPostgresqlFirewallRules: []inventory.AzureAsset{
						mockPostgresFirewallRuleAsset("fr2", psqlFirewallRuleProps("name-fr2", "0.0.0.1", "196.198.198.255")),
					},
				}),
				addExtension(mockFlexiblePostgresAsset("id4", "flex-psql-a"), map[string]any{
					inventory.ExtensionPostgresqlConfigurations: []inventory.AzureAsset{
						mockPostgresConfigAsset("log_disconnections", psqlConfigProps("on")),
					},
					inventory.ExtensionPostgresqlFirewallRules: []inventory.AzureAsset{
						mockPostgresFirewallRuleAsset("fr3", psqlFirewallRuleProps("name-fr3", "0.0.0.2", "196.198.198.254")),
					},
				}),
			},
			configRes: map[string]psqlEnricherResponse{
				"psql-a": {
					assetRes(
						mockPostgresConfigAsset("log_checkpoints", psqlConfigProps("on")),
						mockPostgresConfigAsset("log_connections", psqlConfigProps("off"))),
					psqlServerTypeSingle,
				},
				"psql-b": {
					assetRes(
						mockPostgresConfigAsset("log_disconnections", psqlConfigProps("on")),
						mockPostgresConfigAsset("connection_throttling", psqlConfigProps("off")),
					), psqlServerTypeSingle,
				},
				"flex-psql-a": {
					assetRes(
						mockPostgresConfigAsset("log_disconnections", psqlConfigProps("on"))),
					psqlServerTypeFlexible,
				},
			},
			firewallRulesRes: map[string]psqlEnricherResponse{
				"psql-a":      {assetRes(mockPostgresFirewallRuleAsset("fr1", psqlFirewallRuleProps("name-fr1", "0.0.0.0", "196.198.198.256"))), psqlServerTypeSingle},
				"psql-b":      {assetRes(mockPostgresFirewallRuleAsset("fr2", psqlFirewallRuleProps("name-fr2", "0.0.0.1", "196.198.198.255"))), psqlServerTypeSingle},
				"flex-psql-a": {assetRes(mockPostgresFirewallRuleAsset("fr3", psqlFirewallRuleProps("name-fr3", "0.0.0.2", "196.198.198.254"))), psqlServerTypeFlexible},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			for serverName, r := range tc.configRes {
				if r.serverType == psqlServerTypeSingle {
					provider.EXPECT().ListSinglePostgresConfigurations(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
				} else {
					provider.EXPECT().ListFlexiblePostgresConfigurations(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
				}
			}

			for serverName, r := range tc.firewallRulesRes {
				if r.serverType == psqlServerTypeSingle {
					provider.EXPECT().ListSinglePostgresFirewallRules(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
				} else {
					provider.EXPECT().ListFlexiblePostgresFirewallRules(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
				}
			}

			e := postgresqlEnricher{provider: provider}

			err := e.Enrich(t.Context(), cmd, tc.input)
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
	return mockAsset(id, name, inventory.PostgreSQLDBAssetType)
}

func mockFlexiblePostgresAsset(id, name string) inventory.AzureAsset {
	return mockAsset(id, name, inventory.FlexiblePostgreSQLDBAssetType)
}

func mockAsset(id, name, assetType string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           name,
		Type:           assetType,
	}
}

func mockPostgresConfigAsset(id string, props map[string]any) inventory.AzureAsset {
	a := mockAsset(id, "name-"+id, inventory.PostgreSQLDBAssetType+"/configuration")
	a.Properties = props
	return a
}

func mockPostgresFirewallRuleAsset(id string, props map[string]any) inventory.AzureAsset {
	a := mockAsset(id, "name-"+id, inventory.PostgreSQLDBAssetType+"/configuration")
	a.Properties = props
	return a
}

func psqlConfigProps(value string) map[string]any {
	return map[string]any{
		"source":       "system-default",
		"value":        value,
		"dataType":     "Boolean",
		"defaultValue": "on",
	}
}

func psqlFirewallRuleProps(name, startIpAddr, endIpAddr string) map[string]any {
	return map[string]any{
		"name":           name,
		"startIPAddress": startIpAddr,
		"endIPAddress":   endIpAddr,
	}
}
