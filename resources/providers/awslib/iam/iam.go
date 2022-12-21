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
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type AccessManagement interface {
	GetIAMRolePermissions(ctx context.Context, roleName string) ([]RolePolicyInfo, error)
	GetPasswordPolicy(ctx context.Context) (awslib.AwsResource, error)
	GetUsers(ctx context.Context) ([]awslib.AwsResource, error)
}

type Client interface {
	ListUsers(ctx context.Context, params *iamsdk.ListUsersInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListUsersOutput, error)
	ListMFADevices(ctx context.Context, params *iamsdk.ListMFADevicesInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListMFADevicesOutput, error)
	ListAccessKeys(ctx context.Context, params *iamsdk.ListAccessKeysInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListAccessKeysOutput, error)
	GetAccessKeyLastUsed(ctx context.Context, params *iamsdk.GetAccessKeyLastUsedInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetAccessKeyLastUsedOutput, error)
	GetAccountPasswordPolicy(ctx context.Context, params *iamsdk.GetAccountPasswordPolicyInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetAccountPasswordPolicyOutput, error)
	GetRolePolicy(ctx context.Context, params *iamsdk.GetRolePolicyInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetRolePolicyOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iamsdk.ListAttachedRolePoliciesInput, optFns ...func(*iamsdk.Options)) (*iamsdk.ListAttachedRolePoliciesOutput, error)
	GenerateCredentialReport(ctx context.Context, params *iamsdk.GenerateCredentialReportInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GenerateCredentialReportOutput, error)
	GetCredentialReport(ctx context.Context, params *iamsdk.GetCredentialReportInput, optFns ...func(*iamsdk.Options)) (*iamsdk.GetCredentialReportOutput, error)
}

type Provider struct {
	log    *logp.Logger
	client Client
}

type RolePolicyInfo struct {
	PolicyARN string
	iamsdk.GetRolePolicyOutput
}

// User Override SDK User type
type User struct {
	Name                string       `json:"name,omitempty"`
	AccessKeys          []AccessKey  `json:"access_keys,omitempty"`
	MFADevices          []AuthDevice `json:"mfa_devices,omitempty"`
	LastAccess          string       `json:"last_access,omitempty"`
	Arn                 string       `json:"arn,omitempty"`
	PasswordEnabled     string       `json:"password_enabled,omitempty"`
	PasswordLastChanged string       `json:"password_last_changed,omitempty"`
	MfaActive           string       `json:"mfa_active,omitempty"`
}

type AuthDevice struct {
	IsVirtual bool `json:"is_virtual,omitempty"`
	types.MFADevice
}

type AccessKey struct {
	AccessKeyId  string `json:"access_key_id,omitempty"`
	Active       bool   `json:"active,omitempty"`
	CreationDate string `json:"creation_date,omitempty"`
	LastAccess   string `json:"last_access,omitempty"`
	HasUsed      bool   `json:"has_used,omitempty"`
}

type PasswordPolicy struct {
	ReusePreventionCount int  `json:"reuse_prevention_count,omitempty"`
	RequireLowercase     bool `json:"require_lowercase,omitempty"`
	RequireUppercase     bool `json:"require_uppercase,omitempty"`
	RequireNumbers       bool `json:"require_numbers,omitempty"`
	RequireSymbols       bool `json:"require_symbols,omitempty"`
	MaxAgeDays           int  `json:"max_age_days,omitempty"`
	MinimumLength        int  `json:"minimum_length,omitempty"`
}

// CredentialReport credential report CSV output
type CredentialReport struct {
	User                  string `csv:"user"`
	Arn                   string `csv:"arn"`
	UserCreation          string `csv:"user_creation_time"`
	PasswordEnabled       string `csv:"password_enabled"`
	PasswordLastUsed      string `csv:"password_last_used"`
	PasswordLastChanged   string `csv:"password_last_changed"`
	PasswordNextRotation  string `csv:"password_next_rotation"`
	MfaActive             string `csv:"mfa_active"`
	AccessKey1Active      string `csv:"access_key_1_active"`
	AccessKey1LastRotated string `csv:"access_key_1_last_rotated"`
	AccessKey2Active      string `csv:"access_key_2_active"`
	AccessKey2LastRotated string `csv:"access_key_2_last_rotated"`
	Cert1Active           string `csv:"cert_1_active"`
	Cert2Active           string `csv:"cert_2_active"`
}

func NewIAMProvider(log *logp.Logger, cfg aws.Config) *Provider {
	svc := iamsdk.NewFromConfig(cfg)
	return &Provider{
		log:    log,
		client: svc,
	}
}
