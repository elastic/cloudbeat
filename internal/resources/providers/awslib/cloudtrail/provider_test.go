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

package cloudtrail

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type (
	cloudtrailClientMocks map[string][][]any
)

func TestProvider_DescribeCloudTrails(t *testing.T) {
	tests := []struct {
		name                           string
		regions                        []string
		want                           []TrailInfo
		wantErr                        bool
		cloudtrailClientMockReturnVals cloudtrailClientMocks
	}{
		{
			name:    "Retrieve trail info from all regions",
			regions: []string{"us-east-1", "us-west-1"},
			want: []TrailInfo{
				{
					Trail: types.Trail{
						TrailARN:                  aws.String("arn:aws:cloudtrail:us-east-1:123456789012:trail/mytrail"),
						Name:                      aws.String("trail"),
						LogFileValidationEnabled:  aws.Bool(true),
						IsMultiRegionTrail:        aws.Bool(true),
						HasCustomEventSelectors:   aws.Bool(true),
						KmsKeyId:                  aws.String("kmsKey_123"),
						CloudWatchLogsLogGroupArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:my-log-group"),
						HomeRegion:                aws.String("us-east-1"),
						S3BucketName:              aws.String("trails_bucket"),
						SnsTopicARN:               aws.String("arn:aws:sns:us-east-1:123456789012:my-topic"),
					},
					Status: &cloudtrail.GetTrailStatusOutput{
						IsLogging: aws.Bool(true),
					},
					EventSelectors: []types.EventSelector{{DataResources: []types.DataResource{
						{
							Type:   aws.String("AWS::S3::Object"),
							Values: []string{"bucket"},
						},
					}, ReadWriteType: types.ReadWriteTypeAll}},
				},
				{
					Trail: types.Trail{
						TrailARN:                  aws.String("arn:aws:cloudtrail:us-west-1:123456789012:trail/mytrail"),
						Name:                      aws.String("trail"),
						LogFileValidationEnabled:  aws.Bool(true),
						IsMultiRegionTrail:        aws.Bool(true),
						HomeRegion:                aws.String("us-west-1"),
						HasCustomEventSelectors:   aws.Bool(true),
						KmsKeyId:                  aws.String("kmsKey_123"),
						CloudWatchLogsLogGroupArn: aws.String("arn:aws:logs:us-west-1:123456789012:log-group:my-log-group"),
						S3BucketName:              aws.String("trails_bucket"),
						SnsTopicARN:               aws.String("arn:aws:sns:us-west-1:123456789012:my-topic"),
					},
					Status: &cloudtrail.GetTrailStatusOutput{
						IsLogging: aws.Bool(false),
					},
					EventSelectors: nil,
				},
			},
			wantErr: false,
			cloudtrailClientMockReturnVals: cloudtrailClientMocks{
				"DescribeTrails": {
					{&cloudtrail.DescribeTrailsOutput{
						TrailList: []types.Trail{
							{
								TrailARN:                  aws.String("arn:aws:cloudtrail:us-east-1:123456789012:trail/mytrail"),
								Name:                      aws.String("trail"),
								CloudWatchLogsLogGroupArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:my-log-group"),
								HasCustomEventSelectors:   aws.Bool(true),
								IsMultiRegionTrail:        aws.Bool(true),
								KmsKeyId:                  aws.String("kmsKey_123"),
								LogFileValidationEnabled:  aws.Bool(true),
								S3BucketName:              aws.String("trails_bucket"),
								SnsTopicARN:               aws.String("arn:aws:sns:us-east-1:123456789012:my-topic"),
								HomeRegion:                aws.String("us-east-1"),
							},
							{
								CloudWatchLogsLogGroupArn: aws.String("arn:aws:logs:us-west-1:123456789012:log-group:my-log-group"),
								HasCustomEventSelectors:   aws.Bool(true),
								IsMultiRegionTrail:        aws.Bool(true),
								KmsKeyId:                  aws.String("kmsKey_123"),
								TrailARN:                  aws.String("arn:aws:cloudtrail:us-west-1:123456789012:trail/mytrail"),
								LogFileValidationEnabled:  aws.Bool(true),
								Name:                      aws.String("trail"),
								S3BucketName:              aws.String("trails_bucket"),
								SnsTopicARN:               aws.String("arn:aws:sns:us-west-1:123456789012:my-topic"),
								HomeRegion:                aws.String("us-west-1"),
							},
						},
					}, nil},
				},
				"GetTrailStatus": {
					{&cloudtrail.GetTrailStatusOutput{IsLogging: aws.Bool(true)}, nil},
					{&cloudtrail.GetTrailStatusOutput{IsLogging: aws.Bool(false)}, nil},
				},
				"GetEventSelectors": {
					{&cloudtrail.GetEventSelectorsOutput{
						EventSelectors: []types.EventSelector{
							{
								DataResources: []types.DataResource{{
									Type:   aws.String("AWS::S3::Object"),
									Values: []string{"bucket"},
								}},
								ReadWriteType: types.ReadWriteTypeAll,
							},
						},
					}, nil}, {&cloudtrail.GetEventSelectorsOutput{}, nil},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			clientMock := &MockClient{}
			for funcName, calls := range tt.cloudtrailClientMockReturnVals {
				for _, returnVals := range calls {
					clientMock.On(funcName, ctx, mock.Anything).Return(returnVals...).Once()
				}
			}

			p := &Provider{
				log:     testhelper.NewLogger(t),
				clients: createMockClients(clientMock, tt.regions),
			}

			trails, err := p.DescribeTrails(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			for i, trail := range trails {
				assert.Equal(t, tt.want[i], trail)
			}
		})
	}
}

func createMockClients(c Client, regions []string) map[string]Client {
	m := make(map[string]Client, 0)
	for _, clientRegion := range regions {
		m[clientRegion] = c
	}

	return m
}
