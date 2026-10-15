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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceArnGovCloudPartition(t *testing.T) {
	const (
		region  = "us-gov-west-1"
		account = "123456789012"
	)
	tests := map[string]struct {
		resource interface{ GetResourceArn() string }
		want     string
	}{
		"instance": {
			resource: Ec2Instance{Instance: types.Instance{InstanceId: aws.String("i-1")}, Region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:ec2/i-1",
		},
		"ebs snapshot": {
			resource: EBSSnapshot{SnapshotId: "snap-1", Region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:ec2/snap-1",
		},
		"vpc": {
			resource: VpcInfo{Vpc: types.Vpc{VpcId: aws.String("vpc-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:vpc/vpc-1",
		},
		"security group": {
			resource: SecurityGroup{SecurityGroup: types.SecurityGroup{GroupId: aws.String("sg-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:security-group/sg-1",
		},
		"nacl": {
			resource: NACLInfo{NetworkAcl: types.NetworkAcl{NetworkAclId: aws.String("acl-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:network-acl/acl-1",
		},
		"network interface": {
			resource: NetworkInterfaceInfo{NetworkInterface: types.NetworkInterface{NetworkInterfaceId: aws.String("eni-1"), OwnerId: aws.String(account)}, region: region},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:network-interface/eni-1",
		},
		"internet gateway": {
			resource: InternetGatewayInfo{InternetGateway: types.InternetGateway{InternetGatewayId: aws.String("igw-1"), OwnerId: aws.String(account)}, region: region},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:internet-gateway/igw-1",
		},
		"nat gateway": {
			resource: NatGatewayInfo{NatGateway: types.NatGateway{NatGatewayId: aws.String("nat-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:natgateway/nat-1",
		},
		"vpc peering connection": {
			resource: VpcPeeringConnectionInfo{VpcPeeringConnection: types.VpcPeeringConnection{VpcPeeringConnectionId: aws.String("pcx-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:vpc-peering-connection/pcx-1",
		},
		"transit gateway attachment": {
			resource: TransitGatewayAttachmentInfo{TransitGatewayAttachment: types.TransitGatewayAttachment{TransitGatewayAttachmentId: aws.String("tgw-attach-1")}, region: region, awsAccount: account},
			want:     "arn:aws-us-gov:ec2:us-gov-west-1:123456789012:transit-gateway-attachment/tgw-attach-1",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.resource.GetResourceArn())
		})
	}
}
