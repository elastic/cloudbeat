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
	"os"

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
	// Token type for JWT-based authentication
	jwtTokenType = "urn:ietf:params:oauth:token-type:jwt"
	// Default scope for GCP Cloud Platform access
	gcpCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

type GoogleAuthProvider struct{}

// FindDefaultCredentials is a wrapper around google.FindDefaultCredentials to make it easier to mock
func (p *GoogleAuthProvider) FindDefaultCredentials(ctx context.Context) (*google.Credentials, error) {
	return google.FindDefaultCredentials(ctx)
}

// FindCloudConnectorsCredentials creates GCP client options using OIDC/Web Identity token-based authentication
// with direct service account impersonation. The target Service Account must trust the OIDC provider.
func (p *GoogleAuthProvider) FindCloudConnectorsCredentials(ctx context.Context, audience string, serviceAccountEmail string) ([]option.ClientOption, error) {
	jwtFilePath := os.Getenv(config.CloudConnectorsJWTPathEnvVar)
	if jwtFilePath == "" {
		return nil, fmt.Errorf("environment variable %s is required for cloud connectors credentials", config.CloudConnectorsJWTPathEnvVar)
	}

	cfg := externalaccount.Config{
		Audience:         audience,
		SubjectTokenType: jwtTokenType,
		TokenURL:         gcpSTSTokenURL,
		Scopes:           []string{gcpCloudPlatformScope},
		CredentialSource: &externalaccount.CredentialSource{
			File:   jwtFilePath,
			Format: externalaccount.Format{Type: "text"},
		},
		ServiceAccountImpersonationURL: gcpIAMCredentialsURL + serviceAccountEmail + ":generateAccessToken",
	}

	tokenSource, err := externalaccount.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create external account token source: %w", err)
	}

	return []option.ClientOption{option.WithTokenSource(tokenSource)}, nil
}
