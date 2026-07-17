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
	"strings"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

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
	iamResolver   instanceProfileResolver
	AccountId     string
	AccountName   string
	statusHandler statushandler.StatusHandlerAPI
}

type ec2InstancesProvider interface {
	DescribeInstances(ctx context.Context) ([]*ec2.Ec2Instance, error)
}

// instanceProfileResolver resolves an IAM instance profile by name, returning
// the profile object (which contains the list of associated roles).
type instanceProfileResolver interface {
	GetInstanceProfile(ctx context.Context, instanceProfileName string) (*iamtypes.InstanceProfile, error)
}

func newEc2InstancesFetcher(logger *clog.Logger, identity *cloud.Identity, provider ec2InstancesProvider, iamResolver instanceProfileResolver, statusHandler statushandler.StatusHandlerAPI) inventory.AssetFetcher {
	return &ec2InstanceFetcher{
		logger:        logger,
		provider:      provider,
		iamResolver:   iamResolver,
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

	// Cache resolved role ARNs per profile ARN to avoid duplicate IAM calls.
	roleArnCache := make(map[string]string)

	for _, i := range instances {
		if i == nil {
			continue
		}

		tags := e.getTags(i)

		// Resolve the IAM role ARN from the instance profile.
		// IamInstanceProfile.Arn is the *instance-profile* ARN, not the role ARN.
		// We emit InstanceProfileArn (accurate, free) and RoleArn (resolved via GetInstanceProfile).
		iamFetcher := inventory.EmptyEnricher()
		if i.IamInstanceProfile != nil {
			profileArn := pointers.Deref(i.IamInstanceProfile.Arn)
			if profileArn != "" {
				roleArn := e.resolveRoleArn(ctx, profileArn, roleArnCache)
				// WithUser links this instance to the IAM Role asset (or falls back to the profile ARN).
				userID := roleArn
				if userID == "" {
					userID = profileArn
				}
				iamFetcher = inventory.WithUser(inventory.User{
					ID: userID,
				})
			}
		}

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
			inventory.WithEntityDetails(e.buildDetails(i, tags, roleArnCache)),
			inventory.WithCreatedAt(i.LaunchTime),
			iamFetcher,
		)
	}
}

// resolveRoleArn attempts to resolve the IAM role ARN for the given instance-profile ARN.
// Results are cached in roleArnCache (keyed by profile ARN) so instances sharing a profile
// trigger only one IAM call. Returns "" if the profile has no roles or the call fails.
func (e *ec2InstanceFetcher) resolveRoleArn(ctx context.Context, profileArn string, roleArnCache map[string]string) string {
	if cached, ok := roleArnCache[profileArn]; ok {
		return cached
	}

	profileName := profileNameFromArn(profileArn)
	if profileName == "" {
		roleArnCache[profileArn] = ""
		return ""
	}

	profile, err := e.iamResolver.GetInstanceProfile(ctx, profileName)
	if err != nil {
		e.logger.Warnf("Could not resolve IAM role for instance profile %s: %v", profileArn, err)
		roleArnCache[profileArn] = ""
		return ""
	}

	roleArn := ""
	if len(profile.Roles) > 0 {
		roleArn = pointers.Deref(profile.Roles[0].Arn)
	}

	roleArnCache[profileArn] = roleArn
	return roleArn
}

// profileNameFromArn extracts the instance-profile name from its ARN.
// Example: "arn:aws:iam::123:instance-profile/MyProfile" → "MyProfile"
// Also handles path-qualified names: ".../instance-profile/division/MyProfile" → "division/MyProfile".
func profileNameFromArn(arn string) string {
	const marker = "instance-profile/"
	idx := strings.Index(arn, marker)
	if idx == -1 {
		return ""
	}
	return arn[idx+len(marker):]
}

// buildDetails collects non-ECS, resource-specific EC2 fields into entity.Details,
// using UpperCamelCase keys. Empty values are omitted so events stay clean and struct
// comparison in tests is stable.
func (e *ec2InstanceFetcher) buildDetails(i *ec2.Ec2Instance, tags map[string]string, roleArnCache map[string]string) map[string]any {
	details := map[string]any{}
	if v := pointers.Deref(i.ImageId); v != "" {
		details["ImageId"] = v
	}
	if v := string(i.Platform); v != "" {
		details["Platform"] = v
	}
	if v := pointers.Deref(i.VpcId); v != "" {
		details["VpcId"] = v
	}
	if v := pointers.Deref(i.SubnetId); v != "" {
		details["SubnetId"] = v
	}
	if i.State != nil {
		if v := string(i.State.Name); v != "" {
			details["State"] = v
		}
	}
	if i.IamInstanceProfile != nil {
		if profileArn := pointers.Deref(i.IamInstanceProfile.Arn); profileArn != "" {
			details["InstanceProfileArn"] = profileArn
			if roleArn, ok := roleArnCache[profileArn]; ok && roleArn != "" {
				details["RoleArn"] = roleArn
			}
		}
	}
	if v := awslib.LookupTag(tags, "owner"); v != "" {
		details["Owner"] = v
	}
	if v := awslib.LookupTag(tags, "costcenter", "cost-center", "cost_center"); v != "" {
		details["CostCenter"] = v
	}
	if v := awslib.LookupTag(tags, "role"); v != "" {
		details["Role"] = v
	}
	return details
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
