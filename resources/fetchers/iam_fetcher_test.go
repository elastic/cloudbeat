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

package fetchers

import (
	"context"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type IamFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type mocksReturnVals map[string][]any

func TestIamFetcherTestSuite(t *testing.T) {
	s := new(IamFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_iam_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *IamFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *IamFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *IamFetcherTestSuite) TestIamFetcher_Fetch() {
	testAccount := "test-account"
	pwdPolicy := iam.PasswordPolicy{
		ReusePreventionCount: 5,
		RequireLowercase:     true,
		RequireUppercase:     true,
		RequireNumbers:       true,
		RequireSymbols:       false,
		MaxAgeDays:           90,
		MinimumLength:        8,
	}

	iamUser := iam.User{
		Name: "test",
		AccessKeys: []iam.AccessKey{{
			AccessKeyId:  "",
			Active:       false,
			CreationDate: time.Time{},
			LastAccess:   time.Time{},
			HasUsed:      false},
		},
		MFADevices:  nil,
		LastAccess:  time.Time{},
		Arn:         "testArn",
		HasLoggedIn: false,
	}

	var tests = []struct {
		name               string
		mocksReturnVals    mocksReturnVals
		account            string
		numExpectedResults int
	}{
		{
			name: "Should get password policy and an IAM user",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy": {pwdPolicy, nil},
				"GetUsers":          {[]awslib.AwsResource{iamUser}, nil},
			},
			account:            testAccount,
			numExpectedResults: 2,
		},
		{
			name: "Receives only an IAM user due to an error in GetPasswordPolicy",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy": {nil, errors.New("Fail to fetch pwd policy")},
				"GetUsers":          {[]awslib.AwsResource{iamUser}, nil},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			name: "Should get only a password policy resource due to an error in GetUsers",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy": {pwdPolicy, nil},
				"GetUsers":          {nil, errors.New("Fail to fetch iam users")},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			name: "Should not get any IAM resources",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy": {nil, errors.New("Fail to fetch pwd policy")},
				"GetUsers":          {nil, errors.New("Fail to fetch iam users")},
			},
			account:            testAccount,
			numExpectedResults: 0,
		},
	}

	for _, test := range tests {
		iamCfg := IAMFetcherConfig{
			AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
		}

		iamProviderMock := &iam.MockAccessManagement{}
		for funcName, returnVals := range test.mocksReturnVals {
			iamProviderMock.On(funcName, context.TODO()).Return(returnVals...)
		}

		iamFetcher := IAMFetcher{
			log:         s.log,
			iamProvider: iamProviderMock,
			cfg:         iamCfg,
			resourceCh:  s.resourceCh,
			cloudIdentity: &awslib.Identity{
				Account: &test.account,
			},
		}

		ctx := context.Background()

		err := iamFetcher.Fetch(ctx, fetching.CycleMetadata{})
		s.NoError(err)

		results := testhelper.CollectResources(s.resourceCh)
		s.Equal(test.numExpectedResults, len(results))
	}
}
