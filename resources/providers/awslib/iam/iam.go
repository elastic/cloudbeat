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
	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type AccessManagement interface {
	GetIAMRolePermissions(ctx context.Context, roleName string) ([]RolePolicyInfo, error)
	GetPasswordPolicy(ctx context.Context) (awslib.AwsResource, error)
	GetUsers(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log    *logp.Logger
	client IAMClient
}

type RolePolicyInfo struct {
	PolicyARN string
	iamsdk.GetRolePolicyOutput
}

type PasswordPolicy struct {
	ReusePreventionCount int
	RequireLowercase     bool
	RequireUppercase     bool
	RequireNumbers       bool
	RequireSymbols       bool
	MaxAgeDays           int
	MinimumLength        int
}

func NewIAMProvider(log *logp.Logger, cfg aws.Config) *Provider {
	svc := iamsdk.NewFromConfig(cfg)
	return &Provider{
		log:    log,
		client: svc,
	}
}

type IAMClient interface {
	ListUsers(ctx context.Context, params *iamsdk.ListUsersInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListUsersOutput, error)
	ListMFADevices(ctx context.Context, params *iamsdk.ListMFADevicesInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListMFADevicesOutput, error)
	ListAccessKeys(ctx context.Context, params *iamsdk.ListAccessKeysInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListAccessKeysOutput, error)
	GetAccessKeyLastUsed(ctx context.Context, params *iamsdk.GetAccessKeyLastUsedInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetAccessKeyLastUsedOutput, error)
	GetAccountPasswordPolicy(ctx context.Context, params *iamsdk.GetAccountPasswordPolicyInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetAccountPasswordPolicyOutput, error)
	GetRolePolicy(ctx context.Context, params *iamsdk.GetRolePolicyInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetRolePolicyOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iamsdk.ListAttachedRolePoliciesInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListAttachedRolePoliciesOutput, error)
}
