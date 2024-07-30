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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"

	"github.com/elastic/cloudbeat/internal/launcher"
)

const (
	DefaultNamespace                = "default"
	VulnerabilityType               = "vuln_mgmt"
	AssetInventoryType              = "asset_inventory"
	ProviderAWS                     = "aws"
	ProviderAzure                   = "azure"
	ProviderGCP                     = "gcp"
	defaultFindingsIndexPrefix      = "logs-cloud_security_posture.findings"
	defaultVulnerabilityIndexPrefix = "logs-cloud_security_posture.vulnerabilities"
)

type Fetcher struct {
	Name string `config:"name"` // Name of the fetcher
}

type Config struct {
	Benchmark              string                  `config:"config.v1.benchmark"`
	Type                   string                  `config:"config.v1.type"`
	Deployment             string                  `config:"config.v1.deployment"`
	AssetInventoryProvider string                  `config:"config.v1.asset_inventory_provider"`
	CloudConfig            CloudConfig             `config:"config.v1"`
	KubeConfig             string                  `config:"kube_config"`
	Period                 time.Duration           `config:"period"`
	Processors             processors.PluginConfig `config:"processors"`
	BundlePath             string                  `config:"bundle_path"`
	PackagePolicyId        string                  `config:"package_policy_id"`
	PackagePolicyRevision  int                     `config:"revision"`
	Index                  string                  `config:"index"`
}

type CloudConfig struct {
	Aws   AwsConfig   `config:"aws"`
	Gcp   GcpConfig   `config:"gcp"`
	Azure AzureConfig `config:"azure"`
}

type AwsConfig struct {
	Cred        aws.ConfigAWS `config:"credentials"`
	AccountType string        `config:"account_type"`
}

type GcpConfig struct {
	// empty for OrganizationAccount
	ProjectId string `config:"project_id"`

	// empty for SingleAccount
	OrganizationId string `config:"organization_id"`

	// SingleAccount or OrganizationAccount
	AccountType string `config:"account_type"`

	GcpClientOpt `config:"credentials"`
}

type GcpClientOpt struct {
	CredentialsJSON     string `config:"credentials_json"`
	CredentialsFilePath string `config:"credentials_file_path"`
}

type AzureConfig struct {
	Credentials AzureClientOpt `config:"credentials"`
	// SingleAccount or OrganizationAccount
	AccountType string `config:"account_type"`
}

type AzureClientOpt struct {
	ClientCredentialsType     string `config:"type"`
	ClientID                  string `config:"client_id"`
	TenantID                  string `config:"tenant_id"`
	ClientSecret              string `config:"client_secret"`
	ClientUsername            string `config:"client_username"`
	ClientPassword            string `config:"client_password"`
	ClientCertificatePath     string `config:"client_certificate_path"`
	ClientCertificatePassword string `config:"client_certificate_password"`
}

const (
	AzureClientCredentialsTypeARMTemplate      = "arm_template"
	AzureClientCredentialsTypeManagedIdentity  = "managed_identity"
	AzureClientCredentialsTypeManual           = "manual"
	AzureClientCredentialsTypeSecret           = "service_principal_with_client_secret"
	AzureClientCredentialsTypeCertificate      = "service_principal_with_client_certificate"
	AzureClientCredentialsTypeUsernamePassword = "service_principal_with_client_username_and_password"
)

const (
	SingleAccount       = "single-account"
	OrganizationAccount = "organization-account"
)

// Datastream returns the name of a Data Stream to publish Cloudbeat events to.
func (c *Config) Datastream() string {
	if c.Index != "" {
		return c.Index
	}
	if c.Type == VulnerabilityType {
		return defaultVulnerabilityIndexPrefix + "-" + DefaultNamespace
	}
	return defaultFindingsIndexPrefix + "-" + DefaultNamespace
}

func New(cfg *config.C) (*Config, error) {
	c, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	if c.Benchmark != "" {
		if !isSupportedBenchmark(c.Benchmark) {
			return c, launcher.NewUnhealthyError(fmt.Sprintf("benchmark '%s' is not supported", c.Benchmark))
		}
	}

	switch c.CloudConfig.Aws.AccountType {
	case "":
	case SingleAccount:
	case OrganizationAccount:
	default:
		return nil, launcher.NewUnhealthyError(fmt.Sprintf(
			"aws.account_type '%s' is not supported",
			c.CloudConfig.Aws.AccountType,
		))
	}

	switch c.CloudConfig.Azure.AccountType {
	case "":
	case SingleAccount:
	case OrganizationAccount:
	default:
		return nil, launcher.NewUnhealthyError(fmt.Sprintf(
			"azure.account_type '%s' is not supported",
			c.CloudConfig.Azure.AccountType,
		))
	}

	return c, nil
}

func defaultConfig() (*Config, error) {
	ret := &Config{
		Period: 4 * time.Hour,
	}

	bundle, err := getBundlePath()
	if err != nil {
		return nil, err
	}

	ret.BundlePath = bundle
	return ret, nil
}

func getBundlePath() (string, error) {
	// The bundle resides on the same location as the executable
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(ex), "bundle.tar.gz"), nil
}

func isSupportedBenchmark(benchmark string) bool {
	for _, s := range SupportedCIS {
		if benchmark == s {
			return true
		}
	}
	return false
}
