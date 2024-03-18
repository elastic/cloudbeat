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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
)

type S3BucketFetcher struct {
	logger   *logp.Logger
	provider s3BucketProvider
}

var s3BucketClassification = inventory.AssetClassification{
	Category:    inventory.CategoryInfrastructure,
	SubCategory: inventory.SubCategoryStorage,
	Type:        inventory.TypeObjectStorage,
	SubStype:    inventory.SubTypeS3,
}

type s3BucketProvider interface {
	DescribeBuckets(ctx context.Context) ([]awslib.AwsResource, error)
}

func NewS3BucketFetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) inventory.AssetFetcher {
	provider := s3.NewProvider(logger, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)
	return &S3BucketFetcher{
		logger:   logger,
		provider: provider,
	}
}

func (s S3BucketFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	awsBuckets, err := s.provider.DescribeBuckets(ctx)
	if err != nil {
		s.logger.Errorf("Could not list s3 buckets: %v", err)
		if len(awsBuckets) == 0 {
			return
		}
	}

	buckets := lo.Map(awsBuckets, func(item awslib.AwsResource, _ int) s3.BucketDescription {
		return item.(s3.BucketDescription)
	})

	for _, bucket := range buckets {
		assetChannel <- inventory.NewAssetEvent(
			s3BucketClassification,
			bucket.GetResourceArn(),
			bucket.GetResourceName(),

			inventory.WithRawAsset(bucket),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   bucket.GetRegion(),
			}),
		)
	}
}
