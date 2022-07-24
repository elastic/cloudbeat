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
	"github.com/pkg/errors"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
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

type EksResource struct {
	awslib.EksCluster
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

func (r EksResource) GetData() interface{} {
	return r
}

func (r EksResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if r.Arn == nil || r.Name == nil {
		return fetching.ResourceMetadata{}, errors.New("received nil pointer")
	}

	return fetching.ResourceMetadata{
		ID:      *r.Arn,
		Type:    fetching.CloudContainerMgmt,
		SubType: fetching.EKSType,
		Name:    *r.Name,
	}, nil
}

func (r EksResource) GetElasticCommonData() any { return nil }
