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
)

type IamFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type IamProviderReturnVals struct {
	pwdPolicy awslib.AwsResource
	err       error
}

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

func (s *IamFetcherTestSuite) TestIamFetcherFetch() {
	pwdPolicy := iam.PasswordPolicy{
		ReusePreventionCount: 5,
		RequireLowercase:     true,
		RequireUppercase:     true,
		RequireNumbers:       true,
		RequireSymbols:       false,
		MaxAgeDays:           90,
		MinimumLength:        8,
	}

	testAccount := "test-account"

	var tests = []struct {
		mockReturnVal      IamProviderReturnVals
		account            string
		numExpectedResults int
	}{
		{
			mockReturnVal: IamProviderReturnVals{
				pwdPolicy: pwdPolicy,
				err:       nil,
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			mockReturnVal: IamProviderReturnVals{
				pwdPolicy: nil,
				err:       errors.New("Fail to fetch pwd policy"),
			},
			account:            testAccount,
			numExpectedResults: 0,
		},
	}

	for _, test := range tests {
		iamCfg := IAMFetcherConfig{
			AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
		}

		iamProvider := &iam.MockAccessManagement{}
		iamProvider.EXPECT().GetPasswordPolicy(context.TODO()).Return(test.mockReturnVal.pwdPolicy, test.mockReturnVal.err)

		eksFetcher := IAMFetcher{
			log:         s.log,
			iamProvider: iamProvider,
			cfg:         iamCfg,
			resourceCh:  s.resourceCh,
			cloudIdentity: &awslib.Identity{
				Account: &test.account,
			},
		}

		ctx := context.Background()

		err := eksFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(results), test.numExpectedResults)
		if test.mockReturnVal.err == nil {
			s.NoError(err)
		}
	}
}
