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

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type ACLFetcher struct {
	log           *logp.Logger
	aclProvider   ec2.ElasticCompute
	cfg           ACLFetcherConfig
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *awslib.Identity
}

type ACLFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type ACLResource struct {
	awslib.AwsResource
	identity *awslib.Identity
}

// Fetch collects ACL resources
// The resources are enriched by the provider and being send to evaluation.
func (f ACLFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting ACLFetcher.Fetch")

	resources, err := f.aclProvider.DescribeNeworkAcl(ctx)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: ACLResource{
				AwsResource: resource,
				identity:    f.cloudIdentity,
			},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f ACLFetcher) Stop() {}

func (r ACLResource) GetData() any {
	return r.AwsResource
}

func (r ACLResource) GetMetadata() (fetching.ResourceMetadata, error) {
	identifier := r.GetResourceArn()
	return fetching.ResourceMetadata{
		ID:      identifier,
		Type:    fetching.EC2Identity,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
	}, nil
}
func (r ACLResource) GetElasticCommonData() any { return nil }
