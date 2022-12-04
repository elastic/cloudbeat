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
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/cloudbeat/resources/fetching"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

func (p Provider) GetPasswordPolicy(ctx context.Context) (awslib.AwsResource, error) {
	output, err := p.client.GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{})
	if err != nil {
		p.log.Debug("Failed to get account password policy: %v", err)
		return PasswordPolicy{}, err
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

	return PasswordPolicy{
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
