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

	"github.com/go-viper/mapstructure/v2"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type vmNetworkSecurityGroupEnricher struct{}

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

func (e vmNetworkSecurityGroupEnricher) Enrich(_ context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	// VMs are connected to NSGs via Network Interfaces (https://learn.microsoft.com/en-us/azure/virtual-network/network-overview)
	// Therefore, aggregate security rules by Network Interface to easy access later on
	securityRulesByNetworkInterface := make(map[string][]extensionNetworkSecurityRules)
	var errAgg error

	for _, asset := range assets {
		if asset.Type != inventory.NetworkSecurityGroupAssetType {
			continue
		}

		if err := e.extractNetworkSecurityGroupRules(asset, securityRulesByNetworkInterface); err != nil {
			errAgg = errors.Join(errAgg, err)
		}
	}

	for idx, asset := range assets {
		if asset.Type != inventory.VirtualMachineAssetType {
			continue
		}

		if err := e.addNetworkRulesToAssetExtensions(&asset, securityRulesByNetworkInterface); err != nil {
			errAgg = errors.Join(errAgg, err)
		}

		assets[idx] = asset
	}

	return errAgg
}

func (e vmNetworkSecurityGroupEnricher) extractNetworkSecurityGroupRules(asset inventory.AzureAsset, securityRulesByNetworkInterface map[string][]extensionNetworkSecurityRules) error {
	var nics []networkInterface
	if err := mapstructure.Decode(asset.Properties["networkInterfaces"], &nics); err != nil {
		return err
	}

	var rawSecurityRules []networkSecurityRules
	if err := mapstructure.Decode(asset.Properties["securityRules"], &rawSecurityRules); err != nil {
		return err
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

	return nil
}

func (e vmNetworkSecurityGroupEnricher) addNetworkRulesToAssetExtensions(vm *inventory.AzureAsset, securityRuleByNetworkInterface map[string][]extensionNetworkSecurityRules) error {
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

	vm.AddExtension(inventory.ExtensionNetwork, map[string]any{
		"securityRules": rules,
	})

	return nil
}
