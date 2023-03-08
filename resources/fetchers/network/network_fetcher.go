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

package network

import (
	"context"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	// Type fetcher
	Type = "aws-ec2-network"
)

type NetworkFetcher struct {
	log           *logp.Logger
	ec2Client     ec2.ElasticCompute
	cfg           ACLFetcherConfig
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *awslib.Identity
}

type ACLFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

func New(options ...Option) *NetworkFetcher {
	f := &NetworkFetcher{}
	for _, opt := range options {
		opt(f)
	}
	return f
}

// Fetch collects network resource such as network acl and security groups
func (f NetworkFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting NetworkFetcher.Fetch")
	resources, err := f.aggregateResources(ctx, f.ec2Client)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: NetworkResource{
				AwsResource: resource,
				identity:    f.cloudIdentity,
			},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f NetworkFetcher) Stop() {}
