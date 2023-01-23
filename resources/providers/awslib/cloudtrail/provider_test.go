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
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mocks                 []any
	cloudtrailClientMocks map[string][2]mocks
)

func TestProvider_DescribeCloudTrails(t *testing.T) {
	tests := []struct {
		name                           string
		want                           []TrailInfo
		wantErr                        bool
		cloudtrailClientMockReturnVals cloudtrailClientMocks
	}{
		{
			cloudtrailClientMockReturnVals: cloudtrailClientMocks{
				"DescribeTrails": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.DescribeTrailsOutput{
						TrailList: []types.Trail{
							{
								CloudWatchLogsLogGroupArn: aws.String("arn:aws:logs:us-east-1:123456789012:log-group:my-log-group"),
								HasCustomEventSelectors:   aws.Bool(true),
								IsMultiRegionTrail:        aws.Bool(true),
								KmsKeyId:                  aws.String("kmsKey_123"),
								TrailARN:                  aws.String("arn:aws:cloudtrail:us-east-1:123456789012:trail/mytrail"),
								LogFileValidationEnabled:  aws.Bool(true),
								Name:                      aws.String("trail"),
								S3BucketName:              aws.String("trails_bucket"),
								SnsTopicARN:               aws.String("arn:aws:sns:us-east-1:123456789012:my-topic"),
							},
						},
					}, nil},
				},
				"GetTrailStatus": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.GetTrailStatusOutput{
						IsLogging: aws.Bool(true),
					}, nil},
				},
				"GetEventSelectors": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.GetEventSelectorsOutput{
						EventSelectors: []types.EventSelector{
							{
								DataResources: []types.DataResource{{
									Type:   aws.String("AWS::S3::Object"),
									Values: []string{"bucket"},
								}},
								ReadWriteType: types.ReadWriteTypeAll,
							}},
					}, nil},
				},
			},
			want: []TrailInfo{
				{
					TrailARN:                  "arn:aws:cloudtrail:us-east-1:123456789012:trail/mytrail",
					Name:                      "trail",
					EnableLogFileValidation:   true,
					IsMultiRegion:             true,
					KMSKeyID:                  "kmsKey_123",
					CloudWatchLogsLogGroupArn: "arn:aws:logs:us-east-1:123456789012:log-group:my-log-group",
					IsLogging:                 true,
					BucketName:                "trails_bucket",
					SnsTopicARN:               "arn:aws:sns:us-east-1:123456789012:my-topic",
					EventSelectors: []EventSelector{{DataResources: []DataResource{
						{
							Type:   "AWS::S3::Object",
							Values: []string{"bucket"},
						}}, ReadWriteType: types.ReadWriteTypeAll}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			for name, call := range tt.cloudtrailClientMockReturnVals {
				mock.On(name, call[0]...).Return(call[1]...)
			}

			p := &Provider{
				log:    logp.NewLogger("TestProvider_DescribeCloudTrails"),
				client: mock,
			}

			trails, err := p.DescribeCloudTrails(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeCloudTrails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, trail := range trails {
				assert.Equal(t, tt.want[i], trail)
			}
		})
	}
}
