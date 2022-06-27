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

package awslib

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/elastic/elastic-agent-libs/logp"
)

type IAMRolePermissionGetter interface {
	GetIAMRolePermissions(ctx context.Context, roleName string) ([]iam.GetRolePolicyResponse, error)
}

type IAMProvider struct {
	log    *logp.Logger
	client *iam.Client
}

func NewIAMProvider(log *logp.Logger, cfg aws.Config) *IAMProvider {
	svc := iam.New(cfg)
	return &IAMProvider{
		log:    log,
		client: svc,
	}
}

func (p IAMProvider) GetIAMRolePermissions(ctx context.Context, roleName string) ([]iam.GetRolePolicyResponse, error) {
	results := make([]iam.GetRolePolicyResponse, 0)
	policiesIdentifiers, err := p.getAllRolePolicies(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to list role %s policies - %w", roleName, err)
	}

	for _, policyId := range policiesIdentifiers {
		input := &iam.GetRolePolicyInput{
			PolicyName: policyId.PolicyName,
			RoleName:   &roleName,
		}
		req := p.client.GetRolePolicyRequest(input)
		policy, err := req.Send(ctx)
		if err != nil {
			p.log.Errorf("Failed to get policy %s: %v", *policyId.PolicyName, err)
			continue
		}
		results = append(results, *policy)
	}

	return results, nil
}

func (p IAMProvider) getAllRolePolicies(ctx context.Context, roleName string) ([]iam.AttachedPolicy, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	}
	req := p.client.ListAttachedRolePoliciesRequest(input)
	allPolicies, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list role %s policies - %w", roleName, err)
	}

	return allPolicies.AttachedPolicies, err
}
