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
	rdsClient "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}

type rdsClientMockReturnVals map[string][][]any

var identifier = "identifier"
var identifier2 = "identifier2"
var arn = "arn"
var arn2 = "arn2"

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
	var tests = []struct {
		name                    string
		rdsClientMockReturnVals rdsClientMockReturnVals
		expected                []awslib.AwsResource
	}{
		{
			name: "Should not return any DB instances when there aren't any",
			rdsClientMockReturnVals: rdsClientMockReturnVals{"DescribeDBInstances": {
				{&rdsClient.DescribeDBInstancesOutput{DBInstances: []types.DBInstance{}}, nil},
			}},
			expected: []awslib.AwsResource(nil),
		},
		{
			name: "Should return DB instances",
			rdsClientMockReturnVals: rdsClientMockReturnVals{"DescribeDBInstances": {
				{&rdsClient.DescribeDBInstancesOutput{DBInstances: []types.DBInstance{{
					DBInstanceIdentifier: &identifier, DBInstanceArn: &arn, StorageEncrypted: false, AutoMinorVersionUpgrade: false,
				}, {
					DBInstanceIdentifier: &identifier2, DBInstanceArn: &arn2, StorageEncrypted: true, AutoMinorVersionUpgrade: true,
				}}}, nil},
			}},
			expected: []awslib.AwsResource{
				DBInstance{Identifier: identifier, Arn: arn, StorageEncrypted: false, AutoMinorVersionUpgrade: false},
				DBInstance{Identifier: identifier2, Arn: arn2, StorageEncrypted: true, AutoMinorVersionUpgrade: true},
			},
		},
	}

	for _, test := range tests {
		rdsClientMock := &MockClient{}
		for funcName, returnVals := range test.rdsClientMockReturnVals {
			for _, vals := range returnVals {
				rdsClientMock.On(funcName, context.TODO(), mock.Anything).Return(vals...).Once()
			}
		}

		rdsProvider := Provider{
			log:    s.log,
			client: rdsClientMock,
		}

		ctx := context.Background()

		results, err := rdsProvider.DescribeDBInstances(ctx)
		s.NoError(err)
		s.Equal(test.expected, results)
	}
}
