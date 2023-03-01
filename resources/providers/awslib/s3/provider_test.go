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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}
type mocks [2][]any
type s3ClientMockReturnVals map[string][]mocks

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
var bucketPolicy BucketPolicy = map[string]any{"foo": "bar"}
var bucketPolicyString = "{\"foo\": \"bar\"}"

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
				"ListBuckets": {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{}}, nil}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return any S3 buckets when there is an error",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{{mock.Anything, mock.Anything}, {nil, errors.New("error")}}},
			},
			expected:    nil,
			expectError: true,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return any S3 buckets when the region can not be fetched",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketEncryption": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketLocation":   {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketPolicy":     {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketVersioning": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
			},
			expected:    nil,
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return an S3 bucket without encryption, versioning, and policy",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketEncryption": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketLocation":   {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}}},
				"GetBucketPolicy":     {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketVersioning": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: nil, BucketPolicy: map[string]any(nil), BucketVersioning: nil}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return an S3 bucket without encryption, policy, and versioning due to regions mismatch",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":       {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketLocation": {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: nil, BucketPolicy: map[string]any(nil), BucketVersioning: nil}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return an S3 bucket with encryption",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketEncryption": {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketEncryptionOutput{
					ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
						Rules: []types.ServerSideEncryptionRule{
							{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
						},
					},
				}, nil}}},
				"GetBucketLocation":   {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}}},
				"GetBucketPolicy":     {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketVersioning": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: aws.String("AES256"), BucketPolicy: map[string]any(nil), BucketVersioning: nil}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return an S3 bucket with bucket policy",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketEncryption": {{{mock.Anything, mock.Anything}, {nil, &smithy.GenericAPIError{Code: EncryptionNotFoundCode}}}},
				"GetBucketLocation":   {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}}},
				"GetBucketPolicy":     {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketPolicyOutput{Policy: &bucketPolicyString}, nil}}},
				"GetBucketVersioning": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: aws.String(NoEncryptionMessage), BucketPolicy: bucketPolicy, BucketVersioning: nil}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return an S3 bucket with bucket versioning",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets":         {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}}}, nil}}},
				"GetBucketEncryption": {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketLocation":   {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}}},
				"GetBucketPolicy":     {{{mock.Anything, mock.Anything}, {nil, errors.New("bla")}}},
				"GetBucketVersioning": {{{mock.Anything, mock.Anything}, {&s3Client.GetBucketVersioningOutput{Status: "Enabled", MFADelete: "Enabled"}, nil}}},
			},
			expected:    []awslib.AwsResource{BucketDescription{Name: bucketName, SSEAlgorithm: nil, BucketPolicy: map[string]any(nil), BucketVersioning: &BucketVersioning{true, true}}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return two S3 buckets from different regions",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}, {Name: &secondBucketName}}}, nil}}},
				"GetBucketEncryption": {
					{{mock.Anything, &s3Client.GetBucketEncryptionInput{Bucket: &bucketName}}, {&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
							},
						},
					}, nil}},
					{{mock.Anything, &s3Client.GetBucketEncryptionInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "aws:kms"}},
							},
						},
					}, nil}},
				},
				"GetBucketLocation": {
					{{mock.Anything, &s3Client.GetBucketLocationInput{Bucket: &bucketName}}, {&s3Client.GetBucketLocationOutput{LocationConstraint: region}, nil}},
					{{mock.Anything, &s3Client.GetBucketLocationInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}},
				},
				"GetBucketPolicy": {
					{{mock.Anything, &s3Client.GetBucketPolicyInput{Bucket: &bucketName}}, {&s3Client.GetBucketPolicyOutput{Policy: &bucketPolicyString}, nil}},
					{{mock.Anything, &s3Client.GetBucketPolicyInput{Bucket: &secondBucketName}}, {nil, errors.New("bla")}},
				},
				"GetBucketVersioning": {
					{{mock.Anything, &s3Client.GetBucketVersioningInput{Bucket: &bucketName}}, {&s3Client.GetBucketVersioningOutput{Status: "Enabled", MFADelete: "Enabled"}, nil}},
					{{mock.Anything, &s3Client.GetBucketVersioningInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketVersioningOutput{Status: "Suspended", MFADelete: "Disabled"}, nil}},
				},
			},
			expected: []awslib.AwsResource{
				BucketDescription{Name: bucketName, SSEAlgorithm: aws.String("AES256"), BucketPolicy: bucketPolicy, BucketVersioning: &BucketVersioning{true, true}},
				BucketDescription{Name: secondBucketName, SSEAlgorithm: aws.String("aws:kms"), BucketPolicy: map[string]any(nil), BucketVersioning: &BucketVersioning{false, false}},
			},
			expectError: false,
			regions:     []string{awslib.DefaultRegion, string(region)},
		},
		{
			name: "Should return two S3 buckets from the same region",
			s3ClientMockReturnVals: s3ClientMockReturnVals{
				"ListBuckets": {{{mock.Anything, mock.Anything}, {&s3Client.ListBucketsOutput{Buckets: []types.Bucket{{Name: &bucketName}, {Name: &secondBucketName}}}, nil}}},
				"GetBucketEncryption": {
					{{mock.Anything, &s3Client.GetBucketEncryptionInput{Bucket: &bucketName}}, {&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "AES256"}},
							},
						},
					}, nil}},
					{{mock.Anything, &s3Client.GetBucketEncryptionInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketEncryptionOutput{
						ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
							Rules: []types.ServerSideEncryptionRule{
								{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: "aws:kms"}},
							},
						},
					}, nil}},
				},
				"GetBucketLocation": {
					{{mock.Anything, &s3Client.GetBucketLocationInput{Bucket: &bucketName}}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}},
					{{mock.Anything, &s3Client.GetBucketLocationInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketLocationOutput{LocationConstraint: ""}, nil}},
				},
				"GetBucketPolicy": {
					{{mock.Anything, &s3Client.GetBucketPolicyInput{Bucket: &bucketName}}, {&s3Client.GetBucketPolicyOutput{Policy: &bucketPolicyString}, nil}},
					{{mock.Anything, &s3Client.GetBucketPolicyInput{Bucket: &secondBucketName}}, {nil, errors.New("bla")}},
				},
				"GetBucketVersioning": {
					{{mock.Anything, &s3Client.GetBucketVersioningInput{Bucket: &bucketName}}, {&s3Client.GetBucketVersioningOutput{Status: "Enabled", MFADelete: "Enabled"}, nil}},
					{{mock.Anything, &s3Client.GetBucketVersioningInput{Bucket: &secondBucketName}}, {&s3Client.GetBucketVersioningOutput{Status: "Suspended", MFADelete: "Disabled"}, nil}},
				},
			},
			expected: []awslib.AwsResource{
				BucketDescription{Name: bucketName, SSEAlgorithm: aws.String("AES256"), BucketPolicy: bucketPolicy, BucketVersioning: &BucketVersioning{true, true}},
				BucketDescription{Name: secondBucketName, SSEAlgorithm: aws.String("aws:kms"), BucketPolicy: map[string]any(nil), BucketVersioning: &BucketVersioning{false, false}},
			},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
	}

	for _, test := range tests {
		s3ClientMock := &MockClient{}
		for funcName, returnVals := range test.s3ClientMockReturnVals {
			for _, vals := range returnVals {
				s3ClientMock.On(funcName, vals[0]...).Return(vals[1]...).Once()
			}
		}

		s3Provider := Provider{
			log:     s.log,
			clients: testhelper.CreateMockClients[Client](s3ClientMock, test.regions),
		}

		ctx := context.Background()

		results, err := s3Provider.DescribeBuckets(ctx)
		if test.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}

		// Using `ElementsMatch` instead of the usual `Equals` since iterating over the regions map does not produce a
		//	guaranteed order
		s.ElementsMatch(test.expected, results, fmt.Sprintf("Test '%s' failed, elements do not match", test.name))
	}
}
