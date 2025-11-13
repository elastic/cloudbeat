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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AWSConfigRoleChaining initializes an assume role provider and a credentials cache for each step in the chain,
// using the previous step's credentials as the client for the next step.
func AWSConfigRoleChaining(initialConfig aws.Config, chain []AWSRoleChainingStep) *aws.Config {
	cnf := initialConfig

	for _, step := range chain {
		client := sts.NewFromConfig(cnf)
		credentialsCache := step.BuildCredentialsCache(client)
		cnf.Credentials = credentialsCache
	}

	return &cnf
}

// AWSRoleChainingStep represents a single step in an AWS role assumption chain.
// Implementations should provide a way to build AWS credentials from an STS client.
type AWSRoleChainingStep interface {
	// BuildCredentialsCache creates a credentials cache using the provided STS client
	BuildCredentialsCache(client *sts.Client) *aws.CredentialsCache
}

// AssumeRoleStep represents a standard AssumeRole operation in the chain.
// This is used for assuming roles with long-term credentials or previously assumed role credentials.
type AssumeRoleStep struct {
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// Options configures the AssumeRole operation (session name, duration, external ID, etc.)
	Options func(aro *stscreds.AssumeRoleOptions)
}

// BuildCredentialsCache implements AWSRoleChainingStep for AssumeRole operations.
func (s *AssumeRoleStep) BuildCredentialsCache(client *sts.Client) *aws.CredentialsCache {
	assumeRoleProvider := stscreds.NewAssumeRoleProvider(
		client,
		s.RoleARN,
		s.Options,
	)
	return aws.NewCredentialsCache(assumeRoleProvider)
}

// WebIdentityRoleStep represents an AssumeRoleWithWebIdentity operation in the chain.
// This is used for OIDC/JWT-based authentication.
type WebIdentityRoleStep struct {
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// WebIdentityTokenFile is the path to a file containing the JWT/OIDC token
	WebIdentityTokenFile string
	// Options configures the AssumeRoleWithWebIdentity operation (session name, duration, etc.)
	Options func(o *stscreds.WebIdentityRoleOptions)
}

// BuildCredentialsCache implements AWSRoleChainingStep for AssumeRoleWithWebIdentity operations.
func (s *WebIdentityRoleStep) BuildCredentialsCache(client *sts.Client) *aws.CredentialsCache {
	tokenRetriever := NewFileTokenRetriever(s.WebIdentityTokenFile)
	webIdentityProvider := stscreds.NewWebIdentityRoleProvider(
		client,
		s.RoleARN,
		tokenRetriever,
		s.Options,
	)
	return aws.NewCredentialsCache(webIdentityProvider)
}

// Compile-time checks to ensure types implement AWSRoleChainingStep
var (
	_ AWSRoleChainingStep = (*AssumeRoleStep)(nil)
	_ AWSRoleChainingStep = (*WebIdentityRoleStep)(nil)
)
