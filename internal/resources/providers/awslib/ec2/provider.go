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

package ec2

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

var (
	subnetAssociationIdFilterName   = "association.subnet-id"
	subnetVpcIdFilterName           = "vpc-id"
	subnetMainAssociationFilterName = "association.main"
)

const (
	snapshotPrefix = "elastic-vulnerability"
)

type Provider struct {
	log          *clog.Logger
	clients      map[string]Client
	awsAccountID string
}

func NewProviderFromClients(log *clog.Logger, awsAccountID string, clients map[string]Client) *Provider {
	return &Provider{
		log:          log,
		clients:      clients,
		awsAccountID: awsAccountID,
	}
}

type Client interface {
	CreateSnapshots(ctx context.Context, params *ec2.CreateSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotsOutput, error)
	DeleteSnapshot(ctx context.Context, params *ec2.DeleteSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSnapshotOutput, error)
	GetEbsEncryptionByDefault(ctx context.Context, params *ec2.GetEbsEncryptionByDefaultInput, optFns ...func(*ec2.Options)) (*ec2.GetEbsEncryptionByDefaultOutput, error)
	ec2.DescribeFlowLogsAPIClient
	ec2.DescribeInstancesAPIClient
	ec2.DescribeInternetGatewaysAPIClient
	ec2.DescribeNatGatewaysAPIClient
	ec2.DescribeNetworkAclsAPIClient
	ec2.DescribeNetworkInterfacesAPIClient
	ec2.DescribeRouteTablesAPIClient
	ec2.DescribeSecurityGroupsAPIClient
	ec2.DescribeSnapshotsAPIClient
	ec2.DescribeSubnetsAPIClient
	ec2.DescribeTransitGatewayAttachmentsAPIClient
	ec2.DescribeTransitGatewaysAPIClient
	ec2.DescribeVolumesAPIClient
	ec2.DescribeVpcsAPIClient
	ec2.DescribeVpcPeeringConnectionsAPIClient
}

func (p *Provider) CreateSnapshots(ctx context.Context, ins *Ec2Instance) ([]EBSSnapshot, error) {
	client := p.clients[ins.Region]
	if client == nil {
		return nil, fmt.Errorf("error in CreateSnapshots no client for region %s", ins.Region)
	}
	input := &ec2.CreateSnapshotsInput{
		InstanceSpecification: &types.InstanceSpecification{
			InstanceId: ins.InstanceId,
		},
		Description: aws.String("Cloudbeat Vulnerability Snapshot."),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: "snapshot",
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("%s-%s", snapshotPrefix, *ins.InstanceId))},
					{Key: aws.String("Workload"), Value: aws.String("Cloudbeat Vulnerability Snapshot")},
				},
			},
		},
	}
	res, err := client.CreateSnapshots(ctx, input)
	if err != nil {
		return nil, err
	}

	result := make([]EBSSnapshot, 0, len(res.Snapshots))
	for _, snap := range res.Snapshots {
		result = append(result, FromSnapshotInfo(snap, ins.Region, p.awsAccountID, *ins))
	}
	return result, nil
}

func (p *Provider) DeleteSnapshot(ctx context.Context, snapshot EBSSnapshot) error {
	client, err := awslib.GetClient(aws.String(snapshot.Region), p.clients)
	if err != nil {
		return err
	}
	_, err = client.DeleteSnapshot(ctx,
		&ec2.DeleteSnapshotInput{SnapshotId: aws.String(snapshot.SnapshotId)},
		func(ec2Options *ec2.Options) {
			ec2Options.Retryer = retry.NewStandard(
				awslib.RetryableCodesOption,
				func(retryOptions *retry.StandardOptions) {
					retryOptions.MaxAttempts = 10
				},
			)
		},
	)
	if err != nil {
		return fmt.Errorf("error deleting snapshot %s: %w", snapshot.SnapshotId, err)
	}

	return nil
}

