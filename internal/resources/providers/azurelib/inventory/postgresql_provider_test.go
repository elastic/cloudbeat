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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type (
	assetSingleConfigFn       func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error)
	assetFlexConfigFn         func() ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error)
	assetSingleFirewallRuleFn func() ([]armpostgresql.FirewallRulesClientListByServerResponse, error)
	assetFlexFirewallRuleFn   func() ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error)
)

func mockAssetSinglePSQLConfiguration(f assetSingleConfigFn) PostgresqlProviderAPI {
	cl := &psqlAzureClientWrapper{
		AssetSingleServerConfigurations: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armpostgresql.ConfigurationsClientListByServerOptions) ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
			return f()
		},
	}

	return &psqlProvider{
		log:    clog.NewLogger("mock_single_psql_config"),
		client: cl,
	}
}

func mockAssetFlexPSQLConfiguration(f assetFlexConfigFn) PostgresqlProviderAPI {
	cl := &psqlAzureClientWrapper{
		AssetFlexibleServerConfigurations: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armpostgresqlflexibleservers.ConfigurationsClientListByServerOptions) ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
			return f()
		},
	}

	return &psqlProvider{
		log:    clog.NewLogger("mock_flex_psql_config"),
		client: cl,
	}
}

func mockAssetSinglePSQLFirewallRule(f assetSingleFirewallRuleFn) PostgresqlProviderAPI {
	cl := &psqlAzureClientWrapper{
		AssetSingleServerFirewallRules: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armpostgresql.FirewallRulesClientListByServerOptions) ([]armpostgresql.FirewallRulesClientListByServerResponse, error) {
			return f()
		},
	}

	return &psqlProvider{
		log:    clog.NewLogger("mock_single_psql_firewall_rules"),
		client: cl,
	}
}

func mockAssetFlexPSQLFirewallRule(f assetFlexFirewallRuleFn) PostgresqlProviderAPI {
	cl := &psqlAzureClientWrapper{
		AssetFlexibleServerFirewallRules: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armpostgresqlflexibleservers.FirewallRulesClientListByServerOptions) ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error) {
			return f()
		},
	}

	return &psqlProvider{
		log:    clog.NewLogger("mock_flexs_psql_firewall_rules"),
		client: cl,
	}
}

