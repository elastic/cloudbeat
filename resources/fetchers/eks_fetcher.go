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

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/elastic/cloudbeat/resources/fetching"
)

const EKSType = "aws-eks"

type EKSFetcher struct {
	cfg         EKSFetcherConfig
	eksProvider *EKSProvider
}

type EKSFetcherConfig struct {
	fetching.BaseFetcherConfig
	ClusterName string `config:"clusterName"`
}

type EKSResource struct {
	*eks.DescribeClusterResponse
}

func NewEKSFetcher(awsCfg AwsFetcherConfig, cfg EKSFetcherConfig) (fetching.Fetcher, error) {
	eks := NewEksProvider(awsCfg.Config)

	return &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}, nil
}

func (f EKSFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	results := make([]fetching.Resource, 0)

	result, err := f.eksProvider.DescribeCluster(ctx, f.cfg.ClusterName)
	results = append(results, EKSResource{result})

	return results, err
}

func (f EKSFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r EKSResource) GetID() (string, error) {
	return "", nil
}

func (r EKSResource) GetData() interface{} {
	return r
}
