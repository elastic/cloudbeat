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
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	"github.com/elastic/elastic-agent-libs/logp"
)

type KmsFetcher struct {
	log        *logp.Logger
	cfg        KmsFetcherConfig
	kms        kms.KMS
	resourceCh chan fetching.ResourceInfo
}

type KmsFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type KmsResource struct {
	key awslib.AwsResource
}

func (f *KmsFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Info("Starting KMSFetcher.Fetch")

	keys, err := f.kms.DescribeSymmetricKeys(ctx)
	if err != nil {
		return err
	}

	for _, key := range keys {
		resource := KmsResource{key}
		f.log.Debugf("Fetched key: %s", key.GetResourceName())
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      resource,
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f *KmsFetcher) Stop() {}

func (r KmsResource) GetData() interface{} {
	return r.key
}

func (r KmsResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.key.GetResourceArn(),
		Type:    fetching.CloudStorage,
		SubType: r.key.GetResourceType(),
		Name:    r.key.GetResourceName(),
	}, nil
}

func (r KmsResource) GetElasticCommonData() any { return nil }
