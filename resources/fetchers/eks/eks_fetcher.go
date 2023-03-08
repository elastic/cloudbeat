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

package eks

import (
	"context"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

const (
	// Type fetcher
	Type = "aws-eks"
)

type EksFetcher struct {
	log         *logp.Logger
	cfg         EksFetcherConfig
	eksProvider awslib.EksClusterDescriber
	resourceCh  chan fetching.ResourceInfo
}

type EksFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
	ClusterName                   string `config:"clusterName"`
}

func New(options ...Option) *EksFetcher {
	f := &EksFetcher{}
	for _, opt := range options {
		opt(f)
	}
	return f
}

func (f EksFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting EksFetcher.Fetch")

	result, err := f.eksProvider.DescribeCluster(ctx, f.cfg.ClusterName)
	if err != nil {
		return err
	}

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      EksResource{result},
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f EksFetcher) Stop() {
}
