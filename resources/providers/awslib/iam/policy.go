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
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"net/url"
)

func (p Provider) GetPolicies(ctx context.Context) ([]awslib.AwsResource, error) {
	var policies []awslib.AwsResource

	input := &iamsdk.ListPoliciesInput{OnlyAttached: true}
	for {
		listPoliciesOutput, err := p.client.ListPolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, policy := range listPoliciesOutput.Policies {
			output, err := p.getPolicyVersion(ctx, policy)
			if err != nil {
				return nil, err
			}

			doc, err := decodePolicyDocument(output.PolicyVersion)
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

func (p Provider) GetSupportPolicy(ctx context.Context) (awslib.AwsResource, error) {
	const awsSupportAccessArn = "arn:aws:iam::aws:policy/AWSSupportAccess"

	policy, err := p.client.GetPolicy(ctx, &iamsdk.GetPolicyInput{PolicyArn: aws.String(awsSupportAccessArn)})
	if err != nil {
		return nil, err
	}

	awsSupportAccessPolicy := Policy{
		Policy: *policy.Policy,
		Roles:  make([]types.PolicyRole, 0),
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

func (p Provider) getPolicyVersion(ctx context.Context, policy types.Policy) (*iamsdk.GetPolicyVersionOutput, error) {
	if policy.Arn == nil || policy.DefaultVersionId == nil {
		return nil, fmt.Errorf("invalid policy: %v", policy)
	}
	return p.client.GetPolicyVersion(ctx, &iamsdk.GetPolicyVersionInput{PolicyArn: policy.Arn, VersionId: policy.DefaultVersionId})
}

func decodePolicyDocument(policyVersion *types.PolicyVersion) (map[string]interface{}, error) {
	if policyVersion == nil || policyVersion.Document == nil {
		return nil, fmt.Errorf("invalid policy version: %v", policyVersion)
	}

	// The policy document is URL-encoded, compliant with RFC 3986
	docString, err := url.QueryUnescape(*policyVersion.Document)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape policy document: %w", err)
	}

	var doc map[string]interface{}
	err = json.Unmarshal([]byte(docString), &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy document: %w", err)
	}

	return doc, nil
}

func (p Policy) GetResourceArn() string {
	return stringOrEmpty(p.Arn)
}

func (p Policy) GetResourceName() string {
	return stringOrEmpty(p.PolicyName)
}

func (p Policy) GetResourceType() string {
	return fetching.PolicyType
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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

	var policies []PolicyDocument
	for i := range policyNames {
		inlinePolicy, err := p.client.GetUserPolicy(ctx, &iamsdk.GetUserPolicyInput{
			PolicyName: &policyNames[i],
			UserName:   identity,
		})

		if err != nil {
			p.log.Errorf("fail to get inline policy for user: %s, policy name: %s", *identity, policyNames[i])
			policies = append(policies, PolicyDocument{PolicyName: policyNames[i]})
			continue
		}

		policies = append(policies, PolicyDocument{PolicyName: policyNames[i], Policy: *inlinePolicy.PolicyDocument})
	}

	p.log.Debugf("inline policies for user: %s, policies: %v", *identity, policies)
	return policies, nil
}
