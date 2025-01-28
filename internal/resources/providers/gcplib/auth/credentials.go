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
}

type ConfigProvider struct {
	AuthProvider GoogleAuthProviderAPI
}

var ErrMissingOrgId = errors.New("organization ID is required for organization account type")
var ErrInvalidCredentialsJSON = errors.New("invalid credentials JSON")
var ErrProjectNotFound = errors.New("no project ID was found")

func (p *ConfigProvider) GetGcpClientConfig(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	// used in cloud shell flow (and development)
	if cfg.CredentialsJSON == "" && cfg.CredentialsFilePath == "" {
		return p.getApplicationDefaultCredentials(ctx, cfg, log)
	}

	// used in the manual flow
	return p.getCustomCredentials(ctx, cfg, log)
}

func (p *ConfigProvider) getGcpFactoryConfig(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (*GcpFactoryConfig, error) {
	parent, err := getGcpConfigParentValue(ctx, *p, cfg)
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

func (p *ConfigProvider) getCustomCredentials(ctx context.Context, cfg config.GcpConfig, log *clog.Logger) (*GcpFactoryConfig, error) {
	log.Info("getCustomCredentialsConfig create credentials options")

	var opts []option.ClientOption
	if cfg.CredentialsFilePath != "" {
		if err := validateJSONFromFile(cfg.CredentialsFilePath); err != nil {
			return nil, err
		}
		log.Infof("Appending credentials file path to gcp client options: %s", cfg.CredentialsFilePath)
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFilePath))
	}
	if cfg.CredentialsJSON != "" {
		if !json.Valid([]byte(cfg.CredentialsJSON)) {
			return nil, ErrInvalidCredentialsJSON
		}
		log.Info("Appending credentials JSON to client options")
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	}

	return p.getGcpFactoryConfig(ctx, cfg, opts)
}
func (p *ConfigProvider) getProjectId(ctx context.Context, cfg config.GcpConfig) (string, error) {
	if cfg.ProjectId != "" {
		return cfg.ProjectId, nil
	}

	// Try to get project ID from metadata server in case we are running on GCP VM
	cred, err := p.AuthProvider.FindDefaultCredentials(ctx)
	if err == nil {
		return cred.ProjectID, nil
	}

	return "", ErrProjectNotFound
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

func getGcpConfigParentValue(ctx context.Context, provider ConfigProvider, cfg config.GcpConfig) (string, error) {
	switch cfg.AccountType {
	case config.OrganizationAccount:
		if cfg.OrganizationId == "" {
			return "", ErrMissingOrgId
		}
		return fmt.Sprintf("organizations/%s", cfg.OrganizationId), nil
	case config.SingleAccount:
		projectId, err := provider.getProjectId(ctx, cfg)
		if err != nil {
			return "", fmt.Errorf("failed to get project ID: %v", err)
		}
		return fmt.Sprintf("projects/%s", projectId), nil
	default:
		return "", fmt.Errorf("invalid gcp account type: %s", cfg.AccountType)
	}
}
