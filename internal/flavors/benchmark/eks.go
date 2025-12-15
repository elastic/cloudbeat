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

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/preset"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/uniqueness"
)

type EKS struct {
	AWSIdentityProvider    awslib.IdentityProviderGetter
	AWSMetadataProvider    awslib.MetadataProvider
	EKSClusterNameProvider awslib.EKSClusterNameProviderAPI
	ClientProvider         k8s.ClientGetterAPI

	leaderElector uniqueness.Manager
}

func (k *EKS) NewBenchmark(ctx context.Context, log *clog.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, idp, err := k.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
		builder.WithIdProvider(idp),
	).BuildK8s(ctx, log, cfg, resourceCh, reg, k.leaderElector)
}

//revive:disable-next-line:function-result-limit
func (k *EKS) initialize(ctx context.Context, log *clog.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := k.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	kubeClient, err := k.ClientProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	benchmarkHelper := NewK8sBenchmarkHelper(log, cfg, kubeClient)
	k.leaderElector = uniqueness.NewLeaderElector(log, kubeClient)

	awsConfig, awsIdentity, err := k.getEksAwsConfig(ctx, cfg, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize AWS config: %w", err)
	}

	clusterNameProvider := k8s.EKSClusterNameProvider{
		AwsCfg:              awsConfig,
		EKSMetadataProvider: k.AWSMetadataProvider,
		ClusterNameProvider: k.EKSClusterNameProvider,
		KubeClient:          kubeClient,
	}
	dp, err := benchmarkHelper.GetK8sDataProvider(ctx, clusterNameProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s data provider: %w", err)
	}

	idp, err := benchmarkHelper.GetK8sIdProvider(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s id provider: %w", err)
	}

	return registry.NewRegistry(
		log,
		registry.WithFetchersMap(preset.NewCisEksFetchers(ctx, log, awsConfig, ch, k.leaderElector, kubeClient, awsIdentity)),
	), dp, idp, nil
}

func (k *EKS) getEksAwsConfig(ctx context.Context, cfg *config.Config, log *clog.Logger) (awssdk.Config, *cloud.Identity, error) {
	if cfg.CloudConfig == (config.CloudConfig{}) || cfg.CloudConfig.Aws.Cred == (aws.ConfigAWS{}) {
		// Optional for EKS
		return awssdk.Config{}, nil, nil
	}

	awsCfg, err := awslib.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred, log.Logger)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}
	metadata, err := k.AWSMetadataProvider.GetMetadata(ctx, *awsCfg)
	if err != nil {
		return *awsCfg, nil, fmt.Errorf("failed to retrieve aws metadata: %w", err)
	}
	awsCfg.Region = metadata.Region

	identity, err := k.AWSIdentityProvider.GetIdentity(ctx, *awsCfg)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	return *awsCfg, identity, nil
}

func (k *EKS) checkDependencies() error {
	if k.AWSIdentityProvider == nil {
		return errors.New("aws identity provider is uninitialized")
	}
	if k.ClientProvider == nil {
		return errors.New("kubernetes client provider is uninitialized")
	}
	if k.EKSClusterNameProvider == nil {
		return errors.New("eks cluster name provider is uninitialized")
	}
	if k.AWSMetadataProvider == nil {
		return errors.New("aws metadata provider is uninitialized")
	}
	return nil
}