func TestListSinglePostgresConfigurations(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    assetSingleConfigFn
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
		"Empty Response": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return wrapSinglePsqlConfigResponse(
					wrapSinglePsqlConfigResult(
						singlePsqlConfigAzure("id1", "log_checkpoints", "on"),
						singlePsqlConfigAzure("id2", "log_connections", "off"),
					),
					wrapSinglePsqlConfigResult(
						singlePsqlConfigAzure("id3", "log_disconnections", "on"),
						singlePsqlConfigAzure("id4", "connection_throttling", "off"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				singlePsqlConfigAsset("id1", "log_checkpoints", "on"),
				singlePsqlConfigAsset("id2", "log_connections", "off"),
				singlePsqlConfigAsset("id3", "log_disconnections", "on"),
				singlePsqlConfigAsset("id4", "connection_throttling", "off"),
			},
		},
		"Lower case values": {
			apiMockCall: func() ([]armpostgresql.ConfigurationsClientListByServerResponse, error) {
				return wrapSinglePsqlConfigResponse(
					wrapSinglePsqlConfigResult(
						singlePsqlConfigAzure("id1", "log_checkpoints", "ON"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				singlePsqlConfigAsset("id1", "log_checkpoints", "on"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetSinglePSQLConfiguration(tc.apiMockCall)
			got, err := p.ListSinglePostgresConfigurations(context.Background(), "subId", "resourceGroup", "psqlInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestListFlexiblePostgresConfigurations(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    assetFlexConfigFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Empty Response": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
				return wrapFlexPsqlConfigResponse(
					wrapFlexPsqlConfigResult(
						flexPsqlConfigAzure("id1", "log_checkpoints", "on"),
						flexPsqlConfigAzure("id2", "log_connections", "off"),
					),
					wrapFlexPsqlConfigResult(
						flexPsqlConfigAzure("id3", "log_disconnections", "on"),
						flexPsqlConfigAzure("id4", "connection_throttling", "off"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				flexPsqlConfigAsset("id1", "log_checkpoints", "on"),
				flexPsqlConfigAsset("id2", "log_connections", "off"),
				flexPsqlConfigAsset("id3", "log_disconnections", "on"),
				flexPsqlConfigAsset("id4", "connection_throttling", "off"),
			},
		},
		"Lower case values": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse, error) {
				return wrapFlexPsqlConfigResponse(
					wrapFlexPsqlConfigResult(
						flexPsqlConfigAzure("id1", "log_checkpoints", "ON"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				flexPsqlConfigAsset("id1", "log_checkpoints", "on"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetFlexPSQLConfiguration(tc.apiMockCall)
			got, err := p.ListFlexiblePostgresConfigurations(context.Background(), "subId", "resourceGroup", "psqlInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestListSinglePostgresFirewallRules(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    assetSingleFirewallRuleFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armpostgresql.FirewallRulesClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Empty Response": {
			apiMockCall: func() ([]armpostgresql.FirewallRulesClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armpostgresql.FirewallRulesClientListByServerResponse, error) {
				return wrapSinglePsqlFirewallRulesResponse(
					wrapSinglePsqlFirewallRulesResult(
						singlePsqlFirewallRuleAzure("id1", "0.0.0.0", "196.81.61.0"),
						singlePsqlFirewallRuleAzure("id2", "199.32.26.89", "156.12.92.0"),
					),
					wrapSinglePsqlFirewallRulesResult(
						singlePsqlFirewallRuleAzure("id3", "0.0.5.0", "56.12.98.88"),
						singlePsqlFirewallRuleAzure("id4", "255.255.255.1", "12.28.19.1"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				singlePsqlFirewallConfigAsset("id1", "0.0.0.0", "196.81.61.0"),
				singlePsqlFirewallConfigAsset("id2", "199.32.26.89", "156.12.92.0"),
				singlePsqlFirewallConfigAsset("id3", "0.0.5.0", "56.12.98.88"),
				singlePsqlFirewallConfigAsset("id4", "255.255.255.1", "12.28.19.1"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetSinglePSQLFirewallRule(tc.apiMockCall)
			got, err := p.ListSinglePostgresFirewallRules(context.Background(), "subId", "resourceGroup", "psqlInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestFlexibleSinglePostgresFirewallRules(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    assetFlexFirewallRuleFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Empty Response": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse, error) {
				return wrapFlexPsqlFirewallRulesResponse(
					wrapFlexPsqlFirewallRulesResult(
						flexPsqlFirewallRuleAzure("id1", "0.0.0.0", "196.81.61.0"),
						flexPsqlFirewallRuleAzure("id2", "199.32.26.89", "156.12.92.0"),
					),
					wrapFlexPsqlFirewallRulesResult(
						flexPsqlFirewallRuleAzure("id3", "0.0.5.0", "56.12.98.88"),
						flexPsqlFirewallRuleAzure("id4", "255.255.255.1", "12.28.19.1"),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				flexPsqlFirewallConfigAsset("id1", "0.0.0.0", "196.81.61.0"),
				flexPsqlFirewallConfigAsset("id2", "199.32.26.89", "156.12.92.0"),
				flexPsqlFirewallConfigAsset("id3", "0.0.5.0", "56.12.98.88"),
				flexPsqlFirewallConfigAsset("id4", "255.255.255.1", "12.28.19.1"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetFlexPSQLFirewallRule(tc.apiMockCall)
			got, err := p.ListFlexiblePostgresFirewallRules(context.Background(), "subId", "resourceGroup", "psqlInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func wrapSinglePsqlConfigResponse(results ...armpostgresql.ConfigurationListResult) []armpostgresql.ConfigurationsClientListByServerResponse {
	return lo.Map(results, func(r armpostgresql.ConfigurationListResult, _ int) armpostgresql.ConfigurationsClientListByServerResponse {
		return armpostgresql.ConfigurationsClientListByServerResponse{
			ConfigurationListResult: r,
		}
	})
}

func wrapSinglePsqlConfigResult(configs ...*armpostgresql.Configuration) armpostgresql.ConfigurationListResult {
	return armpostgresql.ConfigurationListResult{
		Value: configs,
	}
}

func singlePsqlConfigAzure(id, name, value string) *armpostgresql.Configuration {
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

func singlePsqlConfigAsset(id, name, value string) AzureAsset {
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
			"source":       "system-default",
			"value":        value,
			"dataType":     "Boolean",
			"defaultValue": "on",
		},
		Extension: nil,
	}
}

func wrapFlexPsqlConfigResponse(results ...armpostgresqlflexibleservers.ConfigurationListResult) []armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse {
	return lo.Map(results, func(r armpostgresqlflexibleservers.ConfigurationListResult, _ int) armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse {
		return armpostgresqlflexibleservers.ConfigurationsClientListByServerResponse{
			ConfigurationListResult: r,
		}
	})
}

func wrapFlexPsqlConfigResult(configs ...*armpostgresqlflexibleservers.Configuration) armpostgresqlflexibleservers.ConfigurationListResult {
	return armpostgresqlflexibleservers.ConfigurationListResult{
		Value: configs,
	}
}

func flexPsqlConfigAzure(id, name, value string) *armpostgresqlflexibleservers.Configuration {
	return &armpostgresqlflexibleservers.Configuration{
		ID:   to.Ptr(id),
		Name: to.Ptr(name),
		Type: to.Ptr("flex-psql/configurations"),
		Properties: &armpostgresqlflexibleservers.ConfigurationProperties{
			Source:        to.Ptr("system-default"),
			Value:         to.Ptr(value),
			AllowedValues: to.Ptr("on,off"),
			DataType:      to.Ptr(armpostgresqlflexibleservers.ConfigurationDataTypeBoolean),
			DefaultValue:  to.Ptr("on"),
			Description:   to.Ptr("Value for config " + name),
		},
	}
}

func flexPsqlConfigAsset(id, name, value string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           name,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "flex-psql/configurations",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"source":       "system-default",
			"value":        value,
			"dataType":     "Boolean",
			"defaultValue": "on",
		},
		Extension: nil,
	}
}

func wrapSinglePsqlFirewallRulesResponse(results ...armpostgresql.FirewallRuleListResult) []armpostgresql.FirewallRulesClientListByServerResponse {
	return lo.Map(results, func(r armpostgresql.FirewallRuleListResult, _ int) armpostgresql.FirewallRulesClientListByServerResponse {
		return armpostgresql.FirewallRulesClientListByServerResponse{
			FirewallRuleListResult: r,
		}
	})
}

func wrapSinglePsqlFirewallRulesResult(rules ...*armpostgresql.FirewallRule) armpostgresql.FirewallRuleListResult {
	return armpostgresql.FirewallRuleListResult{
		Value: rules,
	}
}

func singlePsqlFirewallRuleAzure(id, startIpAddr, endIpAddr string) *armpostgresql.FirewallRule {
	return &armpostgresql.FirewallRule{
		ID:   to.Ptr(id),
		Name: to.Ptr("name-" + id),
		Type: to.Ptr("psql/firewall-rule"),
		Properties: &armpostgresql.FirewallRuleProperties{
			StartIPAddress: to.Ptr(startIpAddr),
			EndIPAddress:   to.Ptr(endIpAddr),
		},
	}
}

func singlePsqlFirewallConfigAsset(id, startIpAddr, endIpAddr string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           "name-" + id,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "psql/firewall-rule",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"startIPAddress": startIpAddr,
			"endIPAddress":   endIpAddr,
		},
		Extension: nil,
	}
}

func wrapFlexPsqlFirewallRulesResponse(results ...armpostgresqlflexibleservers.FirewallRuleListResult) []armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse {
	return lo.Map(results, func(r armpostgresqlflexibleservers.FirewallRuleListResult, _ int) armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse {
		return armpostgresqlflexibleservers.FirewallRulesClientListByServerResponse{
			FirewallRuleListResult: r,
		}
	})
}

func wrapFlexPsqlFirewallRulesResult(rules ...*armpostgresqlflexibleservers.FirewallRule) armpostgresqlflexibleservers.FirewallRuleListResult {
	return armpostgresqlflexibleservers.FirewallRuleListResult{
		Value: rules,
	}
}

func flexPsqlFirewallRuleAzure(id, startIpAddr, endIpAddr string) *armpostgresqlflexibleservers.FirewallRule {
	return &armpostgresqlflexibleservers.FirewallRule{
		ID:   to.Ptr(id),
		Name: to.Ptr("name-" + id),
		Type: to.Ptr("flex-psql/firewall-rule"),
		Properties: &armpostgresqlflexibleservers.FirewallRuleProperties{
			StartIPAddress: to.Ptr(startIpAddr),
			EndIPAddress:   to.Ptr(endIpAddr),
		},
	}
}

func flexPsqlFirewallConfigAsset(id, startIpAddr, endIpAddr string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           "name-" + id,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "flex-psql/firewall-rule",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"startIPAddress": startIpAddr,
			"endIPAddress":   endIpAddr,
		},
		Extension: nil,
	}
}
