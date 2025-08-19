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
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql/fake"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type (
	encryptionProtectorFn       func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error)
	auditingPoliciesFn          func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error)
	transparentDataEncryptionFn func(dbName string) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error)
	databaseFn                  func() ([]armsql.DatabasesClientListByServerResponse, error)
	threatProtectionFn          func() ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error)
)

func mockAssetEncryptionProtector(t *testing.T, f encryptionProtectorFn) SQLProviderAPI {
	wrapper := &sqlAzureClientWrapper{
		AssetEncryptionProtector: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armsql.EncryptionProtectorsClientListByServerOptions) ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
			return f()
		},
	}

	return &sqlProvider{
		log:    testhelper.NewLogger(t),
		client: wrapper,
	}
}

func mockAssetBlobAuditingPolicies(t *testing.T, f auditingPoliciesFn) SQLProviderAPI {
	wrapper := &sqlAzureClientWrapper{
		AssetBlobAuditingPolicies: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armsql.ServerBlobAuditingPoliciesClientGetOptions) (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
			return f()
		},
	}

	return &sqlProvider{
		log:    testhelper.NewLogger(t),
		client: wrapper,
	}
}

func mockAssetTransparentDataEncryption(t *testing.T, tdesFn transparentDataEncryptionFn, dbsFn databaseFn) SQLProviderAPI {
	wrapper := &sqlAzureClientWrapper{
		AssetDatabases: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armsql.DatabasesClientListByServerOptions) ([]armsql.DatabasesClientListByServerResponse, error) {
			return dbsFn()
		},
		AssetTransparentDataEncryptions: func(_ context.Context, _, _, _, dbName string, _ *arm.ClientOptions, _ *armsql.TransparentDataEncryptionsClientListByDatabaseOptions) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error) {
			return tdesFn(dbName)
		},
	}

	return &sqlProvider{
		log:    testhelper.NewLogger(t),
		client: wrapper,
	}
}

func mockAssetThreatProtection(t *testing.T, f threatProtectionFn) SQLProviderAPI {
	wrapper := &sqlAzureClientWrapper{
		AssetServerAdvancedThreatProtectionSettings: func(_ context.Context, _, _, _ string, _ *arm.ClientOptions, _ *armsql.ServerAdvancedThreatProtectionSettingsClientListByServerOptions) ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error) {
			return f()
		},
	}

	return &sqlProvider{
		log:    testhelper.NewLogger(t),
		client: wrapper,
	}
}

