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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
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

// GCPCloudConnectorsParams holds GCP-specific parameters for the Cloud Connectors auth flow.
type GCPCloudConnectorsParams struct {
	Audience            string // Workload Identity Federation audience URL
	ServiceAccountEmail string // Target service account to impersonate
	CloudConnectorID    string // Deployment connector ID (Terraform output); session name = ResourceID-CloudConnectorID
}

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
func (p *GoogleAuthProvider) FindCloudConnectorsCredentials(ctx context.Context, ccConfig config.CloudConnectorsConfig, params GCPCloudConnectorsParams) ([]option.ClientOption, error) {
	// Validate required configuration
	if ccConfig.JWTFilePath == "" {
		return nil, errors.New("cloud connectors config JWTFilePath is required")
	}

	if ccConfig.GlobalRoleARN == "" {
		return nil, errors.New("cloud connectors config GlobalRoleARN is required")
	}

	if ccConfig.ResourceID == "" {
		return nil, errors.New("cloud connectors config ResourceID is required")
	}

	if params.CloudConnectorID == "" {
		return nil, errors.New("cloud connectors config CloudConnectorID is required")
	}

	// Session name must match GCP Workload Identity Federation: elastic_resource_id-cloud_connector_id
	sessionName := ccConfig.ResourceID + "-" + params.CloudConnectorID

	// Create STS client and credentials cache at initialization (like role chaining)
	stsClient := sts.New(sts.Options{Region: defaultAWSRegion})
	credsCache := awslib.NewWebIdentityCredentialsCache(
		stsClient,
		ccConfig.GlobalRoleARN,
		ccConfig.JWTFilePath,
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = sessionName
		},
	)

	// Create the AWS credentials supplier with the pre-initialized cache
	credSupplier := &awsCredentialsSupplier{
		region:     defaultAWSRegion,
		credsCache: credsCache,
	}

	cfg := externalaccount.Config{
		Audience:                       params.Audience,
		SubjectTokenType:               awsTokenType,
		TokenURL:                       gcpSTSTokenURL,
		Scopes:                         []string{gcpCloudPlatformScope},
		AwsSecurityCredentialsSupplier: credSupplier,
		ServiceAccountImpersonationURL: gcpIAMCredentialsURL + params.ServiceAccountEmail + ":generateAccessToken",
	}

	tokenSource, err := externalaccount.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create external account token source: %w", err)
	}

	return []option.ClientOption{option.WithTokenSource(tokenSource)}, nil
}

// awsCredentialsSupplier implements externalaccount.AwsSecurityCredentialsSupplier
// It provides cached AWS credentials to GCP for Workload Identity Federation.
// The credentials cache is initialized once and automatically refreshes when expired.
type awsCredentialsSupplier struct {
	region     string
	credsCache *aws.CredentialsCache
}

// AwsRegion returns the AWS region for the credentials.
func (s *awsCredentialsSupplier) AwsRegion(_ context.Context, _ externalaccount.SupplierOptions) (string, error) {
	return s.region, nil
}

// AwsSecurityCredentials retrieves cached AWS credentials for GCP WIF.
// The cache automatically refreshes credentials when they expire.
func (s *awsCredentialsSupplier) AwsSecurityCredentials(ctx context.Context, _ externalaccount.SupplierOptions) (*externalaccount.AwsSecurityCredentials, error) {
	creds, err := s.credsCache.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AWS credentials: %w", err)
	}

	return &externalaccount.AwsSecurityCredentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
	}, nil
}
