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

package awsfetcher

import (
	"context"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type ec2InstanceFetcher struct {
	logger      *clog.Logger
	provider    ec2InstancesProvider
	AccountId   string
	AccountName string
}

type ec2InstancesProvider interface {
	DescribeInstances(ctx context.Context) ([]*ec2.Ec2Instance, error)
}

func newEc2InstancesFetcher(logger *clog.Logger, identity *cloud.Identity, provider ec2InstancesProvider) inventory.AssetFetcher {
	return &ec2InstanceFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (e *ec2InstanceFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	e.logger.Info("Fetching EC2 Instances")
	defer e.logger.Info("Fetching EC2 Instances - Finished")

	instances, err := e.provider.DescribeInstances(ctx)
	if err != nil {
		e.logger.Errorf("Could not list ec2 instances: %v", err)
		return
	}

	for _, i := range instances {
		if i == nil {
			continue
		}

		iamFetcher := inventory.EmptyEnricher()
		if i.IamInstanceProfile != nil {
			iamFetcher = inventory.WithUser(inventory.User{
				ID: pointers.Deref(i.IamInstanceProfile.Arn),
			})
		}

		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			i.GetResourceArn(),
			pointers.Deref(i.PrivateDnsName),

			inventory.WithRelatedAssetIds([]string{pointers.Deref(i.InstanceId)}),
			inventory.WithRawAsset(i),
			inventory.WithLabels(e.getTags(i)),
			inventory.WithCloud(inventory.Cloud{
				Provider:         inventory.AwsCloudProvider,
				Region:           i.Region,
				AvailabilityZone: e.getAvailabilityZone(i),
				AccountID:        e.AccountId,
				AccountName:      e.AccountName,
				InstanceID:       pointers.Deref(i.InstanceId),
				InstanceName:     i.GetResourceName(),
				MachineType:      string(i.InstanceType),
				ServiceName:      "AWS EC2",
			}),
			inventory.WithHost(inventory.Host{
				ID:           pointers.Deref(i.InstanceId),
				Name:         pointers.Deref(i.PrivateDnsName),
				Architecture: string(i.Architecture),
				Type:         string(i.InstanceType),
				IP:           pointers.Deref(i.PublicIpAddress),
				MacAddress:   i.GetResourceMacAddresses(),
			}),
			iamFetcher,
		)
	}
}

func (e *ec2InstanceFetcher) getTags(instance *ec2.Ec2Instance) map[string]string {
	tags := make(map[string]string, len(instance.Tags))
	for _, t := range instance.Tags {
		if t.Key == nil {
			continue
		}

		tags[*t.Key] = pointers.Deref(t.Value)
	}
	return tags
}

func (e *ec2InstanceFetcher) getAvailabilityZone(instance *ec2.Ec2Instance) string {
	if instance.Placement == nil {
		return ""
	}

	return pointers.Deref(instance.Placement.AvailabilityZone)
}
