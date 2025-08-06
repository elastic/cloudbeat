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

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
)

type NetworkFetcher struct {
	log        *clog.Logger
	ec2Client  ec2.ElasticCompute
	resourceCh chan fetching.ResourceInfo
}

type ACLFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type NetworkResource struct {
	awslib.AwsResource
}

func NewNetworkFetcher(log *clog.Logger, ec2Client ec2.ElasticCompute, ch chan fetching.ResourceInfo) *NetworkFetcher {
	return &NetworkFetcher{
		log:        log,
		ec2Client:  ec2Client,
		resourceCh: ch,
	}
}

// Fetch collects network resource such as network acl and security groups
func (f NetworkFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Debug("Starting NetworkFetcher.Fetch")
	resources := f.aggregateResources(ctx, f.ec2Client)

	for _, resource := range resources {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: NetworkResource{
				AwsResource: resource,
			},
			CycleMetadata: cycleMetadata,
		}
	}

	return nil
}

func (f NetworkFetcher) Stop() {}

func (r NetworkResource) GetData() any {
	return r.AwsResource
}

func (r NetworkResource) GetIds() []string {
	return []string{r.GetResourceArn()}
}

func (r NetworkResource) GetMetadata() (fetching.ResourceMetadata, error) {
	identifier := r.GetResourceArn()
	return fetching.ResourceMetadata{
		ID:      identifier,
		Type:    fetching.CloudCompute,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
		Region:  r.GetRegion(),
	}, nil
}

func (r NetworkResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud.service.name": "EC2",
	}, nil
}

func (f NetworkFetcher) aggregateResources(ctx context.Context, client ec2.ElasticCompute) []awslib.AwsResource {
	var resources []awslib.AwsResource
	nacl, err := client.DescribeNetworkAcl(ctx)
	if err != nil {
		f.log.Errorf(ctx, "failed to describe network acl: %v", err)
	}
	resources = append(resources, nacl...)

	securityGroups, err := client.DescribeSecurityGroups(ctx)
	if err != nil {
		f.log.Errorf(ctx, "failed to describe security groups: %v", err)
	}
	resources = append(resources, securityGroups...)
	vpcs, err := client.DescribeVpcs(ctx)
	if err != nil {
		f.log.Errorf(ctx, "failed to describe vpcs: %v", err)
	}
	resources = append(resources, vpcs...)
	ebsEncryption, err := client.GetEbsEncryptionByDefault(ctx)
	if err != nil {
		f.log.Errorf(ctx, "failed to get ebs encryption by default: %v", err)
	}

	if ebsEncryption != nil {
		resources = append(resources, ebsEncryption...)
	}

	return resources
}
