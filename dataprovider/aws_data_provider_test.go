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
	"errors"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

var accountName = "accountName"
var accountId = "accountId"

type AwsDataProviderTestSuite struct {
	suite.Suite
	log *logp.Logger
}

func TestAwsDataProviderTestSuite(t *testing.T) {
	s := new(AwsDataProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_aws_data_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *AwsDataProviderTestSuite) SetupTest() {}

func (s *AwsDataProviderTestSuite) TearDownTest() {}

type identityMock struct {
	identity *awslib.Identity
	error    error
}

type aliasMock struct {
	alias string
	error error
}

func (s *AwsDataProviderTestSuite) TestAwsDataProvider_FetchData() {
	var tests = []struct {
		name         string
		identityMock identityMock
		aliasMock    aliasMock
		expected     *commonAwsData
		expectError  bool
	}{
		{
			name: "should return account id and account name",
			identityMock: identityMock{
				identity: &awslib.Identity{Account: &accountId},
				error:    nil,
			},
			aliasMock:   aliasMock{accountName, nil},
			expected:    &commonAwsData{accountId: accountId, accountName: accountName},
			expectError: false,
		},
		{
			name: "should return an error when fails to get account id",
			identityMock: identityMock{
				identity: nil,
				error:    errors.New("bla"),
			},
			aliasMock:   aliasMock{accountName, nil},
			expected:    &commonAwsData{accountId: accountId, accountName: accountName},
			expectError: true,
		},
		{
			name: "should return an error when fails to get account name",
			identityMock: identityMock{
				identity: &awslib.Identity{Account: &accountId},
				error:    nil,
			},
			aliasMock:   aliasMock{"", errors.New("bla")},
			expected:    &commonAwsData{accountId: accountId, accountName: accountName},
			expectError: true,
		},
	}

	for _, test := range tests {
		identityProvider := &awslib.MockIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(test.identityMock.identity, test.identityMock.error)

		iamProvider := &iam.MockAccessManagement{}
		iamProvider.EXPECT().GetAccountAlias(mock.Anything).Return(test.aliasMock.alias, test.aliasMock.error)

		dataProvider := awsDataProvider{log: s.log, identityProvider: identityProvider, iamProvider: iamProvider}
		ctx := context.Background()

		result, err := dataProvider.FetchData(ctx)
		if test.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
			s.Equal(result, test.expected)
		}
	}
}
