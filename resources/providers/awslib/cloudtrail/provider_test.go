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
					TrailARN:                  "arn:aws:cloudtrail:us-east-1:123456789012:trail/mytrail",
					Name:                      "trail",
					EnableLogFileValidation:   true,
					IsMultiRegion:             true,
					KMSKeyID:                  "kmsKey_123",
					CloudWatchLogsLogGroupArn: "arn:aws:logs:us-east-1:123456789012:log-group:my-log-group",
					IsLogging:                 true,
					Region:                    "us-east-1",
					BucketName:                "trails_bucket",
					SnsTopicARN:               "arn:aws:sns:us-east-1:123456789012:my-topic",
					EventSelectors: []EventSelector{{DataResources: []DataResource{
						{
							Type:   "AWS::S3::Object",
							Values: []string{"bucket"},
						}}, ReadWriteType: types.ReadWriteTypeAll}},
				},
				{
					TrailARN:                  "arn:aws:cloudtrail:us-west-1:123456789012:trail/mytrail",
					Name:                      "trail",
					EnableLogFileValidation:   true,
					IsMultiRegion:             true,
					Region:                    "us-west-1",
					KMSKeyID:                  "kmsKey_123",
					CloudWatchLogsLogGroupArn: "arn:aws:logs:us-west-1:123456789012:log-group:my-log-group",
					IsLogging:                 false,
					BucketName:                "trails_bucket",
					SnsTopicARN:               "arn:aws:sns:us-west-1:123456789012:my-topic",
					EventSelectors:            nil,
				},
			},
			wantErr: false,
			cloudtrailClientMockReturnVals: cloudtrailClientMocks{
				"DescribeTrails": {
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
							}},
					}, nil}, {&cloudtrail.GetEventSelectorsOutput{}, nil},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientMock := &MockClient{}
			for funcName, calls := range tt.cloudtrailClientMockReturnVals {
				for _, returnVals := range calls {
					clientMock.On(funcName, context.TODO(), mock.Anything).Return(returnVals...).Once()
				}
			}

			p := &Provider{
				log:     logp.NewLogger("TestProvider_DescribeCloudTrails"),
				clients: createMockClients(clientMock, tt.regions),
			}

			trails, err := p.DescribeTrails(context.Background())
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

func createMockClients(c Client, regions []string) map[string]Client {
	var m = make(map[string]Client, 0)
	for _, clientRegion := range regions {
		m[clientRegion] = c
	}

	return m
}
