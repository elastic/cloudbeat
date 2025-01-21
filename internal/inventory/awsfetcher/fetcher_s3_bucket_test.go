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

package awsfetcher

import (
	"testing"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3ctrltypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestS3BucketFetcher_Fetch(t *testing.T) {
	bucket1 := s3.BucketDescription{
		Name:         "bucket-1",
		SSEAlgorithm: nil,
		BucketPolicy: map[string]any{
			"Version": "2012-10-17",
			"Statement": []map[string]any{
				{
					"Sid":    "Test 1",
					"Effect": "Allow",
					"Principal": map[string]any{
						"AWS":     "dima",
						"service": "aws.com",
					},
					"Action":   []string{"read", "update", "delete"},
					"Resource": []string{"s3/bucket", "s3/bucket/*"},
				},
				{
					"Sid":    "Test 2",
					"Effect": "Deny",
					"Principal": map[string]any{
						"AWS": "romulo",
					},
					"Action":   []string{"delete"},
					"Resource": []string{"s3/bucket"},
				},
			},
		},
		BucketVersioning: &s3.BucketVersioning{
			Enabled:   true,
			MfaDelete: true,
		},
		PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
			BlockPublicAcls: pointers.Ref(true),
		},
		AccountPublicAccessBlockConfiguration: &s3ctrltypes.PublicAccessBlockConfiguration{
			BlockPublicAcls: pointers.Ref(true),
		},
		Region: "europe-west-1",
	}

	bucket2 := s3.BucketDescription{
		Name:         "bucket-2",
		SSEAlgorithm: nil,
		BucketPolicy: map[string]any{
			"Version": "2012-10-17",
			"Statement": map[string]any{
				"Sid":       "Test 1",
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    "read",
				"Resource":  "s3/bucket",
			},
		},
		BucketVersioning: &s3.BucketVersioning{
			Enabled:   false,
			MfaDelete: false,
		},
		PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
			BlockPublicAcls: pointers.Ref(false),
		},
		AccountPublicAccessBlockConfiguration: &s3ctrltypes.PublicAccessBlockConfiguration{
			BlockPublicAcls: pointers.Ref(false),
		},
		Region: "europe-west-1",
	}
	in := []awslib.AwsResource{bucket1, bucket2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsS3Bucket,
			"arn:aws:s3:::bucket-1",
			"bucket-1",
			inventory.WithRawAsset(bucket1),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "europe-west-1",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS S3",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsS3Bucket,
			"arn:aws:s3:::bucket-2",
			"bucket-2",
			inventory.WithRawAsset(bucket2),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "europe-west-1",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS S3",
			}),
		),
	}

	logger := clog.NewLogger("test_fetcher_s3_bucket")
	provider := newMockS3BucketProvider(t)
	provider.EXPECT().DescribeBuckets(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newS3BucketFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
