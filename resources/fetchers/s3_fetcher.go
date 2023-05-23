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

package fetchers

import (
	"context"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/elastic-agent-libs/logp"
)

type S3Fetcher struct {
	log        *logp.Logger
	cfg        S3FetcherConfig
	s3         s3.S3
	resourceCh chan fetching.ResourceInfo
}

type S3FetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type S3Resource struct {
	bucket awslib.AwsResource
}

func (f *S3Fetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting S3Fetcher.Fetch")
	buckets, err := f.s3.DescribeBuckets(ctx)
	if err != nil {
		f.log.Errorf("failed to load buckets from S3: %v", err)
		return nil
	}

	for _, bucket := range buckets {
		resource := S3Resource{bucket}
		f.log.Debugf("Fetched bucket: %s", bucket.GetResourceName())
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      resource,
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f *S3Fetcher) Stop() {}

func (r S3Resource) GetData() interface{} {
	return r.bucket
}

func (r S3Resource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.bucket.GetResourceArn(),
		Type:    fetching.CloudStorage,
		SubType: r.bucket.GetResourceType(),
		Name:    r.bucket.GetResourceName(),
		Region:  r.bucket.GetRegion(),
	}, nil
}

func (r S3Resource) GetElasticCommonData() any { return nil }