func (p *Provider) DescribeInstances(ctx context.Context) ([]*Ec2Instance, error) {
	instances, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]*Ec2Instance, error) {
		input := &ec2.DescribeInstancesInput{}
		allInstances := []types.Instance{}
		for {
			output, err := c.DescribeInstances(ctx, input)
			if err != nil {
				return nil, err
			}
			for _, reservation := range output.Reservations {
				allInstances = append(allInstances, reservation.Instances...)
			}
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []*Ec2Instance
		for _, instance := range allInstances {
			result = append(result, &Ec2Instance{
				Instance:   instance,
				awsAccount: p.awsAccountID,
				Region:     region,
			})
		}
		return result, nil
	})
	return lo.Flatten(instances), err
}

func (p *Provider) DescribeInternetGateways(ctx context.Context) ([]awslib.AwsResource, error) {
	gateways, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeInternetGatewaysInput{}
		all := []types.InternetGateway{}
		for {
			output, err := c.DescribeInternetGateways(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.InternetGateways...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &InternetGatewayInfo{
				InternetGateway: item,
				region:          region,
			})
		}
		return result, nil
	})
	return lo.Flatten(gateways), err
}

func (p *Provider) DescribeNatGateways(ctx context.Context) ([]awslib.AwsResource, error) {
	gateways, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeNatGatewaysInput{}
		all := []types.NatGateway{}
		for {
			output, err := c.DescribeNatGateways(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.NatGateways...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &NatGatewayInfo{
				NatGateway: item,
				awsAccount: p.awsAccountID,
				region:     region,
			})
		}
		return result, nil
	})
	return lo.Flatten(gateways), err
}

func (p *Provider) DescribeNetworkAcl(ctx context.Context) ([]awslib.AwsResource, error) {
	nacl, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var allAcls []types.NetworkAcl
		input := ec2.DescribeNetworkAclsInput{}
		for {
			output, err := c.DescribeNetworkAcls(ctx, &input)
			if err != nil {
				return nil, err
			}
			allAcls = append(allAcls, output.NetworkAcls...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, nacl := range allAcls {
			result = append(result, &NACLInfo{
				nacl,
				p.awsAccountID,
				region,
			})
		}
		return result, nil
	})
	return lo.Flatten(nacl), err
}

func (p *Provider) DescribeNetworkInterfaces(ctx context.Context) ([]awslib.AwsResource, error) {
	interfaces, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeNetworkInterfacesInput{}
		all := []types.NetworkInterface{}
		for {
			output, err := c.DescribeNetworkInterfaces(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.NetworkInterfaces...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &NetworkInterfaceInfo{
				NetworkInterface: item,
				region:           region,
			})
		}
		return result, nil
	})
	return lo.Flatten(interfaces), err
}

func (p *Provider) DescribeSecurityGroups(ctx context.Context) ([]awslib.AwsResource, error) {
	securityGroups, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var all []types.SecurityGroup
		input := &ec2.DescribeSecurityGroupsInput{}
		for {
			output, err := c.DescribeSecurityGroups(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.SecurityGroups...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, sg := range all {
			result = append(result, &SecurityGroup{sg, p.awsAccountID, region})
		}
		return result, nil
	})
	return lo.Flatten(securityGroups), err
}

// TODO: Maybe we should bulk request snapshots?
// This will limit us scaling the pipeline
func (p *Provider) DescribeSnapshots(ctx context.Context, snapshot EBSSnapshot) ([]EBSSnapshot, error) {
	client := p.clients[snapshot.Region]
	if client == nil {
		return nil, fmt.Errorf("error in DescribeSnapshots no client for region %s", snapshot.Region)
	}
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []string{snapshot.SnapshotId},
	}
	res, err := client.DescribeSnapshots(ctx, input)
	if err != nil {
		return nil, err
	}

	result := make([]EBSSnapshot, 0, len(res.Snapshots))
	for _, snap := range res.Snapshots {
		result = append(result, FromSnapshot(snap, snapshot.Region, p.awsAccountID, snapshot.Instance))
	}
	return result, nil
}

