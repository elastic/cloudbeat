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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

var (
	subnetAssociationIdFilterName   = "association.subnet-id"
	subnetVpcIdFilterName           = "vpc-id"
	subnetMainAssociationFilterName = "association.main"
)

type Provider struct {
	log          *logp.Logger
	clients      map[string]Client
	awsAccountID string
}

type Client interface {
	DescribeNetworkAcls(ctx context.Context, params *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeFlowLogs(ctx context.Context, params *ec2.DescribeFlowLogsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeFlowLogsOutput, error)
	GetEbsEncryptionByDefault(ctx context.Context, params *ec2.GetEbsEncryptionByDefaultInput, optFns ...func(*ec2.Options)) (*ec2.GetEbsEncryptionByDefaultOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	CreateSnapshots(ctx context.Context, params *ec2.CreateSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.CreateSnapshotsOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DeleteSnapshot(ctx context.Context, params *ec2.DeleteSnapshotInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSnapshotOutput, error)
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
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
			result = append(result, NACLInfo{
				nacl,
				p.awsAccountID,
				region,
			})
		}
		return result, nil
	})
	return lo.Flatten(nacl), err
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
			result = append(result, SecurityGroup{sg, p.awsAccountID, region})
		}
		return result, nil
	})
	return lo.Flatten(securityGroups), err
}

func (p *Provider) DescribeVPCs(ctx context.Context) ([]awslib.AwsResource, error) {
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

			result = append(result, VpcInfo{
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

func (p *Provider) DescribeInstances(ctx context.Context) ([]*Ec2Instance, error) {
	insances, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]*Ec2Instance, error) {
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
	return lo.Flatten(insances), err
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
					{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("elastic-vulnerability-%s", *ins.InstanceId))},
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

func (p *Provider) DeleteSnapshot(ctx context.Context, snapshot EBSSnapshot) error {
	client, err := awslib.GetClient(aws.String(snapshot.Region), p.clients)
	if err != nil {
		return err
	}
	_, err = client.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{SnapshotId: aws.String(snapshot.SnapshotId)})
	if err != nil {
		return fmt.Errorf("error deleting snapshot %s: %w", snapshot.SnapshotId, err)
	}

	return nil
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
