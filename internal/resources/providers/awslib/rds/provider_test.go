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

package rds

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rdsClient "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
}

type (
	rdsClientMockReturnVals       map[string]map[string][]any
	ec2GetRouteTableForSubnetVals [][]any
)

var (
	identifier           = "identifier"
	identifier2          = "identifier2"
	arn                  = "arn"
	arn2                 = "arn2"
	destinationCidrBlock = "0.0.0.0/0"
	gatewayId            = "igw=12345678"
)

func TestProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)
	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {}

func (s *ProviderTestSuite) TearDownTest() {}

func (s *ProviderTestSuite) TestProvider_DescribeDBInstances() {
	tests := []struct {
		name                          string
		rdsClientMockReturnVals       rdsClientMockReturnVals
		ec2GetRouteTableForSubnetVals ec2GetRouteTableForSubnetVals
		expected                      []awslib.AwsResource
	}{
		{
			name: "Should not return any DB instances when there aren't any",
			rdsClientMockReturnVals: rdsClientMockReturnVals{
				"DescribeDBInstances": {
					awslib.DefaultRegion: {&rdsClient.DescribeDBInstancesOutput{DBInstances: []types.DBInstance{}}, nil},
				},
			},
			expected: []awslib.AwsResource{},
		},
		{
			name: "Should return DB instances",
			rdsClientMockReturnVals: rdsClientMockReturnVals{
				"DescribeDBInstances": {
					awslib.DefaultRegion: {&rdsClient.DescribeDBInstancesOutput{
						DBInstances: []types.DBInstance{
							{
								DBInstanceIdentifier:    &identifier,
								DBInstanceArn:           &arn,
								StorageEncrypted:        aws.Bool(false),
								AutoMinorVersionUpgrade: aws.Bool(false),
								PubliclyAccessible:      aws.Bool(false),
								DBSubnetGroup:           &types.DBSubnetGroup{VpcId: &identifier, Subnets: []types.Subnet{}},
							},
							{
								DBInstanceIdentifier:    &identifier2,
								DBInstanceArn:           &arn2,
								StorageEncrypted:        aws.Bool(true),
								AutoMinorVersionUpgrade: aws.Bool(true),
								PubliclyAccessible:      aws.Bool(true),
								DBSubnetGroup: &types.DBSubnetGroup{VpcId: &identifier, Subnets: []types.Subnet{
									{SubnetIdentifier: &identifier},
									{SubnetIdentifier: &identifier2},
								}},
							},
						},
					}, nil},
				},
			},
			ec2GetRouteTableForSubnetVals: ec2GetRouteTableForSubnetVals{
				{ec2types.RouteTable{}, errors.New("asd")},
				{ec2types.RouteTable{RouteTableId: &identifier, Routes: []ec2types.Route{{DestinationCidrBlock: &destinationCidrBlock, GatewayId: &gatewayId}}}, nil},
			},
			expected: []awslib.AwsResource{
				DBInstance{
					Identifier:              identifier,
					Arn:                     arn,
					StorageEncrypted:        false,
					AutoMinorVersionUpgrade: false,
					PubliclyAccessible:      false,
					Subnets:                 []Subnet{},
					region:                  awslib.DefaultRegion,
				},
				DBInstance{
					Identifier:              identifier2,
					Arn:                     arn2,
					StorageEncrypted:        true,
					AutoMinorVersionUpgrade: true,
					PubliclyAccessible:      true, Subnets: []Subnet{
						{ID: identifier, RouteTable: nil},
						{ID: identifier2, RouteTable: &RouteTable{
							ID:     identifier,
							Routes: []Route{{DestinationCidrBlock: &destinationCidrBlock, GatewayId: &gatewayId}},
						}},
					},
					region: awslib.DefaultRegion,
				},
			},
		},
	}

	for _, test := range tests {
		clients := map[string]Client{}
		for fn, e := range test.rdsClientMockReturnVals {
			for region, calls := range e {
				m := &MockClient{}
				m.On(fn, mock.Anything, mock.Anything).Return(calls...).Once()
				clients[region] = m
			}
		}

		mockEc2 := &ec2.MockElasticCompute{}
		for _, calls := range test.ec2GetRouteTableForSubnetVals {
			mockEc2.On("GetRouteTableForSubnet", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(calls...).Once()
		}

		rdsProvider := Provider{
			log:     testhelper.NewLogger(s.T()),
			clients: clients,
			ec2:     mockEc2,
		}

		t := s.T()
		ctx := t.Context()

		results, err := rdsProvider.DescribeDBInstances(ctx)
		s.Require().NoError(err)
		s.Equal(test.expected, results)
	}
}
