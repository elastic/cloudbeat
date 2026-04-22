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

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	libbeatgcp "github.com/elastic/beats/v7/x-pack/libbeat/common/identityfederation/gcp"

	"github.com/elastic/cloudbeat/internal/config"
)

// GCPIdentityFederationParams holds GCP-specific parameters for the Identity Federation auth flow.
type GCPIdentityFederationParams struct {
	Audience            string // Workload Identity Federation audience URL
	ServiceAccountEmail string // Target service account to impersonate
	IdentityFederationID string // Deployment connector ID (Terraform output cloud_connector_id); session name = ResourceID-IdentityFederationID
}

type GoogleAuthProvider struct{}

// FindDefaultCredentials is a wrapper around google.FindDefaultCredentials to make it easier to mock
func (p *GoogleAuthProvider) FindDefaultCredentials(ctx context.Context) (*google.Credentials, error) {
	return google.FindDefaultCredentials(ctx)
}

// FindIdentityFederationCredentials creates GCP client options using AWS Workload Identity Federation
// with direct service account impersonation.
//
// The authentication flow:
// 1. Reads JWT from file (ccConfig.JWTFilePath)
// 2. Assumes Elastic's AWS role using AssumeRoleWithWebIdentity
// 3. Uses AWS credentials for GCP Workload Identity Federation token exchange
// 4. Impersonates the target service account in the customer's GCP project
func (p *GoogleAuthProvider) FindIdentityFederationCredentials(ctx context.Context, ccConfig config.CloudConnectorsConfig, params GCPIdentityFederationParams) ([]option.ClientOption, error) {
	if ccConfig.JWTFilePath == "" {
		return nil, errors.New("identity federation config JWTFilePath is required")
	}
	if ccConfig.GlobalRoleARN == "" {
		return nil, errors.New("identity federation config GlobalRoleARN is required")
	}
	if ccConfig.ResourceID == "" {
		return nil, errors.New("identity federation config ResourceID is required")
	}
	if params.IdentityFederationID == "" {
		return nil, errors.New("identity federation config IdentityFederationID is required")
	}

	// Session name must match GCP Workload Identity Federation: elastic_resource_id-identity_federation_id
	sessionName := ccConfig.ResourceID + "-" + params.IdentityFederationID

	tokenSource, err := libbeatgcp.NewTokenSource(ctx, libbeatgcp.Params{
		Audience:            params.Audience,
		GlobalRoleARN:       ccConfig.GlobalRoleARN,
		JWTFilePath:         ccConfig.JWTFilePath,
		SessionName:         sessionName,
		ServiceAccountEmail: params.ServiceAccountEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create external account token source: %w", err)
	}

	return []option.ClientOption{option.WithTokenSource(tokenSource)}, nil
}
