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
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	k8sprovider "github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/uniqueness"
)

type EKS struct {
	AWSCfgProvider         awslib.ConfigProviderAPI
	AWSIdentityProvider    awslib.IdentityProviderGetter
	AWSMetadataProvider    awslib.MetadataProvider
	EKSClusterNameProvider awslib.EKSClusterNameProviderAPI
	ClientProvider         k8sprovider.ClientGetterAPI

	leaderElector uniqueness.Manager
}

func (k *EKS) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, error) {
	if err := k.checkDependencies(); err != nil {
		return nil, nil, err
	}

	kubeClient, err := k.ClientProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	k.leaderElector = uniqueness.NewLeaderElector(log, kubeClient)

	awsConfig, awsIdentity, err := k.getEksAwsConfig(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AWS config: %w", err)
	}

	clusterNameProvider := k8s.EKSClusterNameProvider{
		AwsCfg:              awsConfig,
		EKSMetadataProvider: k.AWSMetadataProvider,
		ClusterNameProvider: k.EKSClusterNameProvider,
		KubeClient:          kubeClient,
	}
	dp, err := getK8sDataProvider(ctx, log, *cfg, kubeClient, clusterNameProvider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create k8s data provider: %w", err)
	}

	return registry.NewRegistry(log, factory.NewCisEksFactory(log, awsConfig, ch, k.leaderElector, kubeClient, awsIdentity)), dp, nil
}

func (k *EKS) Run(ctx context.Context) error { return k.leaderElector.Run(ctx) }
func (k *EKS) Stop()                         { k.leaderElector.Stop() }

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
	missing := k.missingDependency()
	if missing != "" {
		return fmt.Errorf("%s is uninitialized", missing)
	}
	return nil
}

func (k *EKS) missingDependency() string {
	if k.AWSCfgProvider == nil {
		return "aws config provider"
	}
	if k.AWSIdentityProvider == nil {
		return "aws identity provider"
	}
	if k.ClientProvider == nil {
		return "kubernetes client provider"
	}
	if k.EKSClusterNameProvider == nil {
		return "EKS cluster name provider"
	}
	if k.AWSMetadataProvider == nil {
		return "AWS metadata provider"
	}
	return ""
}
