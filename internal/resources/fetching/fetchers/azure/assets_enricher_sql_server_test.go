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
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestSQLServerEnricher_Enrich(t *testing.T) {
	tcs := map[string]struct {
		input               []inventory.AzureAsset
		expected            []inventory.AzureAsset
		expectError         bool
		epRes               map[string]enricherResponse
		bapRes              map[string]enricherResponse
		tdeRes              map[string]enricherResponse
		threatProtectionRes map[string]enricherResponse
		firewallRulesRes    map[string]enricherResponse
	}{
		"Some assets have encryption protection, others don't": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockSQLServer("id1", "serverName1"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
				}),
				mockOther("id4"),
				mockSQLServer("id2", "serverName2"),
			},
			epRes: map[string]enricherResponse{
				"serverName1": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
				"serverName2": noRes(),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"Multiple encryption protectors": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockSQLServer("id1", "serverName1"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
						mockEncryptionProtector("ep2", epProps("serverKey2", false)),
					},
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": assetRes(
					mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					mockEncryptionProtector("ep2", epProps("serverKey2", false))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
			},
		},
		"Error in one encryption protector": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				addExtension(mockSQLServer("id2", "serverName2"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": errorRes(errors.New("error")),
				"serverName2": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"Error in one blob audit policy": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				addExtension(mockSQLServer("id2", "serverName2"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": errorRes(errors.New("error")),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"Error in one threat protection": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				addExtension(mockSQLServer("id2", "serverName2"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": errorRes(errors.New("error")),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"Multiple transparent data encryption": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockSQLServer("id1", "serverName1"), map[string]any{
					inventory.ExtensionSQLTransparentDataEncryptions: []inventory.AzureAsset{
						mockTransparentDataEncryption("tde1", tdeProps("Enabled")),
						mockTransparentDataEncryption("tde2", tdeProps("Disabled")),
					},
				}),
				mockSQLServer("id2", "serverName2"),
			},
			epRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": assetRes(
					mockTransparentDataEncryption("tde1", tdeProps("Enabled")),
					mockTransparentDataEncryption("tde2", tdeProps("Disabled")),
				),
				"serverName2": noRes(),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"Error in one transparent data encryption": {
			expectError: true,
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				mockSQLServer("id2", "serverName2"),
			},
			expected: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
				addExtension(mockSQLServer("id2", "serverName2"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
					inventory.ExtensionSQLBlobAuditPolicy: mockBlobAuditingPolicies("ep2", bapProps("Enabled")),
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": assetRes(mockBlobAuditingPolicies("ep2", bapProps("Enabled"))),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": errorRes(errors.New("error")),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": noRes(),
				"serverName2": noRes(),
			},
		},
		"All enrichments": {
			input: []inventory.AzureAsset{
				mockSQLServer("id1", "serverName1"),
			},
			expected: []inventory.AzureAsset{
				addExtension(mockSQLServer("id1", "serverName1"), map[string]any{
					inventory.ExtensionSQLEncryptionProtectors: []inventory.AzureAsset{
						mockEncryptionProtector("ep1", epProps("serverKey1", true)),
					},
					inventory.ExtensionSQLBlobAuditPolicy: mockBlobAuditingPolicies("ep1", bapProps("Disabled")),
					inventory.ExtensionSQLTransparentDataEncryptions: []inventory.AzureAsset{
						mockTransparentDataEncryption("tde1", tdeProps("Enabled")),
					},
					inventory.ExtensionSQLAdvancedThreatProtectionSettings: []inventory.AzureAsset{
						mockThreatProtection("tde1", threatProtectionPros("Enabled")),
					},
					inventory.ExtensionSQLFirewallRules: []inventory.AzureAsset{
						mockFirewallRule("id1", "name1"),
					},
				}),
			},
			epRes: map[string]enricherResponse{
				"serverName1": assetRes(mockEncryptionProtector("ep1", epProps("serverKey1", true))),
			},
			bapRes: map[string]enricherResponse{
				"serverName1": assetRes(mockBlobAuditingPolicies("ep1", bapProps("Disabled"))),
			},
			tdeRes: map[string]enricherResponse{
				"serverName1": assetRes(mockTransparentDataEncryption("tde1", tdeProps("Enabled"))),
			},
			threatProtectionRes: map[string]enricherResponse{
				"serverName1": assetRes(mockThreatProtection("tde1", threatProtectionPros("Enabled"))),
			},
			firewallRulesRes: map[string]enricherResponse{
				"serverName1": assetRes(mockFirewallRule("id1", "name1")),
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			cmd := cycle.Metadata{}

			provider := azurelib.NewMockProviderAPI(t)
			for serverName, r := range tc.epRes {
				provider.EXPECT().ListSQLEncryptionProtector(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			for serverName, r := range tc.bapRes {
				provider.EXPECT().GetSQLBlobAuditingPolicies(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			for serverName, r := range tc.tdeRes {
				provider.EXPECT().ListSQLTransparentDataEncryptions(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			for serverName, r := range tc.threatProtectionRes {
				provider.EXPECT().ListSQLAdvancedThreatProtectionSettings(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			for serverName, r := range tc.firewallRulesRes {
				provider.EXPECT().ListSQLFirewallRules(mock.Anything, "subId", "group", serverName).Return(r.assets, r.err)
			}

			e := sqlServerEnricher{provider: provider}

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

func addExtension(a inventory.AzureAsset, ext map[string]any) inventory.AzureAsset {
	a.Extension = ext
	return a
}

func mockSQLServer(id, name string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:             id,
		SubscriptionId: "subId",
		ResourceGroup:  "group",
		Name:           name,
		Type:           inventory.SQLServersAssetType,
	}
}

func mockEncryptionProtector(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/encryptionProtector",
		Properties: props,
	}
}

func mockBlobAuditingPolicies(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/blobAuditPolicy",
		Properties: props,
	}
}

func mockTransparentDataEncryption(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/transparentDataEncryption",
		Properties: props,
	}
}

func mockThreatProtection(id string, props map[string]any) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:         id,
		Type:       inventory.SQLServersAssetType + "/threatProtection",
		Properties: props,
	}
}

func mockFirewallRule(id, name string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:   id,
		Name: name,
		Properties: map[string]any{
			"startIpAddress": "0.0.0.0",
			"endIpAddress":   "0.0.0.0",
		},
	}
}

func mockOther(id string) inventory.AzureAsset {
	return inventory.AzureAsset{
		Id:   id,
		Type: "otherType",
	}
}

func epProps(keyName string, rotationEnabled bool) map[string]any {
	return map[string]any{
		"kind":                "azurekeyvault",
		"serverKeyType":       "AzureKeyVault",
		"autoRotationEnabled": rotationEnabled,
		"serverKeyName":       keyName,
		"subregion":           "",
		"thumbprint":          "",
		"uri":                 "",
	}
}

func bapProps(state string) map[string]any {
	return map[string]any{
		"state":                        state,
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
	}
}

func tdeProps(state string) map[string]any {
	return map[string]any{
		"state": state,
	}
}

func threatProtectionPros(state string) map[string]any {
	creationTime, _ := time.Parse("2006-01-02", "2023-01-01")
	return map[string]any{
		"state":        state,
		"creationTime": creationTime,
	}
}
