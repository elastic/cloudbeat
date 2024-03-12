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

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	gcp_auth "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
	gcp_inventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

type Strategy interface {
	NewBenchmark(ctx context.Context, log *logp.Logger, cfg *config.Config) (builder.Benchmark, error)
	checkDependencies() error
}

func GetStrategy(cfg *config.Config) (Strategy, error) {
	switch cfg.Benchmark {
	case config.CIS_AWS:
		if cfg.CloudConfig.Aws.AccountType == config.OrganizationAccount {
			return &AWSOrg{
				IAMProvider:      &iam.Provider{},
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
			cfgProvider:         &azure_auth.ConfigProvider{AuthProvider: &azure_auth.AzureAuthProvider{}},
			providerInitializer: &azurelib.ProviderInitializer{}}, nil
	}
	return nil, fmt.Errorf("unknown benchmark: '%s'", cfg.Benchmark)
}
