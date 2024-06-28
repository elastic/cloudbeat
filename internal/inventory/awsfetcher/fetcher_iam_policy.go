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

var iamPolicyClassification = inventory.AssetClassification{
	Category:    inventory.CategoryIdentity,
	SubCategory: inventory.SubCategoryCloudProviderAccount,
	Type:        inventory.TypePermissions,
	SubType:     inventory.SubTypeIAM,
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
			iamPolicyClassification,
			inventory.Identifiers(
				inventory.Arns(policy.GetResourceArn()),
				inventory.Ids(pointers.Deref(policy.PolicyId)),
			),
			resource.GetResourceName(),

			inventory.WithRawAsset(policy),
			inventory.WithResourcePolicies(convertPolicy(policy.Document)...),
			inventory.WithTags(i.getTags(policy)),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   awslib.GlobalRegion,
				Account: inventory.AssetCloudAccount{
					Id:   i.AccountId,
					Name: i.AccountName,
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS IAM",
				},
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

func convertPolicy(policy map[string]any) []inventory.AssetResourcePolicy {
	if len(policy) == 0 {
		return nil
	}

	version, hasVersion := policy["Version"].(string)
	if !hasVersion {
		version = ""
	}

	switch statements := policy["Statement"].(type) {
	case []map[string]any:
		return convertStatements(statements, version)
	case []any:
		return convertAnyStatements(statements, version)
	case map[string]any:
		return []inventory.AssetResourcePolicy{convertStatement(statements, &version)}
	}
	return nil
}

func convertAnyStatements(statements []any, version string) []inventory.AssetResourcePolicy {
	policies := make([]inventory.AssetResourcePolicy, 0, len(statements))
	for _, statement := range statements {
		policies = append(policies, convertStatement(statement.(map[string]any), &version))
	}
	return policies
}

func convertStatements(statements []map[string]any, version string) []inventory.AssetResourcePolicy {
	policies := make([]inventory.AssetResourcePolicy, 0, len(statements))
	for _, statement := range statements {
		policies = append(policies, convertStatement(statement, &version))
	}
	return policies
}

func convertStatement(statement map[string]any, version *string) inventory.AssetResourcePolicy {
	p := inventory.AssetResourcePolicy{}
	p.Version = version

	if sid, ok := statement["Sid"]; ok {
		p.Id = pointers.Ref(sid.(string))
	}

	if effect, ok := statement["Effect"]; ok {
		p.Effect = effect.(string)
	}

	if anyPrincipal, ok := statement["Principal"]; ok {
		switch principal := anyPrincipal.(type) {
		case string:
			p.Principal = map[string]any{principal: principal}
		case map[string]any:
			p.Principal = principal
		}
	}

	if action, ok := statement["Action"]; ok {
		p.Action = anyToSliceString(action)
	}

	if notAction, ok := statement["NotAction"]; ok {
		p.NotAction = anyToSliceString(notAction)
	}

	if resource, ok := statement["Resource"]; ok {
		p.Resource = anyToSliceString(resource)
	}

	if noResource, ok := statement["NoResource"]; ok {
		p.NoResource = anyToSliceString(noResource)
	}

	if condition, ok := statement["Condition"]; ok {
		p.Condition = condition.(map[string]any)
	}

	return p
}

func anyToSliceString(anyString any) []string {
	switch s := anyString.(type) {
	case string:
		return []string{s}
	case []string:
		return s
	}

	return nil
}