// IterOwnedSnapshots will iterate over the snapshots owned by cloudbeat (snapshotPrefix) that are older than the
// specified before time. A snapshot will be yielded if:
// - It has a tag with key "Name" and value starting with snapshotPrefix
// - It is older than the specified before time
// - It is "owned" by the current account (owner ID is "self")
func (p *Provider) IterOwnedSnapshots(ctx context.Context, before time.Time) iter.Seq[EBSSnapshot] {
	return func(yield func(EBSSnapshot) bool) {
		_, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
			input := &ec2.DescribeSnapshotsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("tag:Name"),
						Values: []string{fmt.Sprintf("%s-*", snapshotPrefix)},
					},
				},
				OwnerIds: []string{"self"},
			}
			paginator := ec2.NewDescribeSnapshotsPaginator(c, input)
			for paginator.HasMorePages() {
				output, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}
				for _, snap := range output.Snapshots {
					if filterSnap(snap, before) {
						p.log.Infof("Found old snapshot %s", *snap.SnapshotId)
						ebsSnap := FromSnapshot(snap, region, p.awsAccountID, Ec2Instance{})
						if !yield(ebsSnap) {
							return nil, nil
						}
					}
				}
			}
			return nil, nil
		})
		if err != nil {
			p.log.Errorf("Error listing owned snapshots: %v", err)
		}
	}
}

func filterSnap(snap types.Snapshot, before time.Time) bool {
	if aws.ToTime(snap.StartTime).After(before) {
		return false
	}

	for _, tag := range snap.Tags {
		if aws.ToString(tag.Key) == "Name" {
			return strings.HasPrefix(aws.ToString(tag.Value), snapshotPrefix)
		}
	}
	return false
}

func (p *Provider) DescribeSubnets(ctx context.Context) ([]awslib.AwsResource, error) {
	subnets, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeSubnetsInput{}
		all := []types.Subnet{}
		for {
			output, err := c.DescribeSubnets(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.Subnets...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &SubnetInfo{
				Subnet: item,
				region: region,
			})
		}
		return result, nil
	})
	return lo.Flatten(subnets), err
}

func (p *Provider) DescribeTransitGatewayAttachments(ctx context.Context) ([]awslib.AwsResource, error) {
	attachments, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeTransitGatewayAttachmentsInput{}
		all := []types.TransitGatewayAttachment{}
		for {
			output, err := c.DescribeTransitGatewayAttachments(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.TransitGatewayAttachments...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &TransitGatewayAttachmentInfo{
				TransitGatewayAttachment: item,
				awsAccount:               p.awsAccountID,
				region:                   region,
			})
		}
		return result, nil
	})
	return lo.Flatten(attachments), err
}

func (p *Provider) DescribeTransitGateways(ctx context.Context) ([]awslib.AwsResource, error) {
	gateways, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &ec2.DescribeTransitGatewaysInput{}
		all := []types.TransitGateway{}
		for {
			output, err := c.DescribeTransitGateways(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.TransitGateways...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, item := range all {
			result = append(result, &TransitGatewayInfo{
				TransitGateway: item,
				region:         region,
			})
		}
		return result, nil
	})
	return lo.Flatten(gateways), err
}

