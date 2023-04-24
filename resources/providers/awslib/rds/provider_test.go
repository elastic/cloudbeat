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
	"context"
	"testing"

	rdsClient "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}

type rdsClientMockReturnVals map[string]map[string][]any

var (
	identifier  = "identifier"
	identifier2 = "identifier2"
	arn         = "arn"
	arn2        = "arn2"
)

func TestProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_rds_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {}

func (s *ProviderTestSuite) TearDownTest() {}

func (s *ProviderTestSuite) TestProvider_DescribeDBInstances() {
	tests := []struct {
		name                    string
		rdsClientMockReturnVals rdsClientMockReturnVals
		expected                []awslib.AwsResource
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
							{DBInstanceIdentifier: &identifier, DBInstanceArn: &arn, StorageEncrypted: false, AutoMinorVersionUpgrade: false},
							{DBInstanceIdentifier: &identifier2, DBInstanceArn: &arn2, StorageEncrypted: true, AutoMinorVersionUpgrade: true},
						},
					}, nil},
				},
			},
			expected: []awslib.AwsResource{
				DBInstance{Identifier: identifier, Arn: arn, StorageEncrypted: false, AutoMinorVersionUpgrade: false, region: awslib.DefaultRegion},
				DBInstance{Identifier: identifier2, Arn: arn2, StorageEncrypted: true, AutoMinorVersionUpgrade: true, region: awslib.DefaultRegion},
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
		rdsProvider := Provider{
			log:     s.log,
			clients: clients,
		}

		ctx := context.Background()

		results, err := rdsProvider.DescribeDBInstances(ctx)
		s.NoError(err)
		s.Equal(test.expected, results)
	}
}
