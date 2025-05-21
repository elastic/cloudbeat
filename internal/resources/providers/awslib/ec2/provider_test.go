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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var onlyDefaultRegion = []string{awslib.DefaultRegion}

func TestProvider_DescribeInternetGateways(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeInternetGateways", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeInternetGateways", mock.Anything, mock.Anything).
					Return(&ec2.DescribeInternetGatewaysOutput{
						InternetGateways: []types.InternetGateway{
							{
								Attachments: []types.InternetGatewayAttachment{
									{State: "available", VpcId: pointers.Ref("vpc-0fda1d140c11370d4")},
								},
								InternetGatewayId: pointers.Ref("igw-0b5dba6f6aee1320c"),
								OwnerId:           pointers.Ref("378890115541"),
								Tags:              []types.Tag{},
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeInternetGateways(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeNatGateways(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNatGateways", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNatGateways", mock.Anything, mock.Anything).
					Return(&ec2.DescribeNatGatewaysOutput{
						NatGateways: []types.NatGateway{
							{
								ConnectivityType: types.ConnectivityTypePrivate,
								CreateTime:       pointers.Ref(time.Now()),
								NatGatewayAddresses: []types.NatGatewayAddress{
									{
										IsPrimary:          pointers.Ref(true),
										NetworkInterfaceId: pointers.Ref("if-abc"),
										PrivateIp:          pointers.Ref("192.168.0.13"),
										Status:             types.NatGatewayAddressStatusSucceeded,
									},
								},
								NatGatewayId: pointers.Ref("ngw-0b5dba6f6aee1320c"),
								State:        types.NatGatewayStateAvailable,
								SubnetId:     pointers.Ref("subnet-0f3223435cc3222"),
								VpcId:        pointers.Ref("vpc-0fda1d140c11370d4"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeNatGateways(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeNetworkAcl(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkAcls", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkAcls", mock.Anything, mock.Anything).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{},
						{},
					},
				}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeNetworkAcl(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything).
					Return(&ec2.DescribeNetworkInterfacesOutput{
						NetworkInterfaces: []types.NetworkInterface{
							{
								Association:        &types.NetworkInterfaceAssociation{},
								Attachment:         &types.NetworkInterfaceAttachment{},
								NetworkInterfaceId: pointers.Ref("if-33289245489578947"),
								SubnetId:           pointers.Ref("subnet-0f3223435cc3222"),
								VpcId:              pointers.Ref("vpc-0fda1d140c11370d4"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeNetworkInterfaces(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeSecurityGroups(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			regions: onlyDefaultRegion,
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{},
						{},
					},
				}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}

			got, err := p.DescribeSecurityGroups(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeSubnets(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSubnets", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSubnets", mock.Anything, mock.Anything).
					Return(&ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{
							{
								AvailabilityZone:        pointers.Ref("eu-north-1b"),
								AvailabilityZoneId:      pointers.Ref("eun1-az2"),
								AvailableIpAddressCount: pointers.Ref(int32(4091)),
								CidrBlock:               pointers.Ref("172.31.32.0/20"),
								DefaultForAz:            pointers.Ref(true),
								MapPublicIpOnLaunch:     pointers.Ref(true),
								State:                   types.SubnetStateAvailable,
								SubnetId:                pointers.Ref("subnet-0a8808bda599a731c"),
								VpcId:                   pointers.Ref("vpc-0fda1d140c11370d4"),
								OwnerId:                 pointers.Ref("378890115541"),
								SubnetArn:               pointers.Ref("arn:aws:ec2:eu-north-1:378890115541:subnet/subnet-0a8808bda599a731c"),
								PrivateDnsNameOptionsOnLaunch: &types.PrivateDnsNameOptionsOnLaunch{
									HostnameType: types.HostnameTypeIpName,
								},
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeSubnets(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeTransitGatewayAttachments(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeTransitGatewayAttachments", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeTransitGatewayAttachments", mock.Anything, mock.Anything).
					Return(&ec2.DescribeTransitGatewayAttachmentsOutput{
						TransitGatewayAttachments: []types.TransitGatewayAttachment{
							{
								CreationTime:               pointers.Ref(time.Now()),
								ResourceId:                 pointers.Ref("tgwa-0a8808bda599a731c"),
								ResourceOwnerId:            pointers.Ref("378890115541"),
								State:                      types.TransitGatewayAttachmentStateAvailable,
								TransitGatewayAttachmentId: pointers.Ref("tgwa-0a8808bda599a731c"),
								TransitGatewayOwnerId:      pointers.Ref("378890115541"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeTransitGatewayAttachments(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeTransitGateways(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeTransitGateways", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeTransitGateways", mock.Anything, mock.Anything).
					Return(&ec2.DescribeTransitGatewaysOutput{
						TransitGateways: []types.TransitGateway{
							{
								CreationTime:      pointers.Ref(time.Now()),
								Options:           &types.TransitGatewayOptions{},
								State:             types.TransitGatewayStateAvailable,
								TransitGatewayArn: pointers.Ref("arn:aws:ec2:eu-north-1:378890115541:transit-gateway/tgw-0a8808bda599a731c"),
								TransitGatewayId:  pointers.Ref("tgw-0a8808bda599a731c"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeTransitGateways(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_DescribeVpcs(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			regions: onlyDefaultRegion,
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{{VpcId: aws.String("vpc-123")}},
				}, nil)

				m.On("DescribeFlowLogs", mock.Anything, mock.Anything).Return(&ec2.DescribeFlowLogsOutput{
					FlowLogs: []types.FlowLog{{FlowLogId: aws.String("fl-123")}},
				}, nil)

				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}

			got, err := p.DescribeVpcs(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_GetEbsEncryptionByDefault(t *testing.T) {
	tests := []struct {
		name    string
		client  func() Client
		want    []awslib.AwsResource
		wantErr bool
		regions []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("GetEbsEncryptionByDefault", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to get ebs"))
				return m
			},
			regions: onlyDefaultRegion,
			wantErr: true,
		},
		{
			name: "with resource",
			client: func() Client {
				m := &MockClient{}
				m.On("GetEbsEncryptionByDefault", mock.Anything, mock.Anything).Return(&ec2.GetEbsEncryptionByDefaultOutput{
					EbsEncryptionByDefault: aws.Bool(true),
				}, nil)
				return m
			},
			want: []awslib.AwsResource{
				&EBSEncryption{
					Enabled:    true,
					region:     awslib.DefaultRegion,
					awsAccount: "aws-account",
				},
			},
			regions: onlyDefaultRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:          testhelper.NewLogger(t),
				clients:      clients,
				awsAccountID: "aws-account",
			}
			got, err := p.GetEbsEncryptionByDefault(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_GetRouteTableForSubnet(t *testing.T) {
	routeTableId := "123456789"
	anotherRouteTableId := "87654321"
	routeTable := types.RouteTable{RouteTableId: &routeTableId}
	anotherRouteTable := types.RouteTable{RouteTableId: &anotherRouteTableId}

	tests := []struct {
		name    string
		client  func() Client
		want    types.RouteTable
		wantErr bool
		regions []string
	}{
		{
			name: "Gets attached route table for subnet",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(&ec2.DescribeRouteTablesOutput{RouteTables: []types.RouteTable{routeTable}}, nil).Once()
				return m
			},
			want:    routeTable,
			wantErr: false,
			regions: onlyDefaultRegion,
		},
		{
			name: "Gets implicitly attached route table for subnet",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(&ec2.DescribeRouteTablesOutput{RouteTables: []types.RouteTable{}}, nil).Once()
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(&ec2.DescribeRouteTablesOutput{RouteTables: []types.RouteTable{anotherRouteTable}}, nil).Once()
				return m
			},
			want:    anotherRouteTable,
			wantErr: false,
			regions: onlyDefaultRegion,
		},
		{
			name: "Errors when fetching attached route table for subnet",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(nil, errors.New("bla")).Once()
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "Errors when fetching implicitly attached route table for subnet",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(&ec2.DescribeRouteTablesOutput{RouteTables: []types.RouteTable{}}, nil).Once()
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(nil, errors.New("bla")).Once()
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "Errors when there is more than 1 attached route table for subnet",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeRouteTables", mock.Anything, mock.Anything).Return(&ec2.DescribeRouteTablesOutput{RouteTables: []types.RouteTable{routeTable, anotherRouteTable}}, nil).Once()
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:          testhelper.NewLogger(t),
				clients:      clients,
				awsAccountID: "aws-account",
			}
			got, err := p.GetRouteTableForSubnet(t.Context(), tt.regions[0], "", "")
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_DescribeVolumes(t *testing.T) {
	expectToken := func(token string) func(input *ec2.DescribeVolumesInput) bool {
		return func(input *ec2.DescribeVolumesInput) bool {
			return *input.NextToken == token
		}
	}

	expectInstances := func(ids ...string) func(input *ec2.DescribeVolumesInput) bool {
		return func(input *ec2.DescribeVolumesInput) bool {
			if len(input.Filters) != 1 {
				return false
			}
			if *input.Filters[0].Name != "attachment.instance-id" {
				return false
			}
			if len(input.Filters[0].Values) != len(ids) {
				return false
			}
			for i, id := range ids {
				if input.Filters[0].Values[i] != id {
					return false
				}
			}
			return true
		}
	}
	mockResult := types.Volume{
		VolumeId:  aws.String("vol-123456789"),
		Encrypted: aws.Bool(true),
		Size:      aws.Int32(8),
		Attachments: []types.VolumeAttachment{
			{
				InstanceId: aws.String("i-123456789"),
				Device:     aws.String("/dev/sda1"),
			},
		},
	}
	expectedVolume := &Volume{
		VolumeId:   "vol-123456789",
		InstanceId: "i-123456789",
		Device:     "/dev/sda1",
		Encrypted:  true,
		Size:       8,
		Region:     awslib.DefaultRegion,
	}

	tests := []struct {
		name      string
		client    func() Client
		instances []*Ec2Instance
		want      []*Volume
		wantErr   bool
		regions   []string
	}{
		{
			name:      "Get 3 volumes from 3 pages",
			instances: []*Ec2Instance{},
			client: func() Client {
				m := &MockClient{}
				m.EXPECT().DescribeVolumes(mock.Anything, mock.MatchedBy(expectInstances())).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}, NextToken: aws.String("1")}, nil).Once()
				m.EXPECT().DescribeVolumes(mock.Anything, mock.MatchedBy(expectToken("1"))).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}, NextToken: aws.String("2")}, nil).Once()
				m.EXPECT().DescribeVolumes(mock.Anything, mock.MatchedBy(expectToken("2"))).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}}, nil).Once()
				return m
			},
			want:    []*Volume{expectedVolume, expectedVolume, expectedVolume},
			wantErr: false,
			regions: onlyDefaultRegion,
		},
		{
			name:      "Get 3 volumes from 1 page",
			instances: []*Ec2Instance{},
			client: func() Client {
				m := &MockClient{}
				m.EXPECT().DescribeVolumes(mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult, mockResult, mockResult}}, nil).Once()
				return m
			},
			want:    []*Volume{expectedVolume, expectedVolume, expectedVolume},
			wantErr: false,
			regions: onlyDefaultRegion,
		},
		{
			name: "Get volumes filtered by instance id from 2 pages",
			instances: []*Ec2Instance{
				{Instance: types.Instance{InstanceId: aws.String("123")}},
				{Instance: types.Instance{InstanceId: aws.String("456")}},
			},
			client: func() Client {
				m := &MockClient{}
				m.EXPECT().DescribeVolumes(mock.Anything, mock.MatchedBy(expectInstances("123", "456"))).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}, NextToken: aws.String("1")}, nil).Once()
				m.EXPECT().DescribeVolumes(mock.Anything, mock.MatchedBy(expectInstances("123", "456"))).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}}, nil).Once()
				return m
			},
			want:    []*Volume{expectedVolume, expectedVolume},
			wantErr: false,
			regions: onlyDefaultRegion,
		},
		{
			name:      "Get error at 3rd page",
			instances: []*Ec2Instance{},
			client: func() Client {
				m := &MockClient{}
				m.EXPECT().DescribeVolumes(mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}, NextToken: aws.String("1")}, nil).Once()
				m.EXPECT().DescribeVolumes(mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{Volumes: []types.Volume{mockResult}, NextToken: aws.String("2")}, nil).Once()
				m.EXPECT().DescribeVolumes(mock.Anything, mock.Anything).Return(nil, errors.New("bla")).Once()
				return m
			},
			want:    []*Volume{},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:          testhelper.NewLogger(t),
				clients:      clients,
				awsAccountID: "aws-account",
			}
			got, err := p.DescribeVolumes(t.Context(), tt.instances)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_DescribeVpcPeeringConnections(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcPeeringConnections", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcPeeringConnections", mock.Anything, mock.Anything).
					Return(&ec2.DescribeVpcPeeringConnectionsOutput{
						VpcPeeringConnections: []types.VpcPeeringConnection{
							{
								AccepterVpcInfo:        &types.VpcPeeringConnectionVpcInfo{},
								ExpirationTime:         pointers.Ref(time.Now()),
								RequesterVpcInfo:       &types.VpcPeeringConnectionVpcInfo{},
								Status:                 &types.VpcPeeringConnectionStateReason{},
								VpcPeeringConnectionId: pointers.Ref("abc123"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.DescribeVpcPeeringConnections(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}
