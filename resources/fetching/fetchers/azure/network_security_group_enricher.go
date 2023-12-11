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
	"errors"
	"github.com/mitchellh/mapstructure"

	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type networkProfile struct {
	NetworkInterfaces []networkInterface `mapstructure:"networkInterfaces"`
}

type networkInterface struct {
	Id string `mapstructure:"id"`
}

type extensionNetworkSecurityRules struct {
	Access                string   `mapstructure:"access"`
	DestinationPortRange  string   `mapstructure:"destinationPortRange"`
	DestinationPortRanges []string `mapstructure:"destinationPortRanges"`
	Direction             string   `mapstructure:"direction"`
	Protocol              string   `mapstructure:"protocol"`
	SourceAddressPrefix   string   `mapstructure:"sourceAddressPrefix"`
	SourceAddressPrefixes []string `mapstructure:"sourceAddressPrefixes"`
}

type networkSecurityRules struct {
	Properties extensionNetworkSecurityRules `mapstructure:"properties"`
}

func enrichVirtualMachinesWithNetworkSecurityGroups(assets []inventory.AzureAsset) error {
	// VMs are connected to NSGs via Network Interfaces (https://learn.microsoft.com/en-us/azure/virtual-network/network-overview)
	// Therefore, aggregate security rules by Network Interface to easy access later on
	securityRulesByNetworkInterface := make(map[string][]extensionNetworkSecurityRules)
	var errAgg error

	for _, asset := range assets {
		var err error
		securityRulesByNetworkInterface, err = extractNetworkSecurityGroupRules(asset, securityRulesByNetworkInterface)
		if err != nil {
			errAgg = errors.Join(errAgg, err)
		}
	}

	for idx, asset := range assets {
		if asset.Type != inventory.VirtualMachineAssetType && asset.Type != inventory.ClassicVirtualMachineAssetType {
			continue
		}

		if err := addNetworkRulesToAssetExtensions(&asset, securityRulesByNetworkInterface); err != nil {
			errAgg = errors.Join(errAgg, err)
		}

		assets[idx] = asset
	}

	return errAgg
}

func extractNetworkSecurityGroupRules(asset inventory.AzureAsset, securityRulesByNetworkInterface map[string][]extensionNetworkSecurityRules) (map[string][]extensionNetworkSecurityRules, error) {
	if asset.Type != inventory.NetworkSecurityGroup {
		return securityRulesByNetworkInterface, nil
	}

	var nics []networkInterface
	if err := mapstructure.Decode(asset.Properties["networkInterfaces"], &nics); err != nil {
		return securityRulesByNetworkInterface, err
	}

	var rawSecurityRules []networkSecurityRules
	if err := mapstructure.Decode(asset.Properties["securityRules"], &rawSecurityRules); err != nil {
		return securityRulesByNetworkInterface, err
	}

	securityRules := make([]extensionNetworkSecurityRules, 0, len(rawSecurityRules))
	for _, rule := range rawSecurityRules {
		securityRules = append(securityRules, rule.Properties)
	}

	for _, nic := range nics {
		interfaceRules, exists := securityRulesByNetworkInterface[nic.Id]
		if !exists {
			interfaceRules = make([]extensionNetworkSecurityRules, 0, len(securityRules))
		}

		securityRulesByNetworkInterface[nic.Id] = append(interfaceRules, securityRules...)
	}

	return securityRulesByNetworkInterface, nil
}

func addNetworkRulesToAssetExtensions(vm *inventory.AzureAsset, securityRuleByNetworkInterface map[string][]extensionNetworkSecurityRules) error {
	var profile networkProfile
	if err := mapstructure.Decode(vm.Properties["networkProfile"], &profile); err != nil {
		return err
	}

	rules := make([]map[string]any, 0)
	for _, nic := range profile.NetworkInterfaces {
		for _, rule := range securityRuleByNetworkInterface[nic.Id] {
			rules = append(rules, map[string]any{
				"access":                rule.Access,
				"destinationPortRange":  rule.DestinationPortRange,
				"destinationPortRanges": rule.DestinationPortRanges,
				"direction":             rule.Direction,
				"protocol":              rule.Protocol,
				"sourceAddressPrefix":   rule.SourceAddressPrefix,
				"sourceAddressPrefixes": rule.SourceAddressPrefixes,
			})
		}
	}

	if len(rules) == 0 {
		return nil
	}

	if vm.Extension == nil {
		vm.Extension = make(map[string]any)
	}

	vm.Extension["network"] = map[string]any{
		"securityRules": rules,
	}

	return nil
}
