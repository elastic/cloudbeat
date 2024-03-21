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

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type Ec2InstanceFetcher struct {
	logger      *logp.Logger
	provider    ec2InstancesProvider
	AccountId   string
	AccountName string
}

type ec2InstancesProvider interface {
	DescribeInstances(ctx context.Context) ([]*ec2.Ec2Instance, error)
}

var ec2InstanceClassification = inventory.AssetClassification{
	Category:    inventory.CategoryInfrastructure,
	SubCategory: inventory.SubCategoryCompute,
	Type:        inventory.TypeVirtualMachine,
	SubType:     inventory.SubTypeEC2,
}

func newEc2Fetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) inventory.AssetFetcher {
	provider := ec2.NewEC2Provider(logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	return &Ec2InstanceFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (e *Ec2InstanceFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	instances, err := e.provider.DescribeInstances(ctx)
	if err != nil {
		e.logger.Errorf("Could not list ec2 instances: %v", err)
		return
	}

	for _, instance := range instances {
		if instance == nil {
			continue
		}

		iamFetcher := inventory.EmptyEnricher()
		if instance.IamInstanceProfile != nil {
			iamFetcher = inventory.WithIAM(inventory.AssetIAM{
				Id:  instance.IamInstanceProfile.Id,
				Arn: instance.IamInstanceProfile.Arn,
			})
		}

		assetChannel <- inventory.NewAssetEvent(
			ec2InstanceClassification,
			instance.GetResourceArn(),
			instance.GetResourceName(),

			inventory.WithRawAsset(instance),
			inventory.WithTags(e.getTags(instance)),
			inventory.WithCloud(inventory.AssetCloud{
				Provider:         inventory.AwsCloudProvider,
				Region:           instance.Region,
				AvailabilityZone: e.getAvailabilityZone(instance),
				Account: inventory.AssetCloudAccount{
					Id:   e.AccountId,
					Name: e.AccountName,
				},
				Instance: &inventory.AssetCloudInstance{
					Id:   pointers.Deref(instance.InstanceId),
					Name: instance.GetResourceName(),
				},
				Machine: &inventory.AssetCloudMachine{
					MachineType: string(instance.InstanceType),
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS EC2",
				},
			}),
			inventory.WithHost(inventory.AssetHost{
				Architecture:    string(instance.Architecture),
				ImageId:         instance.ImageId,
				InstanceType:    string(instance.InstanceType),
				Platform:        string(instance.Platform),
				PlatformDetails: instance.PlatformDetails,
			}),
			iamFetcher,
			inventory.WithNetwork(inventory.AssetNetwork{
				NetworkId:        instance.VpcId,
				SubnetId:         instance.SubnetId,
				Ipv6Address:      instance.Ipv6Address,
				PublicIpAddress:  instance.PublicIpAddress,
				PrivateIpAddress: instance.PrivateIpAddress,
				PublicDnsName:    instance.PublicDnsName,
				PrivateDnsName:   instance.PrivateDnsName,
			}),
		)
	}
}

func (e *Ec2InstanceFetcher) getTags(instance *ec2.Ec2Instance) map[string]string {
	tags := make(map[string]string, len(instance.Tags))
	for _, t := range instance.Tags {
		if t.Key == nil {
			continue
		}

		tags[*t.Key] = pointers.Deref(t.Value)
	}
	return tags
}

func (e *Ec2InstanceFetcher) getAvailabilityZone(instance *ec2.Ec2Instance) *string {
	if instance.Placement == nil {
		return nil
	}

	return instance.Placement.AvailabilityZone
}
