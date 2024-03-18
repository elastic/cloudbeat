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

package aws

import (
	"context"
	"testing"
	"time"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	s3ctrltypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestS3BucketFetcher_Fetch(t *testing.T) {
	bucket1 := s3.BucketDescription{
		Name:         "bucket-1",
		SSEAlgorithm: nil,
		BucketPolicy: nil,
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
		BucketPolicy: nil,
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
			s3BucketClassification,
			"arn:aws:s3:::bucket-1",
			"bucket-1",
			inventory.WithRawAsset(bucket1),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "europe-west-1",
			}),
		),
		inventory.NewAssetEvent(
			s3BucketClassification,
			"arn:aws:s3:::bucket-2",
			"bucket-2",
			inventory.WithRawAsset(bucket2),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "europe-west-1",
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_s3_bucket")
	provider := newMockS3BucketProvider(t)
	provider.EXPECT().DescribeBuckets(mock.Anything).Return(in, nil)

	fetcher := S3BucketFetcher{
		logger:   logger,
		provider: provider,
	}

	ch := make(chan inventory.AssetEvent)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	go func() {
		fetcher.Fetch(ctx, ch)
	}()

	received := make([]inventory.AssetEvent, 0, len(expected))
	for len(expected) != len(received) {
		select {
		case <-ctx.Done():
			assert.ElementsMatch(t, expected, received)
			return
		case event := <-ch:
			received = append(received, event)
		}
	}

	assert.ElementsMatch(t, expected, received)
}
