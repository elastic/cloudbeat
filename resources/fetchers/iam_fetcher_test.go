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
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IamFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	errorCh    chan error
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
	s.resourceCh = make(chan fetching.ResourceInfo)
	s.errorCh = make(chan error, 1)
}

func (s *IamFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
	close(s.errorCh)
}

func (s *IamFetcherTestSuite) TestIamFetcherFetch() {
	var tests = []struct {
		role        string
		iamResponse []iam.GetRolePolicyResponse
	}{
		{
			"some_role",
			[]iam.GetRolePolicyResponse{},
		},
	}

	for _, test := range tests {
		eksConfig := IAMFetcherConfig{
			BaseFetcherConfig: fetching.BaseFetcherConfig{},
			RoleName:          test.role,
		}
		iamProvider := &awslib.MockIAMRolePermissionGetter{}

		iamProvider.EXPECT().GetIAMRolePermissions(mock.Anything, test.role).
			Return(test.iamResponse, nil)

		expectedResource := IAMResource{test.iamResponse}

		eksFetcher := IAMFetcher{
			log:         s.log,
			cfg:         eksConfig,
			iamProvider: iamProvider,
			resourceCh:  s.resourceCh,
		}

		ctx := context.Background()
		go func(ch chan error) {
			ch <- eksFetcher.Fetch(ctx, fetching.CycleMetadata{})
		}(s.errorCh)
		results := testhelper.WaitForResources(s.resourceCh, 1, 2)
		iamResource := results[0].Resource.(IAMResource)

		s.Equal(expectedResource, iamResource)
		s.Nil(<-s.errorCh)
	}
}
