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

package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestProvider_DescribeTrails(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name    string
		clients map[string]func() any
		want    []awslib.AwsResource
		wantErr bool
	}{
		{
			name: "Failed to describe trails",
			clients: map[string]func() any{
				"s3Provider": func() any {
					return &s3.MockS3{}
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", ctx).Return(nil, errors.New("bad, very bad"))
					return m
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No trails found",
			clients: map[string]func() any{
				"s3Provider": func() any {
					return &s3.MockS3{}
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", ctx).Return([]cloudtrail.TrailInfo{}, nil)
					return m
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Trails found without s3 bucket data",
			clients: map[string]func() any{
				"s3Provider": func() any {
					m := &s3.MockS3{}
					m.On("GetBucketPolicy", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("no bucket policy"))
					m.On("GetBucketACL", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("no bucket ACL data"))
					m.On("GetBucketLogging", ctx, mock.Anything, mock.Anything).Return(s3.Logging{}, errors.New("no bucket logging data"))
					return m
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", ctx).Return([]cloudtrail.TrailInfo{{Trail: types.Trail{
						TrailARN: aws.String("test-arn"),
					}}}, nil)
					return m
				},
			},
			want: []awslib.AwsResource{
				EnrichedTrail{
					TrailInfo: cloudtrail.TrailInfo{Trail: types.Trail{
						TrailARN: aws.String("test-arn"),
					}},
					BucketInfo: TrailBucket{},
				},
			},
			wantErr: false,
		},
		{
			name: "Trails found with s3 bucket data",
			clients: map[string]func() any{
				"s3Provider": func() any {
					m := &s3.MockS3{}
					m.On("GetBucketPolicy", ctx, mock.Anything, mock.Anything).Return(s3.BucketPolicy{}, nil)
					m.On("GetBucketACL", ctx, mock.Anything, mock.Anything).Return(&s3Client.GetBucketAclOutput{}, nil)
					m.On("GetBucketLogging", ctx, mock.Anything, mock.Anything).Return(s3.Logging{Enabled: true}, nil)
					return m
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", ctx).Return([]cloudtrail.TrailInfo{
						{
							Trail: types.Trail{
								TrailARN: aws.String("test-arn"),
							},
						},
					}, nil)
					return m
				},
			},
			want: []awslib.AwsResource{
				EnrichedTrail{
					TrailInfo: cloudtrail.TrailInfo{Trail: types.Trail{
						TrailARN: aws.String("test-arn"),
					}},
					BucketInfo: TrailBucket{
						ACL:    &s3Client.GetBucketAclOutput{},
						Policy: s3.BucketPolicy{},
						Logging: s3.Logging{
							Enabled: true,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:           testhelper.NewLogger(t),
				s3Provider:    tt.clients["s3Provider"]().(s3.S3),
				trailProvider: tt.clients["cloudTrailProvider"]().(cloudtrail.TrailService),
			}

			got, err := p.DescribeTrails(ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			for i, g := range got {
				assert.Equal(t, tt.want[i], g)
			}
		})
	}
}
