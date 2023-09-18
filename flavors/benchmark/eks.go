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

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/preset"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/uniqueness"
)

type EKS struct {
	AWSCfgProvider         awslib.ConfigProviderAPI
	AWSIdentityProvider    awslib.IdentityProviderGetter
	AWSMetadataProvider    awslib.MetadataProvider
	EKSClusterNameProvider awslib.EKSClusterNameProviderAPI
	ClientProvider         k8s.ClientGetterAPI
}

func (k *EKS) NewBenchmark(ctx context.Context, log *logp.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)
	if err := k.checkDependencies(); err != nil {
		return nil, err
	}

	kubeClient, err := k.ClientProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	benchmarkHelper := NewK8sBenchmarkHelper(log, cfg, kubeClient)
	leaderElector := uniqueness.NewLeaderElector(log, kubeClient)

	awsConfig, awsIdentity, err := k.getEksAwsConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS config: %w", err)
	}

	clusterNameProvider := k8s.EKSClusterNameProvider{
		AwsCfg:              awsConfig,
		EKSMetadataProvider: k.AWSMetadataProvider,
		ClusterNameProvider: k.EKSClusterNameProvider,
		KubeClient:          kubeClient,
	}
	bdp, err := benchmarkHelper.GetK8sDataProvider(ctx, clusterNameProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s data provider: %w", err)
	}

	idp, err := benchmarkHelper.GetK8sIdProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s id provider: %w", err)
	}

	fetchers := preset.NewCisEksFetchers(log, awsConfig, resourceCh, leaderElector, kubeClient, awsIdentity)
	reg := registry.NewRegistry(log, registry.WithFetchersMap(fetchers))

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
		builder.WithIdProvider(idp),
	).Build(ctx, log, cfg, resourceCh, reg)
}

func (k *EKS) getEksAwsConfig(ctx context.Context, cfg *config.Config) (awssdk.Config, *cloud.Identity, error) {
	if cfg.CloudConfig == (config.CloudConfig{}) || cfg.CloudConfig.Aws.Cred == (aws.ConfigAWS{}) {
		// Optional for EKS
		return awssdk.Config{}, nil, nil
	}

	awsCfg, err := k.AWSCfgProvider.InitializeAWSConfig(ctx, cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identity, err := k.AWSIdentityProvider.GetIdentity(ctx, *awsCfg)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	return *awsCfg, identity, nil
}

func (k *EKS) checkDependencies() error {
	if k.AWSCfgProvider == nil {
		return fmt.Errorf("aws config provider is uninitialized")
	}
	if k.AWSIdentityProvider == nil {
		return fmt.Errorf("aws identity provider is uninitialized")
	}
	if k.ClientProvider == nil {
		return fmt.Errorf("kubernetes client provider is uninitialized")
	}
	if k.EKSClusterNameProvider == nil {
		return fmt.Errorf("eks cluster name provider is uninitialized")
	}
	if k.AWSMetadataProvider == nil {
		return fmt.Errorf("aws metadata provider is uninitialized")
	}
	return nil
}