func (p *Provider) DescribeVolumes(ctx context.Context, instances []*Ec2Instance) ([]*Volume, error) {
	instanceFilter := lo.Map(instances, func(ins *Ec2Instance, _ int) string { return *ins.InstanceId })
	volumes, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]*Volume, error) {
		input := &ec2.DescribeVolumesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("attachment.instance-id"),
					Values: instanceFilter,
				},
			},
		}
		allVolumes := []types.Volume{}
		for {
			output, err := c.DescribeVolumes(ctx, input)
			if err != nil {
				return nil, err
			}
			allVolumes = append(allVolumes, output.Volumes...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []*Volume
		for _, vol := range allVolumes {
			if len(vol.Attachments) != 1 {
				p.log.Errorf("Volume %s has %d attachments", *vol.VolumeId, len(vol.Attachments))
				continue
			}

			result = append(result, &Volume{
				VolumeId:   *vol.VolumeId,
				Size:       int(*vol.Size),
				Region:     region,
				Encrypted:  *vol.Encrypted,
				InstanceId: *vol.Attachments[0].InstanceId,
				Device:     *vol.Attachments[0].Device,
			})
		}
		return result, nil
	})
	return lo.Flatten(volumes), err
}

func (p *Provider) DescribeVpcPeeringConnections(ctx context.Context) ([]awslib.AwsResource, error) {
	peerings, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var all []types.VpcPeeringConnection
		input := &ec2.DescribeVpcPeeringConnectionsInput{}
		for {
			output, err := c.DescribeVpcPeeringConnections(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.VpcPeeringConnections...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, peering := range all {
			result = append(result, &VpcPeeringConnectionInfo{
				VpcPeeringConnection: peering,
				awsAccount:           p.awsAccountID,
				region:               region,
			})
		}
		return result, nil
	})
	return lo.Flatten(peerings), err
}

func (p *Provider) DescribeVpcs(ctx context.Context) ([]awslib.AwsResource, error) {
	vpcs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var all []types.Vpc
		input := &ec2.DescribeVpcsInput{}
		for {
			output, err := c.DescribeVpcs(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.Vpcs...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		var result []awslib.AwsResource
		for _, vpc := range all {
			logs, err := c.DescribeFlowLogs(ctx, &ec2.DescribeFlowLogsInput{Filter: []types.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: []string{*vpc.VpcId},
				},
			}})
			if err != nil {
				p.log.Errorf("Error fetching flow logs for VPC %s: %v", *vpc.VpcId, err.Error())
				continue
			}

			result = append(result, &VpcInfo{
				Vpc:        vpc,
				FlowLogs:   logs.FlowLogs,
				awsAccount: p.awsAccountID,
				region:     region,
			})
		}
		return result, nil
	})
	return lo.Flatten(vpcs), err
}

func (p *Provider) GetEbsEncryptionByDefault(ctx context.Context) ([]awslib.AwsResource, error) {
	return awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) (awslib.AwsResource, error) {
		res, err := c.GetEbsEncryptionByDefault(ctx, &ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return nil, err
		}
		return &EBSEncryption{
			Enabled:    *res.EbsEncryptionByDefault,
			region:     region,
			awsAccount: p.awsAccountID,
		}, nil
	})
}

func (p *Provider) GetRouteTableForSubnet(ctx context.Context, region string, subnetId string, vpcId string) (types.RouteTable, error) {
	client, err := awslib.GetClient(&region, p.clients)
	if err != nil {
		return types.RouteTable{}, err
	}

	// Fetching route tables explicitly attached to the subnet
	routeTables, err := client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{Name: &subnetAssociationIdFilterName, Values: []string{subnetId}},
		},
	})
	if err != nil {
		return types.RouteTable{}, err
	}

	// If there are no route tables explicitly attached to the subnet, it means the VPC main subnet is implicitly attached
	if len(routeTables.RouteTables) == 0 {
		routeTables, err = client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{Filters: []types.Filter{
			{Name: &subnetMainAssociationFilterName, Values: []string{"true"}},
			{Name: &subnetVpcIdFilterName, Values: []string{vpcId}},
		}})

		if err != nil {
			return types.RouteTable{}, err
		}
	}

	// A subnet should not have more than 1 attached route table
	if len(routeTables.RouteTables) != 1 {
		return types.RouteTable{}, fmt.Errorf("subnet %s has %d route tables", subnetId, len(routeTables.RouteTables))
	}

	return routeTables.RouteTables[0], nil
}
