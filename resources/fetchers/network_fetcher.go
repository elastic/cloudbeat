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
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/elastic-agent-libs/logp"
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

type NetworkResource struct {
	awslib.AwsResource
	identity *awslib.Identity
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

func (r NetworkResource) GetData() any {
	return r.AwsResource
}

func (r NetworkResource) GetMetadata() (fetching.ResourceMetadata, error) {
	identifier := r.GetResourceArn()
	return fetching.ResourceMetadata{
		ID:      identifier,
		Type:    fetching.EC2Identity,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
		Region:  r.GetRegion(),
	}, nil
}

func (r NetworkResource) GetElasticCommonData() any { return nil }

func (f NetworkFetcher) aggregateResources(ctx context.Context, client ec2.ElasticCompute) ([]awslib.AwsResource, error) {
	var resources []awslib.AwsResource
	nacl, err := client.DescribeNetworkAcl(ctx)
	if err != nil {
		f.log.Errorf("failed to describe network acl: %v", err)
	}
	resources = append(resources, nacl...)

	securityGroups, err := client.DescribeSecurityGroups(ctx)
	if err != nil {
		f.log.Errorf("failed to describe security groups: %v", err)
	}
	resources = append(resources, securityGroups...)
	vpcs, err := client.DescribeVPCs(ctx)
	if err != nil {
		f.log.Errorf("failed to describe vpcs: %v", err)
	}
	resources = append(resources, vpcs...)
	ebsEncryption, err := client.GetEbsEncryptionByDefault(ctx)
	if err != nil {
		f.log.Errorf("failed to get ebs encryption by default: %v", err)
	}

	if ebsEncryption != nil {
		resources = append(resources, ebsEncryption...)
	}

	return resources, nil
}
