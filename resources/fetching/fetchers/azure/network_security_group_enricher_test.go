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

	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestEnrichVirtualMachinesWithNetworkSecurityGroup(t *testing.T) {
	// Todo
	// 	- test with multiple nsg to multiple vms
	// 	- test with one nsg to multiple vms
	//  - test with multiple vms to one nsg
	// 	- test with legacy vms

	tests := map[string]struct {
		input  []inventory.AzureAsset
		output []inventory.AzureAsset
	}{
		"1 Security Network Group with default security rules maps to 1 Virtual Machine with same Network Interface": {
			input: []inventory.AzureAsset{
				{Id: "vm#1", Type: inventory.VirtualMachineAssetType, Properties: map[string]any{
					"networkProfile": map[string]any{
						"networkInterfaces": []map[string]any{
							{
								"id": "nic#1",
							},
						},
					},
				}},
				{Id: "nsg#1", Type: inventory.NetworkSecurityGroup, Properties: map[string]any{
					"networkInterfaces": []map[string]any{
						{
							"id": "nic#1",
						},
					},
					"securityRules": []map[string]any{
						{
							"properties": map[string]any{
								"provisioningState":          "Succeeded",
								"destinationAddressPrefixes": make([]string, 0),
								"destinationAddressPrefix":   "*",
								"sourceAddressPrefixes":      make([]string, 0),
								"destinationPortRanges":      make([]string, 0),
								"sourceAddressPrefix":        "*",
								"destinationPortRange":       "22",
								"sourcePortRanges":           make([]string, 0),
								"sourcePortRange":            "*",
								"protocol":                   "TCP",
								"direction":                  "Inbound",
								"priority":                   100,
								"access":                     "Allow",
							},
						},
					},
				}},
			},
			output: []inventory.AzureAsset{
				{Id: "vm#1", Type: inventory.VirtualMachineAssetType, Properties: map[string]any{
					"networkProfile": map[string]any{
						"networkInterfaces": []map[string]any{
							{
								"id": "nic#1",
							},
						},
					},
				},
					Extension: map[string]any{
						"network": map[string]any{
							"securityRules": []map[string]any{
								{
									"access":                "Allow",
									"destinationPortRange":  "22",
									"destinationPortRanges": make([]string, 0),
									"direction":             "Inbound",
									"protocol":              "TCP",
									"sourceAddressPrefix":   "*",
									"sourceAddressPrefixes": make([]string, 0),
								},
							},
						},
					},
				},
				{Id: "nsg#1", Type: inventory.NetworkSecurityGroup, Properties: map[string]any{
					"networkInterfaces": []map[string]any{
						{
							"id": "nic#1",
						},
					},
					"securityRules": []map[string]any{
						{
							"properties": map[string]any{
								"provisioningState":          "Succeeded",
								"destinationAddressPrefixes": make([]string, 0),
								"destinationAddressPrefix":   "*",
								"sourceAddressPrefixes":      make([]string, 0),
								"destinationPortRanges":      make([]string, 0),
								"sourceAddressPrefix":        "*",
								"destinationPortRange":       "22",
								"sourcePortRanges":           make([]string, 0),
								"sourcePortRange":            "*",
								"protocol":                   "TCP",
								"direction":                  "Inbound",
								"priority":                   100,
								"access":                     "Allow",
							},
						},
					},
				}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			enrichVirtualMachinesWithNetworkSecurityGroups(tc.input)
			assert.Equal(t, tc.output, tc.input)
		})
	}
}
