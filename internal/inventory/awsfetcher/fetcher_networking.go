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

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type networkingFetcher struct {
	logger      *clog.Logger
	provider    networkingProvider
	AccountId   string
	AccountName string
}

type (
	networkDescribeFunc func(context.Context) ([]awslib.AwsResource, error)
	networkingProvider  interface {
		DescribeInternetGateways(context.Context) ([]awslib.AwsResource, error)
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
)

func newNetworkingFetcher(logger *clog.Logger, identity *cloud.Identity, provider networkingProvider) inventory.AssetFetcher {
	return &networkingFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *networkingFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       networkDescribeFunc
		classification inventory.AssetClassification
	}{
		{"Internet Gateways", s.provider.DescribeInternetGateways, inventory.AssetClassificationAwsInternetGateway},
		{"NAT Gateways", s.provider.DescribeNatGateways, inventory.AssetClassificationAwsNatGateway},
		{"Network ACLs", s.provider.DescribeNetworkAcl, inventory.AssetClassificationAwsNetworkAcl},
		{"Network Interfaces", s.provider.DescribeNetworkInterfaces, inventory.AssetClassificationAwsNetworkInterface},
		{"Security Groups", s.provider.DescribeSecurityGroups, inventory.AssetClassificationAwsSecurityGroup},
		{"Subnets", s.provider.DescribeSubnets, inventory.AssetClassificationAwsSubnet},
		{"Transit Gateways", s.provider.DescribeTransitGateways, inventory.AssetClassificationAwsTransitGateway},
		{"Transit Gateway Attachments", s.provider.DescribeTransitGatewayAttachments, inventory.AssetClassificationAwsTransitGatewayAttachment},
		{"VPC Peering Connections", s.provider.DescribeVpcPeeringConnections, inventory.AssetClassificationAwsVpcPeeringConnection},
		{"VPCs", s.provider.DescribeVpcs, inventory.AssetClassificationAwsVpc},
	}
	for _, r := range resourcesToFetch {
		s.fetch(ctx, r.name, r.function, r.classification, assetChannel)
	}
}

func (s *networkingFetcher) fetch(ctx context.Context, resourceName string, function networkDescribeFunc, classification inventory.AssetClassification, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Infof("Fetching %s", resourceName)
	defer s.logger.Infof("Fetching %s - Finished", resourceName)

	awsResources, err := function(ctx)
	if err != nil {
		s.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range awsResources {
		assetChannel <- inventory.NewAssetEvent(
			classification,
			[]string{item.GetResourceArn(), pointers.Deref(s.retrieveId(item))},
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
					Name: "AWS Networking",
				},
			}),
			s.networkEnricher(item),
		)
	}
}

//nolint:revive
func (s *networkingFetcher) networkEnricher(item awslib.AwsResource) inventory.AssetEnricher {
	var enricher inventory.AssetEnricher

	switch obj := item.(type) {
	case *ec2.InternetGatewayInfo:
		vpcIds := []string{}
		for _, attachment := range obj.InternetGateway.Attachments {
			id := pointers.Deref(attachment.VpcId)
			if id != "" {
				vpcIds = append(vpcIds, id)
			}
		}
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			VpcIds: vpcIds,
		})
	case *ec2.NatGatewayInfo:
		ifaceIds := []string{}
		for _, iface := range obj.NatGateway.NatGatewayAddresses {
			id := pointers.Deref(iface.NetworkInterfaceId)
			if id != "" {
				ifaceIds = append(ifaceIds, id)
			}
		}
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			NetworkInterfaceIds: ifaceIds,
			SubnetIds:           []string{pointers.Deref(obj.NatGateway.SubnetId)},
			VpcIds:              []string{pointers.Deref(obj.NatGateway.VpcId)},
		})
	case *ec2.NACLInfo:
		subnetIds := []string{}
		for _, association := range obj.NetworkAcl.Associations {
			id := pointers.Deref(association.SubnetId)
			if id != "" {
				subnetIds = append(subnetIds, id)
			}
		}
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			SubnetIds: subnetIds,
			VpcIds:    []string{pointers.Deref(obj.NetworkAcl.VpcId)},
		})
	case *ec2.NetworkInterfaceInfo:
		secGroupIds := []string{}
		for _, secGroup := range obj.NetworkInterface.Groups {
			id := pointers.Deref(secGroup.GroupId)
			secGroupIds = append(secGroupIds, id)
		}
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			SecurityGroupIds: secGroupIds,
			SubnetIds:        []string{pointers.Deref(obj.NetworkInterface.SubnetId)},
			VpcIds:           []string{pointers.Deref(obj.NetworkInterface.VpcId)},
		})
	case *ec2.SecurityGroup:
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			VpcIds: []string{pointers.Deref(obj.VpcId)},
		})
	case *ec2.SubnetInfo:
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			VpcIds: []string{pointers.Deref(obj.Subnet.VpcId)},
		})
	case *ec2.TransitGatewayAttachmentInfo:
		routeTableId := ""
		if obj.TransitGatewayAttachment.Association != nil {
			routeTableId = pointers.Deref(obj.TransitGatewayAttachment.Association.TransitGatewayRouteTableId)
		}
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			RouteTableIds:     []string{routeTableId},
			TransitGatewayIds: []string{pointers.Deref(obj.TransitGatewayAttachment.TransitGatewayId)},
		})
	case *ec2.VpcPeeringConnectionInfo:
		enricher = inventory.WithNetwork(inventory.AssetNetwork{
			VpcIds: []string{
				pointers.Deref(obj.VpcPeeringConnection.AccepterVpcInfo.VpcId),
				pointers.Deref(obj.VpcPeeringConnection.RequesterVpcInfo.VpcId),
			},
		})
	case *ec2.VpcInfo:
		enricher = inventory.EmptyEnricher()
	default:
		s.logger.Warnf("Unsupported Networking Fetcher type %T (enricher)", obj)
		enricher = inventory.EmptyEnricher()
	}

	return enricher
}

func (s *networkingFetcher) retrieveId(awsResource awslib.AwsResource) *string {
	switch resource := awsResource.(type) {
	case *ec2.InternetGatewayInfo:
		return resource.InternetGateway.InternetGatewayId
	case *ec2.NatGatewayInfo:
		return resource.NatGateway.NatGatewayId
	case *ec2.NACLInfo:
		return resource.NetworkAclId
	case *ec2.NetworkInterfaceInfo:
		return resource.NetworkInterface.NetworkInterfaceId
	case *ec2.SecurityGroup:
		return resource.GroupId
	case *ec2.SubnetInfo:
		return resource.Subnet.SubnetId
	case *ec2.TransitGatewayAttachmentInfo:
		return resource.TransitGatewayAttachment.TransitGatewayAttachmentId
	case *ec2.TransitGatewayInfo:
		return resource.TransitGateway.TransitGatewayId
	case *ec2.VpcPeeringConnectionInfo:
		return resource.VpcPeeringConnection.VpcPeeringConnectionId
	case *ec2.VpcInfo:
		return resource.Vpc.VpcId
	default:
		s.logger.Warnf("Unsupported Networking Fetcher type %T (id)", resource)
		return nil
	}
}
