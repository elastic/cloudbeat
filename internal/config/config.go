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
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/launcher"
)

const (
	DefaultNamespace                = "default"
	VulnerabilityType               = "vuln_mgmt"
	AssetInventoryType              = "asset_inventory"
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
	Cred                  aws.ConfigAWS `config:"credentials"`
	AccountType           string        `config:"account_type"`
	CloudConnectors       bool          `config:"supports_cloud_connectors"`
	CloudConnectorsConfig CloudConnectorsConfig
}

type GcpConfig struct {
	// empty for OrganizationAccount
	ProjectId string `config:"project_id"`

	// empty for SingleAccount
	OrganizationId string `config:"organization_id"`

	// SingleAccount or OrganizationAccount
	AccountType string `config:"account_type"`

	GcpCallOpt GcpCallOpt `config:"call_options"`

	GcpClientOpt `config:"credentials"`
}

type GcpClientOpt struct {
	CredentialsJSON     string `config:"credentials_json"`
	CredentialsFilePath string `config:"credentials_file_path"`
}

type GcpCallOpt struct {
	ListAssetsTimeout  time.Duration `config:"list_assets_timeout"`
	ListAssetsPageSize int32         `config:"list_assets_page_size"`
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
	ClientAssertionPath       string `config:"client_assertion_path"`
}

const (
	AzureClientCredentialsTypeARMTemplate     = "arm_template"
	AzureClientCredentialsTypeManagedIdentity = "managed_identity"
	AzureClientCredentialsTypeManual          = "manual"
	AzureClientCredentialsTypeSecret          = "service_principal_with_client_secret"
	AzureClientCredentialsTypeCertificate     = "service_principal_with_client_certificate"
	AzureClientCredentialsTypeCloudConnectors = "cloud_connectors"
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

// DatastreamNamespace returns the inferred namespace setting from the Agent Policy
func (c *Config) DatastreamNamespace() string {
	if c.Index == "" {
		return DefaultNamespace
	}
	elems := strings.Split(c.Index, "-")
	return elems[len(elems)-1]
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

	// apply env var overwrites
	overwritesFromEnvVars(c)

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

	if c.CloudConfig.Aws.CloudConnectors {
		c.CloudConfig.Aws.CloudConnectorsConfig = newCloudConnectorsConfig()
	}

	return c, nil
}

func defaultConfig() (*Config, error) {
	ret := &Config{
		Period: 4 * time.Hour,
		CloudConfig: CloudConfig{
			Gcp: defaultGCPConfig(),
		},
	}

	bundle, err := getBundlePath()
	if err != nil {
		return nil, err
	}

	ret.BundlePath = bundle
	return ret, nil
}

func defaultGCPConfig() GcpConfig {
	return GcpConfig{
		GcpCallOpt: GcpCallOpt{
			// default value from sdk is 1m; we use 4m to exceed quota window.
			// https://github.com/googleapis/google-cloud-go/blob/952cd7fd419af9eb74f5d30a111ae936094b0645/asset/apiv1/asset_client.go#L96
			ListAssetsTimeout: 4 * time.Minute,

			// default value from sdk is 100; we use 200
			// https://github.com/googleapis/google-cloud-go/blob/a6c85f6387ee6aa291e786c882637fb03f3302f4/asset/apiv1/assetpb/asset_service.pb.go#L767-L769
			ListAssetsPageSize: 200,
		},
	}
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

// Cloud Connectors roles and resource id must be provided by the system (controller)
// and not user input (package policy) for security reasons.

const (
	CloudConnectorsLocalRoleEnvVar  = "CLOUD_CONNECTORS_LOCAL_ROLE"
	CloudConnectorsGlobalRoleEnvVar = "CLOUD_CONNECTORS_GLOBAL_ROLE"
	CloudResourceIDEnvVar           = "CLOUD_RESOURCE_ID"
)

type CloudConnectorsConfig struct {
	LocalRoleARN  string
	GlobalRoleARN string
	ResourceID    string
}

func newCloudConnectorsConfig() CloudConnectorsConfig {
	return CloudConnectorsConfig{
		LocalRoleARN:  os.Getenv(CloudConnectorsLocalRoleEnvVar),
		GlobalRoleARN: os.Getenv(CloudConnectorsGlobalRoleEnvVar),
		ResourceID:    os.Getenv(CloudResourceIDEnvVar),
	}
}

const (
	CloudbeatGCPListAssetPageSizeEnvVar = "CLOUDBEAT_GCP_LIST_ASSETS_PAGE_SIZE"
	CloudbeatGCPListAssetTimeoutEnvVar  = "CLOUDBEAT_GCP_LIST_ASSETS_TIMEOUT"
)

func overwritesFromEnvVars(c *Config) {
	log := clog.NewLogger("config")
	logErr := func(name string, value string, err error) {
		log.Errorw(
			"error trying to parse config env variable",
			logp.String("name", name),
			logp.String("value", value),
			logp.Error(err),
		)
	}

	if v, exists := os.LookupEnv(CloudbeatGCPListAssetPageSizeEnvVar); exists {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			c.CloudConfig.Gcp.GcpCallOpt.ListAssetsPageSize = int32(i)
		} else {
			logErr(CloudbeatGCPListAssetPageSizeEnvVar, v, err)
		}
	}

	if v, exists := os.LookupEnv(CloudbeatGCPListAssetTimeoutEnvVar); exists {
		if d, err := time.ParseDuration(v); err == nil {
			c.CloudConfig.Gcp.GcpCallOpt.ListAssetsTimeout = d
		} else {
			logErr(CloudbeatGCPListAssetPageSizeEnvVar, v, err)
		}
	}
}
