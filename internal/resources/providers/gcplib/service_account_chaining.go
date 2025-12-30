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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

// GCPServiceAccountChainingStep represents a single step in a GCP service account impersonation chain.
// Implementations should provide a way to build a token source from a previous token source.
type GCPServiceAccountChainingStep interface {
	// BuildTokenSource creates a token source using the previous step's token source
	// For the first step, previousTokenSource may be nil
	BuildTokenSource(ctx context.Context, previousTokenSource oauth2.TokenSource) (oauth2.TokenSource, error)
}

// GCPClientOptionsChaining builds client options by chaining service account impersonation steps.
// Each step uses the previous step's token source to build the next token source.
func GCPClientOptionsChaining(ctx context.Context, chain []GCPServiceAccountChainingStep) ([]option.ClientOption, error) {
	var tokenSource oauth2.TokenSource

	for _, step := range chain {
		ts, err := step.BuildTokenSource(ctx, tokenSource)
		if err != nil {
			return nil, err
		}
		tokenSource = ts
	}

	return []option.ClientOption{
		option.WithTokenSource(tokenSource),
	}, nil
}

// ExternalAccountStep represents an external account (OIDC/Workload Identity Federation) authentication step.
// This is typically the first step in the chain when using OIDC tokens.
type ExternalAccountStep struct {
	Config externalaccount.Config
}

// BuildTokenSource implements GCPServiceAccountChainingStep for external account authentication.
func (s *ExternalAccountStep) BuildTokenSource(ctx context.Context, _ oauth2.TokenSource) (oauth2.TokenSource, error) {
	// External account is always the first step, so previousTokenSource is ignored
	return externalaccount.NewTokenSource(ctx, s.Config)
}

// ImpersonateServiceAccountStep represents a service account impersonation step.
// This step impersonates a service account using credentials from the previous step.
type ImpersonateServiceAccountStep struct {
	TargetPrincipal string
	Scopes          []string
	Delegates       []string
}

// BuildTokenSource implements GCPServiceAccountChainingStep for service account impersonation.
func (s *ImpersonateServiceAccountStep) BuildTokenSource(ctx context.Context, previousTokenSource oauth2.TokenSource) (oauth2.TokenSource, error) {
	config := impersonate.CredentialsConfig{
		TargetPrincipal: s.TargetPrincipal,
		Scopes:          s.Scopes,
		Delegates:       s.Delegates,
	}

	return impersonate.CredentialsTokenSource(ctx, config, option.WithTokenSource(previousTokenSource))
}

// Compile-time checks to ensure types implement GCPServiceAccountChainingStep
var (
	_ GCPServiceAccountChainingStep = (*ExternalAccountStep)(nil)
	_ GCPServiceAccountChainingStep = (*ImpersonateServiceAccountStep)(nil)
)
