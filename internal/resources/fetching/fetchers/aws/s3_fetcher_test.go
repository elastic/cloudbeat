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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type S3FetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

type s3mocksReturnVals map[string][]any

func TestS3FetcherTestSuite(t *testing.T) {
	s := new(S3FetcherTestSuite)

	suite.Run(t, s)
}

func (s *S3FetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *S3FetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *S3FetcherTestSuite) TestFetcher_Fetch() {
	tests := []struct {
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
			t := s.T()
			ctx := t.Context()
			s3ProviderMock := &s3.MockS3{}
			for funcName, returnVals := range test.s3mocksReturnVals {
				s3ProviderMock.On(funcName, ctx).Return(returnVals...)
			}

			s3Fetcher := S3Fetcher{
				log:        testhelper.NewLogger(s.T()),
				s3:         s3ProviderMock,
				resourceCh: s.resourceCh,
			}

			err := s3Fetcher.Fetch(ctx, cycle.Metadata{})
			s.Require().NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.Len(results, test.numExpectedResults)
		})
	}
}

func (s *S3FetcherTestSuite) TestS3Resource_GetMetadata() {
	r := S3Resource{
		bucket: s3.BucketDescription{
			Name:         "test-bucket-name",
			SSEAlgorithm: nil,
		},
	}
	meta, err := r.GetMetadata()
	s.Require().NoError(err)
	s.Equal(fetching.ResourceMetadata{ID: "arn:aws:s3:::test-bucket-name", Type: "cloud-storage", SubType: "aws-s3", Name: "test-bucket-name"}, meta)
	m, err := r.GetElasticCommonData()
	s.Require().NoError(err)
	s.Len(m, 1)
	s.Contains(m, "cloud.service.name")
}
