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

package awsfetcher

import (
	"context"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type fetchmapFunc func(context.Context) ([]awslib.AwsResource, error)
type networkingFetcher struct {
	logger      *logp.Logger
	provider    networkingProvider
	AccountId   string
	AccountName string
}

// TODO(kuba): dynamic classification
var networkingClassification = inventory.AssetClassification{
	Category:    inventory.CategoryInfrastructure,
	SubCategory: inventory.SubCategoryNetwork,
	Type:        inventory.TypeRelationalDatabase,
}

type networkingProvider interface {
	// Internet Gateway
	DescribeNatGateways(context.Context) ([]awslib.AwsResource, error)
	DescribeNetworkAcl(context.Context) ([]awslib.AwsResource, error)
	DescribeNetworkInterfaces(context.Context) ([]awslib.AwsResource, error)
	DescribeSecurityGroups(context.Context) ([]awslib.AwsResource, error)
	DescribeSubnets(context.Context) ([]awslib.AwsResource, error)
	DescribeTransitGatewayAttachments(context.Context) ([]awslib.AwsResource, error)
	DescribeTransitGateways(context.Context) ([]awslib.AwsResource, error)
	DescribeVpcPeeringConnections(context.Context) ([]awslib.AwsResource, error)
	DescribeVpcs(context.Context) ([]awslib.AwsResource, error)
}

func newNetworkingFetcher(logger *logp.Logger, identity *cloud.Identity, provider networkingProvider) inventory.AssetFetcher {
	return &networkingFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *networkingFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	// Type, SubType
	fetchmap := map[string]fetchmapFunc{
		// Internet Gateway
		"NAT Gateways":                s.provider.DescribeNatGateways,               // Virtual Network, NAT Gateway
		"Network ACLs":                s.provider.DescribeNetworkAcl,                // Identity, Authorization, ACL, VPC ACL
		"Network Interfaces":          s.provider.DescribeNetworkInterfaces,         // Interface, EC2 Network Interface
		"Security Groups":             s.provider.DescribeSecurityGroups,            // Firewall, Security Group
		"Subnets":                     s.provider.DescribeSubnets,                   // Subnet, EC2 Subnet
		"Transit Gateways":            s.provider.DescribeTransitGateways,           // Virtual Network, Transit Gateway
		"Transit Gateway Attachments": s.provider.DescribeTransitGatewayAttachments, // Virtual Network, Transit Gateway Attachment
		"VPC Peering Connections":     s.provider.DescribeVpcPeeringConnections,     // Peering, VPC Peering Connection
		"VPCs":                        s.provider.DescribeVpcs,                      // Virtual Network, VPC
	}
	for resourceName, function := range fetchmap {
		s.fetch(ctx, resourceName, function, assetChannel)
	}
}

func (s *networkingFetcher) fetch(ctx context.Context, resourceName string, function fetchmapFunc, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Infof("Fetching %s", resourceName)
	defer s.logger.Infof("Fetching %s - Finished", resourceName)

	awsResources, err := function(ctx)
	if err != nil {
		s.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	classification := networkingClassification
	// TODO
	classification.SubType = "TODO(kuba)"

	for _, item := range awsResources {
		assetChannel <- inventory.NewAssetEvent(
			classification,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   item.GetRegion(),
				Account: inventory.AssetCloudAccount{
					Id:   s.AccountId,
					Name: s.AccountName,
				},
				Service: &inventory.AssetCloudService{
					// TODO(kuba):
					Name: "Networking",
				},
			}),
		)
	}
}
