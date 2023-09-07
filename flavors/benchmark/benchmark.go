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

package benchmark

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	azure_auth "github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	azure_inventory "github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
	gcp_auth "github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
	gcp_inventory "github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
)

type Benchmark interface {
	Run(ctx context.Context) error
	Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error)
	Stop()

	checkDependencies() error
}

func NewBenchmark(cfg *config.Config) (Benchmark, error) {
	switch cfg.Benchmark {
	case config.CIS_AWS:
		if cfg.CloudConfig.Aws.AccountType == config.OrganizationAccount {
			return &AWSOrg{
				IdentityProvider: awslib.IdentityProvider{},
				AccountProvider:  awslib.AccountProvider{},
			}, nil
		}

		return &AWS{
			IdentityProvider: awslib.IdentityProvider{},
		}, nil
	case config.CIS_EKS:
		return &EKS{
			AWSCfgProvider:         awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
			AWSIdentityProvider:    awslib.IdentityProvider{},
			AWSMetadataProvider:    awslib.Ec2MetadataProvider{},
			EKSClusterNameProvider: awslib.EKSClusterNameProvider{},
			ClientProvider:         k8s.ClientGetter{},
			leaderElector:          nil,
		}, nil
	case config.CIS_K8S:
		return &K8S{
			ClientProvider: k8s.ClientGetter{},
			leaderElector:  nil,
		}, nil
	case config.CIS_GCP:
		return &GCP{
			CfgProvider:          &gcp_auth.ConfigProvider{AuthProvider: &gcp_auth.GoogleAuthProvider{}},
			inventoryInitializer: &gcp_inventory.ProviderInitializer{},
		}, nil
	case config.CIS_AZURE:
		return &Azure{
			// IdentityProvider:     &azure_identity.Provider{},
			CfgProvider:          &azure_auth.ConfigProvider{AuthProvider: &azure_auth.AzureAuthProvider{}},
			inventoryInitializer: &azure_inventory.ProviderInitializer{}}, nil
	}
	return nil, fmt.Errorf("unknown benchmark: '%s'", cfg.Benchmark)
}
