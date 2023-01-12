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
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type PolicyDocument struct {
	PolicyName string `json:"PolicyName,omitempty"`
	Policy     string `json:"policy,omitempty"`
}

func (p Provider) listAttachedPolicies(ctx context.Context, identity *string) ([]types.AttachedPolicy, error) {
	input := &iamsdk.ListAttachedUserPoliciesInput{UserName: identity}
	var policies []types.AttachedPolicy
	for {
		output, err := p.client.ListAttachedUserPolicies(ctx, input)
		if err != nil {
			return nil, err
		}
		policies = append(policies, output.AttachedPolicies...)
		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	return policies, nil
}

func (p Provider) listInlinePolicies(ctx context.Context, identity *string) ([]PolicyDocument, error) {
	input := &iamsdk.ListUserPoliciesInput{
		UserName: identity,
	}
	var policyNames []string
	for {
		output, err := p.client.ListUserPolicies(ctx, input)
		if err != nil {
			return nil, err
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

	return policies, nil
}
