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

	"github.com/elastic/beats/v7/libbeat/ecs"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type iamPolicyFetcher struct {
	logger      *logp.Logger
	provider    iamPolicyProvider
	AccountId   string
	AccountName string
}

type iamPolicyProvider interface {
	GetPolicies(ctx context.Context) ([]awslib.AwsResource, error)
}

func newIamPolicyFetcher(logger *logp.Logger, identity *cloud.Identity, provider iamPolicyProvider) inventory.AssetFetcher {
	return &iamPolicyFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (i *iamPolicyFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	i.logger.Info("Fetching IAM Policies")
	defer i.logger.Info("Fetching IAM Policies - Finished")

	policies, err := i.provider.GetPolicies(ctx)
	if err != nil {
		i.logger.Errorf("Could not list policies: %v", err)
		if len(policies) == 0 {
			return
		}
	}

	for _, resource := range policies {
		if resource == nil {
			continue
		}

		policy, ok := resource.(iam.Policy)
		if !ok {
			i.logger.Errorf("Could not get info about policy: %s", resource.GetResourceArn())
			continue
		}

		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamPolicy,
			policy.GetResourceArn(),
			resource.GetResourceName(),

			inventory.WithRawAsset(policy),
			inventory.WithLabels(i.getTags(policy)),
			inventory.WithCloud(ecs.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      awslib.GlobalRegion,
				AccountID:   i.AccountId,
				AccountName: i.AccountName,
				ServiceName: "AWS IAM",
			}),
		)
	}
}

func (i *iamPolicyFetcher) getTags(policy iam.Policy) map[string]string {
	tags := make(map[string]string, len(policy.Tags))

	for _, tag := range policy.Tags {
		tags[pointers.Deref(tag.Key)] = pointers.Deref(tag.Value)
	}

	return tags
}
