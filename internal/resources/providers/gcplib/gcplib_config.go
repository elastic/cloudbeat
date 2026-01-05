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

package gcplib

import (
	"context"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

// InitializeGCPConfigCloudConnectors initializes GCP config for Cloud Connectors deployment.
// It uses OIDC-based authentication with service account impersonation chaining.
func InitializeGCPConfigCloudConnectors(ctx context.Context, cfg config.GcpConfig) ([]option.ClientOption, error) {
	oidcFilePath := os.Getenv(config.CloudConnectorsJWTPathEnvVar)
	if oidcFilePath == "" {
		return nil, errors.New("unable to initialize GCP config for Cloud Connectors: CLOUD_CONNECTORS_ID_TOKEN_FILE not set")
	}

	return NewGCPConfigOIDCChain(ctx, oidcFilePath, cfg)
}

// NewGCPConfigOIDCChain creates GCP client options using OIDC/Web Identity token-based authentication
// with a two-step service account impersonation chain.
// This function performs a two-step authentication chain:
//  1. Uses the OIDC token to authenticate via Workload Identity Federation and impersonate the Elastic Global Service Account
//  2. Uses the Global Service Account credentials to impersonate the customer's Target Service Account
func NewGCPConfigOIDCChain(ctx context.Context, jwtFilePath string, cfg config.GcpConfig) ([]option.ClientOption, error) {
	ccCfg := cfg.CloudConnectorsConfig

	// Validate required configuration
	if ccCfg.GlobalServiceAccount == "" {
		return nil, errors.New("global service account is required for Cloud Connectors")
	}
	if cfg.ServiceAccountEmail == "" {
		return nil, errors.New("target service account email is required for Cloud Connectors")
	}
	if ccCfg.GlobalAudience == "" {
		return nil, errors.New("global audience is required for Cloud Connectors")
	}

	// Construct the service account impersonation URL for the Global Service Account
	impersonationURL := fmt.Sprintf(
		"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken",
		ccCfg.GlobalServiceAccount,
	)

	// Build the authentication chain
	chain := []GCPServiceAccountChainingStep{
		// Chain Step 1 - Authenticate via OIDC and impersonate Elastic Global Service Account
		&ExternalAccountStep{
			Config: externalaccount.Config{
				Audience:         ccCfg.GlobalAudience,
				SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
				TokenURL:         "https://sts.googleapis.com/v1/token",
				CredentialSource: &externalaccount.CredentialSource{
					File: jwtFilePath,
					Format: externalaccount.Format{
						Type: "text",
					},
				},
				ServiceAccountImpersonationURL: impersonationURL,
			},
		},
		// Chain Step 2 - Impersonate customer's Target Service Account
		&ImpersonateServiceAccountStep{
			TargetPrincipal: cfg.ServiceAccountEmail,
			Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
		},
	}

	return GCPClientOptionsChaining(ctx, chain)
}
