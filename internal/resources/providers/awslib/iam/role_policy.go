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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (p Provider) GetIAMRolePermissions(ctx context.Context, roleName string) ([]RolePolicyInfo, error) {
	results := make([]RolePolicyInfo, 0)
	policiesIdentifiers, err := p.getAllRolePolicies(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to list role %s policies - %w", roleName, err)
	}

	for _, policyId := range policiesIdentifiers {
		policyArn := policyId.PolicyArn
		input := &iam.GetRolePolicyInput{
			PolicyName: policyId.PolicyName,
			RoleName:   &roleName,
		}

		policy, err := p.client.GetRolePolicy(ctx, input)
		if err != nil {
			p.log.Errorf(ctx, "Failed to get policy %s: %v", *policyId.PolicyName, err)
			continue
		}

		results = append(results, RolePolicyInfo{
			PolicyARN:           *policyArn,
			GetRolePolicyOutput: *policy,
		})
	}

	return results, nil
}

func (p Provider) getAllRolePolicies(ctx context.Context, roleName string) ([]types.AttachedPolicy, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	}
	allPolicies, err := p.client.ListAttachedRolePolicies(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list role %s policies - %w", roleName, err)
	}

	return allPolicies.AttachedPolicies, err
}

func (r RolePolicyInfo) GetResourceArn() string {
	return ""
}

func (r RolePolicyInfo) GetResourceName() string {
	return ""
}

func (r RolePolicyInfo) GetResourceType() string {
	return ""
}
