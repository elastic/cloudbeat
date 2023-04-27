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
	"errors"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type S3FetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type s3mocksReturnVals map[string][]any

func TestS3FetcherTestSuite(t *testing.T) {
	s := new(S3FetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_s3_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *S3FetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *S3FetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *S3FetcherTestSuite) TestFetcher_Fetch() {
	var tests = []struct {
		name               string
		s3mocksReturnVals  s3mocksReturnVals
		numExpectedResults int
	}{
		{
			name: "Should not get any S3 buckets",
			s3mocksReturnVals: s3mocksReturnVals{
				"DescribeBuckets": {nil, errors.New("bad, very bad")},
			},
			numExpectedResults: 0,
		},
		{
			name: "Should get an S3 bucket",
			s3mocksReturnVals: s3mocksReturnVals{
				"DescribeBuckets": {[]awslib.AwsResource{s3.BucketDescription{Name: "my test bucket", SSEAlgorithm: nil}}, nil},
			},
			numExpectedResults: 1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s3FetcherCfg := S3FetcherConfig{
				AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
			}

			s3ProviderMock := &s3.MockS3{}
			for funcName, returnVals := range test.s3mocksReturnVals {
				s3ProviderMock.On(funcName, context.TODO()).Return(returnVals...)
			}

			s3Fetcher := S3Fetcher{
				log:        s.log,
				cfg:        s3FetcherCfg,
				s3:         s3ProviderMock,
				resourceCh: s.resourceCh,
			}

			ctx := context.Background()

			err := s3Fetcher.Fetch(ctx, fetching.CycleMetadata{})
			s.NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.Equal(test.numExpectedResults, len(results))
		})
	}
}
