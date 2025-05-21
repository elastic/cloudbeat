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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

func TestEnrichVirtualMachinesWithNetworkSecurityGroup(t *testing.T) {
	enricher := vmNetworkSecurityGroupEnricher{}

	tests := map[string]struct {
		input  []inventory.AzureAsset
		output []inventory.AzureAsset
	}{
		"1 NSG maps to 1 VM (and one NSG without interface)": {
			input: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{}, withSecRules(secRule("8080", "Block"))),
			},
			output: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}, withExtendedSecRules(extendedSecRule("22", "Allow"))),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{}, withSecRules(secRule("8080", "Block"))),
			},
		},

		"2 NSG map to 1 VM, 1 NSG doesn't match": {
			input: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1", "nic#3"}),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#2"}, withSecRules(secRule("80", "Block"))),
				mockNSGs("nsg#3", []string{"nic#3"}, withSecRules(secRule("8080", "Block"))),
			},
			output: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1", "nic#3"}, withExtendedSecRules(
					extendedSecRule("22", "Allow"),
					extendedSecRule("8080", "Block"),
				)),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#2"}, withSecRules(secRule("80", "Block"))),
				mockNSGs("nsg#3", []string{"nic#3"}, withSecRules(secRule("8080", "Block"))),
			},
		},

		"1 NSG maps to 2 VMS (and NSG without rule)": {
			input: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}),
				mockVMs("vm#2", []string{"nic#1"}),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#1"}),
			},
			output: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}, withExtendedSecRules(extendedSecRule("22", "Allow"))),
				mockVMs("vm#2", []string{"nic#1"}, withExtendedSecRules(extendedSecRule("22", "Allow"))),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#1"}),
			},
		},

		"2 NSG with same interface to 1 VM (and VM without interface)": {
			input: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}),
				mockVMs("vm#2", []string{}),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#1"}, withSecRules(secRule("8080", "Block"))),
			},
			output: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1"}, withExtendedSecRules(
					extendedSecRule("22", "Allow"),
					extendedSecRule("8080", "Block"),
				)),
				mockVMs("vm#2", []string{}),
				mockNSGs("nsg#1", []string{"nic#1"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#1"}, withSecRules(secRule("8080", "Block"))),
			},
		},

		"Multiple NSGs to Multiple VMS": {
			input: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1", "nic#2"}),
				mockVMs("vm#2", []string{"nic#5", "nic#4", "nic#6"}),
				mockVMs("vm#3", []string{"nic#3"}),
				mockNSGs("nsg#1", []string{"nic#1", "nic#3"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#2", "nic#5"}, withSecRules(secRule("80", "Block"), secRule("7000", "Block"))),
				mockNSGs("nsg#3", []string{"nic#4"}, withSecRules(secRule("8080", "Allow"))),
				mockNSGs("nsg#4", []string{"nic#6"}, withSecRules(secRule("7812", "Block"))),
			},

			output: []inventory.AzureAsset{
				mockVMs("vm#1", []string{"nic#1", "nic#2"}, withExtendedSecRules(
					extendedSecRule("22", "Allow"),
					extendedSecRule("80", "Block"),
					extendedSecRule("7000", "Block"),
				)),
				mockVMs("vm#2", []string{"nic#5", "nic#4", "nic#6"}, withExtendedSecRules(
					extendedSecRule("80", "Block"),
					extendedSecRule("7000", "Block"),
					extendedSecRule("8080", "Allow"),
					extendedSecRule("7812", "Block"),
				)),
				mockVMs("vm#3", []string{"nic#3"}, withExtendedSecRules(
					extendedSecRule("22", "Allow"),
				)),
				mockNSGs("nsg#1", []string{"nic#1", "nic#3"}, withSecRules(secRule("22", "Allow"))),
				mockNSGs("nsg#2", []string{"nic#2", "nic#5"}, withSecRules(secRule("80", "Block"), secRule("7000", "Block"))),
				mockNSGs("nsg#3", []string{"nic#4"}, withSecRules(secRule("8080", "Allow"))),
				mockNSGs("nsg#4", []string{"nic#6"}, withSecRules(secRule("7812", "Block"))),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := enricher.Enrich(t.Context(), newCycle(1), tc.input)
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Equal(t, tc.output, tc.input)
		})
	}
}

type mockAssetOption func(asset *inventory.AzureAsset)

func withSecRules(rules ...map[string]any) mockAssetOption {
	return func(asset *inventory.AzureAsset) {
		asset.Properties["securityRules"] = rules
	}
}

func withExtendedSecRules(rules ...map[string]any) mockAssetOption {
	return func(asset *inventory.AzureAsset) {
		asset.AddExtension(inventory.ExtensionNetwork, map[string]any{
			"securityRules": rules,
		})
	}
}

func mockVMs(id string, nics []string, opts ...mockAssetOption) inventory.AzureAsset {
	asset := inventory.AzureAsset{
		Id:   id,
		Type: inventory.VirtualMachineAssetType,
		Properties: map[string]any{
			"networkProfile": map[string]any{
				"networkInterfaces": mapNics(nics),
			},
		},
	}

	for _, opt := range opts {
		opt(&asset)
	}

	return asset
}

func mockNSGs(id string, nics []string, opts ...mockAssetOption) inventory.AzureAsset {
	asset := inventory.AzureAsset{
		Id:   id,
		Type: inventory.NetworkSecurityGroupAssetType,
		Properties: map[string]any{
			"networkInterfaces": mapNics(nics),
		},
	}

	for _, opt := range opts {
		opt(&asset)
	}

	return asset
}

func secRule(destinationPortRange, access string) map[string]any {
	return map[string]any{
		"properties": map[string]any{
			"provisioningState":          "Succeeded",
			"destinationAddressPrefixes": make([]string, 0),
			"destinationAddressPrefix":   "*",
			"sourceAddressPrefixes":      make([]string, 0),
			"destinationPortRanges":      make([]string, 0),
			"sourceAddressPrefix":        "*",
			"destinationPortRange":       destinationPortRange,
			"sourcePortRanges":           make([]string, 0),
			"sourcePortRange":            "*",
			"protocol":                   "TCP",
			"direction":                  "Inbound",
			"priority":                   100,
			"access":                     access,
		},
	}
}

func extendedSecRule(destinationPortRange, access string) map[string]any {
	return map[string]any{
		"access":                access,
		"destinationPortRange":  destinationPortRange,
		"destinationPortRanges": make([]string, 0),
		"direction":             "Inbound",
		"protocol":              "TCP",
		"sourceAddressPrefix":   "*",
		"sourceAddressPrefixes": make([]string, 0),
	}
}

func mapNics(nics []string) []map[string]any {
	nicMap := make([]map[string]any, len(nics))
	for i, nic := range nics {
		nicMap[i] = map[string]any{
			"id": nic,
		}
	}
	return nicMap
}
