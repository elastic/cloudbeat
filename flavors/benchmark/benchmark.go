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
	"errors"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	k8sprovider "github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type Benchmark interface {
	Run(ctx context.Context) error
	Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, dependencies *Dependencies) (registry.Registry, dataprovider.CommonDataProvider, error)
	Stop()
}

func NewBenchmark(cfg *config.Config) (Benchmark, error) {
	switch cfg.Benchmark {
	case config.CIS_AWS:
		return &AWS{}, nil
	case config.CIS_EKS:
		return &EKS{}, nil
	case config.CIS_K8S:
		return &K8S{}, nil
	}
	return nil, fmt.Errorf("unknown benchmark: '%s'", cfg.Benchmark)
}

type Dependencies struct {
	awsCfgProvider         awslib.ConfigProviderAPI
	identityProvider       awslib.IdentityProviderGetter
	kubernetesProvider     k8sprovider.ClientGetterAPI
	metadataProvider       awslib.MetadataProvider
	awsClusterNameProvider awslib.EKSClusterNameProviderAPI
}

func NewDependencies(
	awsCfgProvider awslib.ConfigProviderAPI,
	identityProvider awslib.IdentityProviderGetter,
	kubernetesProvider k8sprovider.ClientGetterAPI,
	metadataProvider awslib.MetadataProvider,
	awsClusterNameProvider awslib.EKSClusterNameProviderAPI,
) *Dependencies {
	return &Dependencies{
		awsCfgProvider:         awsCfgProvider,
		identityProvider:       identityProvider,
		kubernetesProvider:     kubernetesProvider,
		metadataProvider:       metadataProvider,
		awsClusterNameProvider: awsClusterNameProvider,
	}
}

func (d *Dependencies) KubernetesClient(log *logp.Logger, kubeConfig string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
	if d.kubernetesProvider == nil {
		return nil, fmt.Errorf("k8s provider is uninitialized")
	}
	return d.kubernetesProvider.GetClient(log, kubeConfig, options)
}

func (d *Dependencies) AWSConfig(ctx context.Context, cfg aws.ConfigAWS) (*awssdk.Config, error) {
	if d.awsCfgProvider == nil {
		return nil, errors.New("aws config provider is uninitialized")
	}
	return d.awsCfgProvider.InitializeAWSConfig(ctx, cfg)
}

func (d *Dependencies) AWSIdentity(ctx context.Context, cfg awssdk.Config) (*awslib.Identity, error) {
	if d.identityProvider == nil {
		return nil, errors.New("aws identity provider is uninitialized")
	}
	return d.identityProvider.GetIdentity(ctx, cfg)
}
