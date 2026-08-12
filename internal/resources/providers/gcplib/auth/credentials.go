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
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type GcpFactoryConfig struct {
	// organizations/%s or projects/%s
	Parent     string
	ClientOpts []option.ClientOption
}

type ConfigProviderAPI interface {
	GetGcpClientConfig(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error)
}

type GoogleAuthProviderAPI interface {
	FindDefaultCredentials(ctx context.Context) (*google.Credentials, error)
	FindIdentityFederationCredentials(ctx context.Context, ccConfig config.CloudConnectorsConfig, params GCPIdentityFederationParams) ([]option.ClientOption, error)
}

// DefaultCredentialsFinder is the minimal interface needed to resolve project ID from
// application default credentials (e.g. metadata server on GCP). *GoogleAuthProvider implements it.
type DefaultCredentialsFinder interface {
	FindDefaultCredentials(ctx context.Context) (*google.Credentials, error)
}

// ParentResolver returns the GCP parent (e.g. "projects/pid" or "organizations/oid")
// for the given config and client options.
type ParentResolver interface {
	GetParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error)
}

type ConfigProvider struct {
	AuthProvider   GoogleAuthProviderAPI
	ParentResolver ParentResolver // required; use DefaultParentResolver in production
}

// NewConfigProvider returns a ConfigProvider wired with the default auth provider
// and default parent resolver (project + organization). Use this in production.
func NewConfigProvider() *ConfigProvider {
	auth := &GoogleAuthProvider{}
	return &ConfigProvider{
		AuthProvider:   auth,
		ParentResolver: NewDefaultParentResolver(auth),
	}
}

var ErrMissingOrgId = errors.New("organization ID is required for organization account type")
var ErrInvalidCredentialsJSON = errors.New("invalid credentials JSON")
var ErrProjectNotFound = errors.New("no project ID was found")

func (p *ConfigProvider) GetGcpClientConfig(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	// used in identity federation flow
	if cfg.GcpClientOpt.ServiceAccountEmail != "" {
		return p.getIdentityFederationCredentials(ctx, cfg, log)
	}

	// used in cloud shell flow (and development)
	if cfg.CredentialsJSON == "" && cfg.CredentialsFilePath == "" {
		return p.getApplicationDefaultCredentials(ctx, cfg, log)
	}

	// used in the manual flow
	return p.getCustomCredentials(ctx, cfg, log)
}

func (p *ConfigProvider) getGcpFactoryConfig(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (*GcpFactoryConfig, error) {
	parent, err := p.ParentResolver.GetParent(ctx, cfg, clientOpts)
	if err != nil {
		return nil, err
	}
	return &GcpFactoryConfig{
		Parent:     parent,
		ClientOpts: clientOpts,
	}, nil
}

// https://cloud.google.com/docs/authentication/application-default-credentials
func (p *ConfigProvider) getApplicationDefaultCredentials(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	log.Info("getDefaultCredentialsConfig create credentials options")
	return p.getGcpFactoryConfig(ctx, cfg, nil)
}

func (p *ConfigProvider) getIdentityFederationCredentials(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	log.Info("creating credentials using AWS Workload Identity Federation and service account impersonation", "provider", "GCP")

	params := GCPIdentityFederationParams{
		Audience:             cfg.Audience,
		ServiceAccountEmail:  cfg.ServiceAccountEmail,
		IdentityFederationID: cfg.IdentityFederationID,
	}
	opts, err := p.AuthProvider.FindIdentityFederationCredentials(ctx, cfg.CloudConnectorsConfig, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity federation credentials: %w", err)
	}

	return p.getGcpFactoryConfig(ctx, cfg, opts)
}

func (p *ConfigProvider) getCustomCredentials(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	log.Info("getCustomCredentialsConfig create credentials options")

	var opts []option.ClientOption
	if cfg.CredentialsFilePath != "" {
		if err := validateJSONFromFile(cfg.CredentialsFilePath); err != nil {
			return nil, err
		}
		log.Infof("Appending credentials file path to gcp client options: %s", cfg.CredentialsFilePath)
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.CredentialsFilePath))
	}
	if cfg.CredentialsJSON != "" {
		if !json.Valid([]byte(cfg.CredentialsJSON)) {
			return nil, ErrInvalidCredentialsJSON
		}
		log.Info("Appending credentials JSON to client options")
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(cfg.CredentialsJSON)))
	}

	return p.getGcpFactoryConfig(ctx, cfg, opts)
}

func validateJSONFromFile(filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %q cannot be found", filePath)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("the file %q cannot be read", filePath)
	}

	if !json.Valid(b) {
		return fmt.Errorf("the file %q does not contain valid JSON", filePath)
	}

	return nil
}
