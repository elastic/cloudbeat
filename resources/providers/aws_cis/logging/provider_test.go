package logging

import (
	"context"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestProvider_DescribeTrails(t *testing.T) {
	logger := logp.NewLogger("cloudbeat_logging_provider_test")

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
					m.On("DescribeTrails", context.Background()).Return(nil, errors.New("bad, very bad"))
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
					m.On("DescribeTrails", context.Background()).Return([]cloudtrail.TrailInfo{}, nil)
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
					m.On("GetBucketPolicy", context.Background(), mock.Anything, mock.Anything).Return(nil, errors.New("no bucket policy"))
					m.On("GetBucketACL", context.Background(), mock.Anything, mock.Anything).Return(nil, errors.New("no bucket ACL data"))
					m.On("GetBucketLogging", context.Background(), mock.Anything, mock.Anything).Return(s3.Logging{}, errors.New("no bucket logging data"))
					return m
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", context.Background()).Return([]cloudtrail.TrailInfo{{Name: "test-trail"}}, nil)
					return m
				},
			},
			want: []awslib.AwsResource{
				EnrichedTrail{
					TrailInfo:  cloudtrail.TrailInfo{Name: "test-trail"},
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
					m.On("GetBucketPolicy", context.Background(), mock.Anything, mock.Anything).Return(s3.BucketPolicy{}, nil)
					m.On("GetBucketACL", context.Background(), mock.Anything, mock.Anything).Return([]s3Types.Grant{}, nil)
					m.On("GetBucketLogging", context.Background(), mock.Anything, mock.Anything).Return(s3.Logging{Enabled: true}, nil)
					return m
				},
				"cloudTrailProvider": func() any {
					m := &cloudtrail.MockTrailService{}
					m.On("DescribeTrails", context.Background()).Return([]cloudtrail.TrailInfo{{Name: "test-trail"}}, nil)
					return m
				},
			},
			want: []awslib.AwsResource{
				EnrichedTrail{
					TrailInfo: cloudtrail.TrailInfo{Name: "test-trail"},
					BucketInfo: TrailBucket{
						Grants: []s3Types.Grant{},
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
				log:           logger,
				s3Provider:    tt.clients["s3Provider"]().(s3.S3),
				trailProvider: tt.clients["cloudTrailProvider"]().(cloudtrail.TrailService),
			}

			got, err := p.DescribeTrails(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeTrails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DescribeTrails() got = %v, want %v", got, tt.want)
			}
		})
	}
}
