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
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type ec2InstanceFetcher struct {
	logger        *clog.Logger
	provider      ec2InstancesProvider
	AccountId     string
	AccountName   string
	statusHandler statushandler.StatusHandlerAPI
}

type ec2InstancesProvider interface {
	DescribeInstances(ctx context.Context) ([]*ec2.Ec2Instance, error)
}

func newEc2InstancesFetcher(logger *clog.Logger, identity *cloud.Identity, provider ec2InstancesProvider, statusHandler statushandler.StatusHandlerAPI) inventory.AssetFetcher {
	return &ec2InstanceFetcher{
		logger:        logger,
		provider:      provider,
		AccountId:     identity.Account,
		AccountName:   identity.AccountAlias,
		statusHandler: statusHandler,
	}
}

func (e *ec2InstanceFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	e.logger.Info("Fetching EC2 Instances")
	defer e.logger.Info("Fetching EC2 Instances - Finished")

	instances, err := e.provider.DescribeInstances(ctx)
	if err != nil {
		e.logger.Errorf("Could not list ec2 instances: %v", err)
		awslib.ReportMissingPermission(e.statusHandler, err)
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

		tags := e.getTags(i)

		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEc2Instance,
			i.GetResourceArn(),
			pointers.Deref(i.PrivateDnsName),

			inventory.WithRelatedAssetIds([]string{pointers.Deref(i.InstanceId)}),
			inventory.WithRawAsset(i),
			inventory.WithLabels(tags),
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
				IP:           buildIPs(i.PublicIpAddress, i.PrivateIpAddress),
				MacAddress:   i.GetResourceMacAddresses(),
			}),
			inventory.WithEntityAttributes(e.buildAttributes(i, tags)),
			inventory.WithCreatedAt(i.LaunchTime),
			iamFetcher,
		)
	}
}

// buildAttributes collects non-ECS, resource-specific EC2 fields into entity.attributes,
// using UpperCamelCase keys. Empty values are omitted so events stay clean and struct
// comparison in tests is stable.
func (e *ec2InstanceFetcher) buildAttributes(i *ec2.Ec2Instance, tags map[string]string) map[string]any {
	attrs := map[string]any{}
	if v := pointers.Deref(i.ImageId); v != "" {
		attrs["ImageId"] = v
	}
	if v := string(i.Platform); v != "" {
		attrs["Platform"] = v
	}
	if v := pointers.Deref(i.VpcId); v != "" {
		attrs["VpcId"] = v
	}
	if v := pointers.Deref(i.SubnetId); v != "" {
		attrs["SubnetId"] = v
	}
	if i.State != nil {
		if v := string(i.State.Name); v != "" {
			attrs["State"] = v
		}
	}
	if i.IamInstanceProfile != nil {
		if v := pointers.Deref(i.IamInstanceProfile.Arn); v != "" {
			attrs["RoleArn"] = v
		}
	}
	if v := awslib.LookupTag(tags, "owner"); v != "" {
		attrs["Owner"] = v
	}
	if v := awslib.LookupTag(tags, "costcenter", "cost-center", "cost_center"); v != "" {
		attrs["CostCenter"] = v
	}
	return attrs
}

// buildIPs collects non-empty IP address strings into a slice, returning nil when none exist.
// Using a nil (not empty) slice is important so that the json:"ip,omitempty" tag suppresses
// the field consistently and struct comparison in tests works without nil/empty mismatches.
func buildIPs(addrs ...*string) []string {
	var ips []string
	for _, addr := range addrs {
		if v := pointers.Deref(addr); v != "" {
			ips = append(ips, v)
		}
	}
	return ips
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
