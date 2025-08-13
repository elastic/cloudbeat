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

package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const awsSupportAccessArn = "arn:aws:iam::aws:policy/AWSSupportAccess"

func (p Provider) GetPolicies(ctx context.Context) ([]awslib.AwsResource, error) {
	policies, err := p.getPolicies(ctx)
	if err != nil {
		return nil, err
	}
	supportPolicy, err := p.getSupportPolicy(ctx)
	if err != nil {
		return nil, err
	}
	return append(policies, supportPolicy), nil
}

func (p Provider) getPolicies(ctx context.Context) ([]awslib.AwsResource, error) {
	var policies []awslib.AwsResource

	input := &iamsdk.ListPoliciesInput{OnlyAttached: true}
	for {
		listPoliciesOutput, err := p.client.ListPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, policy := range listPoliciesOutput.Policies {
			if pointers.Deref(policy.Arn) == awsSupportAccessArn {
				// Fetch this one explicitly with getSupportPolicy().
				// The reasoning is that we want to attach roles to the AWS support access policy. If we don't skip it
				// here, we will produce it another time in getSupportPolicy(), leading to duplicated resources. We
				// cannot just fetch the roles here either because if the AWS support access policy is not attached,
				// we will never see it.
				// See: https://github.com/elastic/cloudbeat/pull/900
				continue
			}

			doc, err := p.getPolicyDocument(ctx, policy)
			if err != nil {
				return nil, err
			}

			policies = append(policies, Policy{
				Policy:   policy,
				Document: doc,
			})
		}

		if !listPoliciesOutput.IsTruncated {
			break
		}
		input.Marker = listPoliciesOutput.Marker
	}

	return policies, nil
}

func (p Provider) getSupportPolicy(ctx context.Context) (awslib.AwsResource, error) {
	policy, err := p.client.GetPolicy(ctx, &iamsdk.GetPolicyInput{PolicyArn: aws.String(awsSupportAccessArn)})
	if err != nil {
		return nil, err
	}

	doc, err := p.getPolicyDocument(ctx, *policy.Policy)
	if err != nil {
		return nil, err
	}
	awsSupportAccessPolicy := Policy{
		Policy:   *policy.Policy,
		Document: doc,
		Roles:    make([]types.PolicyRole, 0),
	}

	input := &iamsdk.ListEntitiesForPolicyInput{
		PolicyArn:    aws.String(awsSupportAccessArn),
		EntityFilter: types.EntityTypeRole,
	}
	for {
		output, err := p.client.ListEntitiesForPolicy(ctx, input)
		if err != nil {
			return nil, err
		}

		awsSupportAccessPolicy.Roles = append(awsSupportAccessPolicy.Roles, output.PolicyRoles...)

		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	return awsSupportAccessPolicy, nil
}

func (p Provider) getPolicyDocument(ctx context.Context, policy types.Policy) (map[string]any, error) {
	if policy.Arn == nil || policy.DefaultVersionId == nil {
		return nil, fmt.Errorf("invalid policy: %v", policy)
	}
	out, err := p.client.GetPolicyVersion(ctx, &iamsdk.GetPolicyVersionInput{PolicyArn: policy.Arn, VersionId: policy.DefaultVersionId})
	if err != nil {
		return nil, err
	}

	doc, err := decodePolicyDocument(out.PolicyVersion)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func decodePolicyDocument(policyVersion *types.PolicyVersion) (map[string]any, error) {
	if policyVersion == nil || policyVersion.Document == nil {
		return nil, fmt.Errorf("invalid policy version: %v", policyVersion)
	}

	// The policy document is URL-encoded, compliant with RFC 3986
	docString, err := url.QueryUnescape(*policyVersion.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape policy document: %w", err)
	}

	var doc map[string]any
	err = json.Unmarshal([]byte(docString), &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy document: %w", err)
	}

	return doc, nil
}

func (p Policy) GetResourceArn() string {
	return pointers.Deref(p.Arn)
}

func (p Policy) GetResourceName() string {
	return pointers.Deref(p.PolicyName)
}

func (p Policy) GetResourceType() string {
	return fetching.PolicyType
}

func (p Policy) GetRegion() string {
	return awslib.GlobalRegion
}

func (p Provider) listAttachedPolicies(ctx context.Context, identity *string) ([]types.AttachedPolicy, error) {
	p.log.Debugf("listAttachedPolicies for user: %s", *identity)
	input := &iamsdk.ListAttachedUserPoliciesInput{UserName: identity}
	var policies []types.AttachedPolicy
	for {
		output, err := p.client.ListAttachedUserPolicies(ctx, input)
		if err != nil {
			return []types.AttachedPolicy{}, err
		}
		policies = append(policies, output.AttachedPolicies...)
		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	p.log.Debugf("attached policies for user: %s, policies: %v", *identity, policies)
	return policies, nil
}

func (p Provider) listInlinePolicies(ctx context.Context, identity *string) ([]PolicyDocument, error) {
	p.log.Debugf("listInlinePolicies for user: %s", *identity)

	input := &iamsdk.ListUserPoliciesInput{
		UserName: identity,
	}
	var policyNames []string
	for {
		output, err := p.client.ListUserPolicies(ctx, input)
		if err != nil {
			return []PolicyDocument{}, err
		}
		policyNames = append(policyNames, output.PolicyNames...)
		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	policies := make([]PolicyDocument, 0, len(policyNames))
	for i := range policyNames {
		inlinePolicy, err := p.client.GetUserPolicy(ctx, &iamsdk.GetUserPolicyInput{
			PolicyName: &policyNames[i],
			UserName:   identity,
		})
		if err != nil {
			p.log.Errorf(ctx, "fail to get inline policy for user: %s, policy name: %s", *identity, policyNames[i])
			policies = append(policies, PolicyDocument{PolicyName: policyNames[i]})
			continue
		}

		policies = append(policies, PolicyDocument{PolicyName: policyNames[i], Policy: *inlinePolicy.PolicyDocument})
	}

	p.log.Debugf("inline policies for user: %s, policies: %v", *identity, policies)
	return policies, nil
}
