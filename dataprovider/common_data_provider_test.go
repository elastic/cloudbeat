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

package dataprovider

import (
	"context"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

var k8sData = commonK8sData{clusterId: "clusterId", nodeId: "nodeId", serverVersion: version.Version{}, clusterName: "clusterName"}
var awsData = commonAwsData{accountId: "accountId", accountName: "string"}

type DataProviderTestSuite struct {
	suite.Suite
	log                 *logp.Logger
	awsDataProviderInit EnvironmentDataProviderInit
	k8sDataProviderInit EnvironmentDataProviderInit
}

func TestDataProviderTestSuite(t *testing.T) {
	s := new(DataProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_data_provider_test_suite")

	k8sDataProviderMock := &MockEnvironmentCommonDataProvider{}
	k8sDataProviderMock.On("FetchData", mock.Anything).Return(k8sData, nil)
	s.k8sDataProviderInit = func(l *logp.Logger, c *config.Config) (EnvironmentCommonDataProvider, error) {
		return k8sDataProviderMock, nil
	}

	awsDataProviderMock := &MockEnvironmentCommonDataProvider{}
	awsDataProviderMock.On("FetchData", mock.Anything).Return(awsData, nil)
	s.awsDataProviderInit = func(l *logp.Logger, c *config.Config) (EnvironmentCommonDataProvider, error) {
		return awsDataProviderMock, nil
	}

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *DataProviderTestSuite) SetupTest() {}

func (s *DataProviderTestSuite) TearDownTest() {}

func (s *DataProviderTestSuite) TestDataProvider_FetchCommonData() {
	var tests = []struct {
		name        string
		commonData  CommonData
		benchmark   string
		expectError bool
	}{
		{
			name:        "should return k8s data for cis_k8s benchmark",
			commonData:  k8sData,
			benchmark:   "cis_k8s",
			expectError: false,
		},
		{
			name:        "should return k8s data for cis_eks benchmark",
			commonData:  k8sData,
			benchmark:   "cis_eks",
			expectError: false,
		},
		{
			name:        "should return aws data for cis_aws benchmark",
			commonData:  awsData,
			benchmark:   "cis_aws",
			expectError: false,
		},
		{
			name:        "should return an error for an unknown benchmark",
			commonData:  k8sData,
			benchmark:   "fake",
			expectError: true,
		},
	}

	for _, test := range tests {
		conf := &config.Config{Benchmark: test.benchmark}
		ctx := context.Background()

		commonDataProvider := commonDataProvider{s.log, conf, s.k8sDataProviderInit, s.awsDataProviderInit}
		result, err := commonDataProvider.FetchCommonData(ctx)
		if test.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
			s.Equal(result, test.commonData)
		}
	}
}
