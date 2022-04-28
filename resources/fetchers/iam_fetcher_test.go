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
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IamFetcherTestSuite struct {
	suite.Suite
}

func TestIamFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(IamFetcherTestSuite))
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
			cfg:         eksConfig,
			iamProvider: iamProvider,
		}

		ctx := context.Background()
		result, err := eksFetcher.Fetch(ctx)
		s.Nil(err)

		iamResource := result[0].(IAMResource)
		s.Equal(expectedResource, iamResource)
	}
}
