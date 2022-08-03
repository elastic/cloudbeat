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
	iam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IamFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
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
	var tests = []struct {
		role        string
		iamResponse []awslib.RolePolicyInfo
	}{
		{
			role: "some_role",
			iamResponse: []awslib.RolePolicyInfo{
				{
					PolicyARN:           "arn:aws:iam::123456789012:policy/TestPolicy",
					GetRolePolicyOutput: iam.GetRolePolicyOutput{},
				},
			},
		},
	}

	for _, test := range tests {
		eksConfig := IAMFetcherConfig{
			AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
			RoleName:             test.role,
		}
		iamProvider := &awslib.MockIamRolePermissionGetter{}

		iamProvider.EXPECT().GetIAMRolePermissions(mock.Anything, test.role).
			Return(test.iamResponse, nil)

		expectedResource := IAMResource{test.iamResponse[0]}

		eksFetcher := IAMFetcher{
			log:         s.log,
			cfg:         eksConfig,
			iamProvider: iamProvider,
			resourceCh:  s.resourceCh,
		}

		ctx := context.Background()

		err := eksFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)
		iamResource := results[0].Resource.(IAMResource)

		s.Equal(expectedResource, iamResource)
		s.NoError(err)
	}
}
