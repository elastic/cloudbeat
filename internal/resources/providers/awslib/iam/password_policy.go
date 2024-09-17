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
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

func (p Provider) GetPasswordPolicy(ctx context.Context) (awslib.AwsResource, error) {
	output, err := p.client.GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{})
	if err != nil {
		p.log.Debug("Failed to get account password policy: %v", err)

		// AWS error reference https://docs.aws.amazon.com/IAM/latest/APIReference/API_GetAccountPasswordPolicy.html
		var awsErr *types.NoSuchEntityException
		if errors.As(err, &awsErr) {
			// Reasoning behind this debug line https://github.com/elastic/cloudbeat/issues/1751
			p.log.Debug("Returning empty password policy because 'password policy not found' is not a valid response. The account has a default password policy")
			return &PasswordPolicy{}, nil
		}

		p.log.Infof("Debug the error %+v", err)

		return nil, err

	}

	policy := output.PasswordPolicy
	reusePrevention := 0
	if policy.PasswordReusePrevention != nil {
		reusePrevention = int(*policy.PasswordReusePrevention)
	}

	maxAge := 0
	if policy.MaxPasswordAge != nil {
		maxAge = int(*policy.MaxPasswordAge)
	}

	minimumLength := 0
	if policy.MinimumPasswordLength != nil {
		minimumLength = int(*policy.MinimumPasswordLength)
	}

	return &PasswordPolicy{
		ReusePreventionCount: reusePrevention,
		RequireLowercase:     policy.RequireLowercaseCharacters,
		RequireUppercase:     policy.RequireUppercaseCharacters,
		RequireNumbers:       policy.RequireNumbers,
		RequireSymbols:       policy.RequireSymbols,
		MaxAgeDays:           maxAge,
		MinimumLength:        minimumLength,
	}, nil
}

func (p PasswordPolicy) GetResourceArn() string {
	return ""
}

func (p PasswordPolicy) GetResourceName() string {
	return "account-password-policy"
}

func (p PasswordPolicy) GetResourceType() string {
	return fetching.PwdPolicyType
}

func (p PasswordPolicy) GetRegion() string {
	return awslib.GlobalRegion
}
