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

package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

const (
	// GCP Security Token Service endpoint for token exchange
	gcpSTSTokenURL = "https://sts.googleapis.com/v1/token"
	// GCP IAM Credentials API endpoint for service account impersonation
	gcpIAMCredentialsURL = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/"
	// Token type for AWS-based authentication
	awsTokenType = "urn:ietf:params:aws:token-type:aws4_request"
	// Default scope for GCP Cloud Platform access
	gcpCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
	// Default AWS region for STS operations
	defaultAWSRegion = "us-east-1"
)

type GoogleAuthProvider struct{}

// FindDefaultCredentials is a wrapper around google.FindDefaultCredentials to make it easier to mock
func (p *GoogleAuthProvider) FindDefaultCredentials(ctx context.Context) (*google.Credentials, error) {
	return google.FindDefaultCredentials(ctx)
}

// FindCloudConnectorsCredentials creates GCP client options using AWS Workload Identity Federation
// with direct service account impersonation.
//
// The authentication flow:
// 1. Reads JWT from file (ccConfig.JWTFilePath)
// 2. Assumes Elastic's AWS role using AssumeRoleWithWebIdentity
// 3. Uses AWS credentials for GCP Workload Identity Federation token exchange
// 4. Impersonates the target service account in the customer's GCP project
func (p *GoogleAuthProvider) FindCloudConnectorsCredentials(ctx context.Context, ccConfig config.CloudConnectorsConfig, audience string, serviceAccountEmail string) ([]option.ClientOption, error) {
	// Validate required configuration
	if ccConfig.JWTFilePath == "" {
		return nil, fmt.Errorf("cloud connectors config JWTFilePath is required")
	}

	if ccConfig.GlobalRoleARN == "" {
		return nil, fmt.Errorf("cloud connectors config GlobalRoleARN is required")
	}

	if ccConfig.ResourceID == "" {
		return nil, fmt.Errorf("cloud connectors config ResourceID is required")
	}

	// Create the AWS credentials supplier that handles the JWT -> AWS role assumption
	credSupplier := &awsCredentialsSupplier{
		jwtFilePath:   ccConfig.JWTFilePath,
		globalRoleARN: ccConfig.GlobalRoleARN,
		roleSessionID: ccConfig.ResourceID,
		region:        defaultAWSRegion,
	}

	cfg := externalaccount.Config{
		Audience:                       audience,
		SubjectTokenType:               awsTokenType,
		TokenURL:                       gcpSTSTokenURL,
		Scopes:                         []string{gcpCloudPlatformScope},
		AwsSecurityCredentialsSupplier: credSupplier,
		ServiceAccountImpersonationURL: gcpIAMCredentialsURL + serviceAccountEmail + ":generateAccessToken",
	}

	tokenSource, err := externalaccount.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create external account token source: %w", err)
	}

	return []option.ClientOption{option.WithTokenSource(tokenSource)}, nil
}

// awsCredentialsSupplier implements externalaccount.AwsSecurityCredentialsSupplier
// It assumes an AWS role using a JWT token and provides the resulting credentials to GCP.
type awsCredentialsSupplier struct {
	jwtFilePath   string
	globalRoleARN string
	roleSessionID string
	region        string
}

// AwsRegion returns the AWS region for the credentials.
func (s *awsCredentialsSupplier) AwsRegion(ctx context.Context, options externalaccount.SupplierOptions) (string, error) {
	return s.region, nil
}

// AwsSecurityCredentials assumes the AWS role using the JWT and returns the temporary credentials.
func (s *awsCredentialsSupplier) AwsSecurityCredentials(ctx context.Context, options externalaccount.SupplierOptions) (*externalaccount.AwsSecurityCredentials, error) {
	// Create STS client without credentials (we're using web identity)
	stsClient := sts.New(sts.Options{
		Region: s.region,
	})

	// Use the AWS SDK's built-in web identity provider
	webIdentityProvider := stscreds.NewWebIdentityRoleProvider(
		stsClient,
		s.globalRoleARN,
		stscreds.IdentityTokenFile(s.jwtFilePath),
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = s.roleSessionID
		},
	)

	// Retrieve credentials using the web identity provider
	creds, err := webIdentityProvider.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role %s with web identity: %w", s.globalRoleARN, err)
	}

	return &externalaccount.AwsSecurityCredentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
	}, nil
}
