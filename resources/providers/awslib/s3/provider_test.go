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

package s3

import (
	"context"
	"errors"
	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

type s3ClientMockReturnVals map[string][][]any

func TestProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_s3_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {}

func (s *ProviderTestSuite) TearDownTest() {}

var bucketName = "MyBucket"
var secondBucketName = "MyAnotherBucket"
var region types.BucketLocationConstraint = "eu-west-1"

func (s *ProviderTestSuite) TestProvider_DescribeBuckets() {
	var tests = []struct {
		name                   string
		regions                []string
		s3ClientMockReturnVals s3ClientMockReturnVals
		expected               []awslib.AwsResource
		expectError            bool
	}{
		{
			name: "Should not return any S3 buckets when there aren't any",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{}}, nil}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return any S3 buckets when there is an error",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{nil, errors.New("error")}},
			},
			expected:    nil,
			expectError: true,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return any S3 buckets when the region can not be fetched",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}},
				"GetBucketEncryption": {{nil, errors.New("bla")}},
				"GetBucketLocation":   {{nil, errors.New("bla")}},
			},
			expected:    nil,
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return an S3 bucket without encryption",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}},
				"GetBucketEncryption": {{nil, errors.New("bla")}},
				"GetBucketLocation":   {{&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: ""}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return an S3 bucket with encryption due to regions mismatch",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}},
				"GetBucketEncryption": {{nil, errors.New("regions mismatch")}},
				"GetBucketLocation":   {{&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: ""}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return an S3 bucket with encryption",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}},
				"GetBucketEncryption": {{&s3Client.GetBucketEncryptionOutput{
					ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
						Rules: []types.ServerSideEncryptionRule{
							{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
						},
					},
				}, nil}},
				"GetBucketLocation": {{&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: "AES256"}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return two S3 buckets from different regions with encryption",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}, {Name: &secondBucketName}}}, nil}},
				"GetBucketEncryption": {{&s3Client.GetBucketEncryptionOutput{
					ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
						Rules: []types.ServerSideEncryptionRule{
							{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
						},
					},
				}, nil},
					{&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
							},
						},
					}, nil}},
				"GetBucketLocation": {{&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: "AES256"}, BucketDescription{Name: secondBucketName, SSEAlgorithm: "AES256"}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return two S3 buckets from the same region with encryption",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}, {Name: &secondBucketName}}}, nil}},
				"GetBucketEncryption": {{&s3Client.GetBucketEncryptionOutput{
					ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
						Rules: []types.ServerSideEncryptionRule{
							{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
						},
					},
				}, nil},
					{&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
							},
						},
					}, nil}},
				"GetBucketLocation": {{&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: "AES256"}, BucketDescription{Name: secondBucketName, SSEAlgorithm: "AES256"}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
	}

	for _, test := range tests {
		s3ClientMock := &MockClient{}
		for funcName, returnVals := range test.s3ClientMockReturnVals {
			for _, vals := range returnVals {
				s3ClientMock.On(funcName, context.TODO(), mock.Anything).Return(vals...).Once()
			}
		}

		s3Provider := Provider{
			log:     s.log,
			clients: createMockClients(s3ClientMock, test.regions),
		}

		ctx := context.Background()

		results, err := s3Provider.DescribeBuckets(ctx)
		if test.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}

		s.Equal(test.expected, results)
	}
}

func createMockClients(c Client, regions []string) map[string]Client {
	var m = make(map[string]Client, 0)
	for _, clientRegion := range regions {
		m[clientRegion] = c
	}

	return m
}