func TestListSQLEncryptionProtector(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    encryptionProtectorFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"No Encryption Protector Response": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with encryption protectors in different pages": {
			apiMockCall: func() ([]armsql.EncryptionProtectorsClientListByServerResponse, error) {
				return wrapEpResponse(
					wrapEpResult(
						epAzure("id1", armsql.ServerKeyTypeAzureKeyVault),
						epAzure("id2", armsql.ServerKeyTypeAzureKeyVault),
					),
					wrapEpResult(
						epAzure("id3", armsql.ServerKeyTypeServiceManaged),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				epAsset("id1", "AzureKeyVault"),
				epAsset("id2", "AzureKeyVault"),
				epAsset("id3", "ServiceManaged"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetEncryptionProtector(t, tc.apiMockCall)
			got, err := p.ListSQLEncryptionProtector(t.Context(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestGetSQLBlobAuditingPolicies(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    auditingPoliciesFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{}, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Response with blob auditing policy": {
			apiMockCall: func() (armsql.ServerBlobAuditingPoliciesClientGetResponse, error) {
				return armsql.ServerBlobAuditingPoliciesClientGetResponse{
					ServerBlobAuditingPolicy: armsql.ServerBlobAuditingPolicy{
						ID:   to.Ptr("id1"),
						Name: to.Ptr("policy"),
						Type: to.Ptr("audit-policy"),
						Properties: &armsql.ServerBlobAuditingPolicyProperties{
							State:                        to.Ptr(armsql.BlobAuditingPolicyStateEnabled),
							IsAzureMonitorTargetEnabled:  to.Ptr(true),
							IsDevopsAuditEnabled:         to.Ptr(false),
							IsManagedIdentityInUse:       to.Ptr(true),
							IsStorageSecondaryKeyInUse:   to.Ptr(true),
							QueueDelayMs:                 to.Ptr(int32(100)),
							RetentionDays:                to.Ptr(int32(90)),
							StorageAccountAccessKey:      to.Ptr("access-key"),
							StorageAccountSubscriptionID: to.Ptr("sub-id"),
							StorageEndpoint:              nil,
							AuditActionsAndGroups:        []*string{to.Ptr("a"), to.Ptr("b")},
						},
					},
				}, nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				{
					Id:             "id1",
					Name:           "policy",
					DisplayName:    "",
					Location:       "global",
					ResourceGroup:  "resourceGroup",
					SubscriptionId: "subId",
					Type:           "audit-policy",
					TenantId:       "",
					Sku:            nil,
					Identity:       nil,
					Properties: map[string]any{
						"state":                        "Enabled",
						"isAzureMonitorTargetEnabled":  true,
						"isDevopsAuditEnabled":         false,
						"isManagedIdentityInUse":       true,
						"isStorageSecondaryKeyInUse":   true,
						"queueDelayMs":                 int32(100),
						"retentionDays":                int32(90),
						"storageAccountAccessKey":      "access-key",
						"storageAccountSubscriptionID": "sub-id",
						"storageEndpoint":              "",
						"auditActionsAndGroups":        []string{"a", "b"},
					},
					Extension: nil,
				},
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetBlobAuditingPolicies(t, tc.apiMockCall)
			got, err := p.GetSQLBlobAuditingPolicies(t.Context(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestListSqlTransparentDataEncryptions(t *testing.T) {
	tcs := map[string]struct {
		tdeFn          transparentDataEncryptionFn
		dbFn           databaseFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on fetching databases": {
			dbFn: func() ([]armsql.DatabasesClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"Error on fetching 1 transparent data encryption (out of 3)": {
			dbFn: func() ([]armsql.DatabasesClientListByServerResponse, error) {
				return wrapDBResponse(
					[]string{"db1"},
					[]string{"db2", "db3"},
				), nil
			},
			tdeFn: func(dbName string) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error) {
				if dbName == "db2" {
					return nil, errors.New("error")
				}
				return wrapTdeResponse(
					wrapTdeResult(tdeAzure(dbName+"-tde1", armsql.TransparentDataEncryptionStateEnabled)),
				), nil
			},
			expectError: true,
			expectedAssets: []AzureAsset{
				tdeAsset("db1-tde1", "db1", "Enabled"),
				tdeAsset("db3-tde1", "db3", "Enabled"),
			},
		},
		"Response of 3 dbs with multiple tdes (in different pages)": {
			dbFn: func() ([]armsql.DatabasesClientListByServerResponse, error) {
				return wrapDBResponse(
					[]string{"db1"},
					[]string{"db2", "db3"},
				), nil
			},
			tdeFn: func(dbName string) ([]armsql.TransparentDataEncryptionsClientListByDatabaseResponse, error) {
				if dbName == "db1" {
					return wrapTdeResponse(
						wrapTdeResult(tdeAzure(dbName+"-tde1", armsql.TransparentDataEncryptionStateEnabled)),
						wrapTdeResult(tdeAzure(dbName+"-tde2", armsql.TransparentDataEncryptionStateDisabled)),
					), nil
				}
				return wrapTdeResponse(
					wrapTdeResult(tdeAzure(dbName+"-tde1", armsql.TransparentDataEncryptionStateEnabled)),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				tdeAsset("db1-tde1", "db1", "Enabled"),
				tdeAsset("db1-tde2", "db1", "Disabled"),
				tdeAsset("db2-tde1", "db2", "Enabled"),
				tdeAsset("db3-tde1", "db3", "Enabled"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetTransparentDataEncryption(t, tc.tdeFn, tc.dbFn)
			got, err := p.ListSQLTransparentDataEncryptions(t.Context(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func TestListSQLAdvancedThreatProtectionSettings(t *testing.T) {
	tcs := map[string]struct {
		apiMockCall    threatProtectionFn
		expectError    bool
		expectedAssets []AzureAsset
	}{
		"Error on calling api": {
			apiMockCall: func() ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error) {
				return nil, errors.New("error")
			},
			expectError:    true,
			expectedAssets: nil,
		},
		"No threat protection settings": {
			apiMockCall: func() ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error) {
				return nil, nil
			},
			expectError:    false,
			expectedAssets: nil,
		},
		"Response with threat protection in different pages": {
			apiMockCall: func() ([]armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse, error) {
				return wrapThreatProtectionResponse(
					wrapThreatProtectionResult(
						threatProtectionAzure("id1", armsql.AdvancedThreatProtectionStateEnabled),
						threatProtectionAzure("id2", armsql.AdvancedThreatProtectionStateDisabled),
					),
					wrapThreatProtectionResult(
						threatProtectionAzure("id3", armsql.AdvancedThreatProtectionStateEnabled),
					),
				), nil
			},
			expectError: false,
			expectedAssets: []AzureAsset{
				threatProtectionAsset("id1", "Enabled"),
				threatProtectionAsset("id2", "Disabled"),
				threatProtectionAsset("id3", "Enabled"),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			p := mockAssetThreatProtection(t, tc.apiMockCall)
			got, err := p.ListSQLAdvancedThreatProtectionSettings(t.Context(), "subId", "resourceGroup", "sqlServerInstanceName")

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedAssets, got)
		})
	}
}

func wrapEpResult(eps ...*armsql.EncryptionProtector) armsql.EncryptionProtectorListResult {
	return armsql.EncryptionProtectorListResult{
		Value: eps,
	}
}

func wrapEpResponse(results ...armsql.EncryptionProtectorListResult) []armsql.EncryptionProtectorsClientListByServerResponse {
	return lo.Map(results, func(r armsql.EncryptionProtectorListResult, _ int) armsql.EncryptionProtectorsClientListByServerResponse {
		return armsql.EncryptionProtectorsClientListByServerResponse{
			EncryptionProtectorListResult: r,
		}
	})
}

func epAzure(id string, keyType armsql.ServerKeyType) *armsql.EncryptionProtector {
	return &armsql.EncryptionProtector{
		ID:       to.Ptr(id),
		Name:     to.Ptr("Name " + id),
		Kind:     to.Ptr("azurekeyvault"),
		Location: to.Ptr("eu-west"),
		Type:     to.Ptr("encryptionProtector"),
		Properties: &armsql.EncryptionProtectorProperties{
			ServerKeyType:       to.Ptr(keyType),
			AutoRotationEnabled: to.Ptr(true),
			ServerKeyName:       to.Ptr("key-" + id),
			Subregion:           to.Ptr("eu-west-1"),
		},
	}
}

func epAsset(id, keyType string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           "Name " + id,
		DisplayName:    "",
		Location:       "eu-west",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "encryptionProtector",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"kind":                "azurekeyvault",
			"serverKeyType":       keyType,
			"autoRotationEnabled": true,
			"serverKeyName":       "key-" + id,
			"subregion":           "eu-west-1",
			"thumbprint":          "",
			"uri":                 "",
		},
		Extension: nil,
	}
}

func wrapDBResponse(ids ...[]string) []armsql.DatabasesClientListByServerResponse {
	return lo.Map(ids, func(ids []string, _ int) armsql.DatabasesClientListByServerResponse {
		values := lo.Map(ids, func(id string, _ int) *armsql.Database {
			return &armsql.Database{
				Name: to.Ptr(id),
			}
		})

		return armsql.DatabasesClientListByServerResponse{
			DatabaseListResult: armsql.DatabaseListResult{
				Value: values,
			},
		}
	})
}

func wrapTdeResponse(results ...armsql.LogicalDatabaseTransparentDataEncryptionListResult) []armsql.TransparentDataEncryptionsClientListByDatabaseResponse {
	return lo.Map(results, func(r armsql.LogicalDatabaseTransparentDataEncryptionListResult, _ int) armsql.TransparentDataEncryptionsClientListByDatabaseResponse {
		return armsql.TransparentDataEncryptionsClientListByDatabaseResponse{
			LogicalDatabaseTransparentDataEncryptionListResult: r,
		}
	})
}

func wrapTdeResult(tdes ...*armsql.LogicalDatabaseTransparentDataEncryption) armsql.LogicalDatabaseTransparentDataEncryptionListResult {
	return armsql.LogicalDatabaseTransparentDataEncryptionListResult{
		Value: tdes,
	}
}

func tdeAzure(id string, state armsql.TransparentDataEncryptionState) *armsql.LogicalDatabaseTransparentDataEncryption {
	return &armsql.LogicalDatabaseTransparentDataEncryption{
		ID:   to.Ptr(id),
		Name: to.Ptr("name-" + id),
		Type: to.Ptr("transparentDataEncryption"),
		Properties: &armsql.TransparentDataEncryptionProperties{
			State: to.Ptr(state),
		},
	}
}

func tdeAsset(id, dbName, state string) AzureAsset {
	return AzureAsset{
		Id:             id,
		Name:           "name-" + id,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "transparentDataEncryption",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"databaseName": dbName,
			"state":        state,
		},
		Extension: nil,
	}
}

func wrapThreatProtectionResponse(results ...armsql.LogicalServerAdvancedThreatProtectionListResult) []armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse {
	return lo.Map(results, func(r armsql.LogicalServerAdvancedThreatProtectionListResult, _ int) armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse {
		return armsql.ServerAdvancedThreatProtectionSettingsClientListByServerResponse{
			LogicalServerAdvancedThreatProtectionListResult: r,
		}
	})
}

func wrapThreatProtectionResult(tps ...*armsql.ServerAdvancedThreatProtection) armsql.LogicalServerAdvancedThreatProtectionListResult {
	return armsql.LogicalServerAdvancedThreatProtectionListResult{
		Value: tps,
	}
}

func threatProtectionAzure(id string, state armsql.AdvancedThreatProtectionState) *armsql.ServerAdvancedThreatProtection {
	creationTime, _ := time.Parse("2006-01-02", "2023-01-01")
	return &armsql.ServerAdvancedThreatProtection{
		ID:   to.Ptr(id),
		Name: to.Ptr("name-" + id),
		Type: to.Ptr("serverAdvancedThreatProtection"),
		Properties: &armsql.AdvancedThreatProtectionProperties{
			State:        to.Ptr(state),
			CreationTime: to.Ptr(creationTime),
		},
	}
}

func threatProtectionAsset(id, state string) AzureAsset {
	creationTime, _ := time.Parse("2006-01-02", "2023-01-01")
	return AzureAsset{
		Id:             id,
		Name:           "name-" + id,
		DisplayName:    "",
		Location:       "global",
		ResourceGroup:  "resourceGroup",
		SubscriptionId: "subId",
		Type:           "serverAdvancedThreatProtection",
		TenantId:       "",
		Sku:            nil,
		Identity:       nil,
		Properties: map[string]any{
			"state":        state,
			"creationTime": creationTime,
		},
		Extension: nil,
	}
}

func TestListSQLFirewallRules(t *testing.T) {
	subID := "11111111-aaaa-bbbb-cccc-dddddddddddd"
	resourceGroup := "rg"
	srv := "srv"

	tests := map[string]struct {
		mockPages [][]*armsql.FirewallRule
		expected  []AzureAsset
	}{
		"single page": {
			mockPages: [][]*armsql.FirewallRule{
				{
					{
						Name: to.Ptr("name1"),
					},
					{
						Name: to.Ptr("name2"),
						Properties: &armsql.ServerFirewallRuleProperties{
							StartIPAddress: to.Ptr("0.0.0.0"),
							EndIPAddress:   to.Ptr("0.0.0.0"),
						},
					},
				},
			},
			expected: []AzureAsset{
				{
					Name:           "name1",
					SubscriptionId: subID,
					ResourceGroup:  resourceGroup,
				},
				{
					Name:           "name2",
					SubscriptionId: subID,
					ResourceGroup:  resourceGroup,
					Properties: map[string]any{
						"startIpAddress": "0.0.0.0",
						"endIpAddress":   "0.0.0.0",
					},
				},
			},
		},
		"two pages": {
			mockPages: [][]*armsql.FirewallRule{
				{
					{
						Name: to.Ptr("name1"),
					},
					{
						Name: to.Ptr("name2"),
						Properties: &armsql.ServerFirewallRuleProperties{
							StartIPAddress: to.Ptr("0.0.0.0"),
							EndIPAddress:   to.Ptr("0.0.0.0"),
						},
					},
				},
				{
					{
						Name: to.Ptr("name3"),
						Properties: &armsql.ServerFirewallRuleProperties{
							StartIPAddress: to.Ptr("0.0.0.0"),
							EndIPAddress:   to.Ptr("0.0.0.0"),
						},
					},
				},
			},
			expected: []AzureAsset{
				{
					Name:           "name1",
					SubscriptionId: subID,
					ResourceGroup:  resourceGroup,
				},
				{
					Name:           "name2",
					SubscriptionId: subID,
					ResourceGroup:  resourceGroup,
					Properties: map[string]any{
						"startIpAddress": "0.0.0.0",
						"endIpAddress":   "0.0.0.0",
					},
				},
				{
					Name:           "name3",
					SubscriptionId: subID,
					ResourceGroup:  resourceGroup,
					Properties: map[string]any{
						"startIpAddress": "0.0.0.0",
						"endIpAddress":   "0.0.0.0",
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fakeSrv := &fake.FirewallRulesServer{}
			fakeSrv.NewListByServerPager = func(_, _ string, _ *armsql.FirewallRulesClientListByServerOptions) azfake.PagerResponder[armsql.FirewallRulesClientListByServerResponse] {
				pager := azfake.PagerResponder[armsql.FirewallRulesClientListByServerResponse]{}

				for _, p := range tc.mockPages {
					page := armsql.FirewallRulesClientListByServerResponse{
						FirewallRuleListResult: armsql.FirewallRuleListResult{
							Value: p,
						},
					}
					pager.AddPage(http.StatusOK, page, nil)
				}

				return pager
			}
			fakeTransport := fake.NewFirewallRulesServerTransport(fakeSrv)

			provider := NewSQLProvider(testhelper.NewLogger(t), nil).(*sqlProvider)
			provider.clientOptions = &arm.ClientOptions{}
			provider.clientOptions.Transport = fakeTransport

			rules, err := provider.ListSQLFirewallRules(t.Context(), subID, resourceGroup, srv)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, rules)
		})
	}
}
