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

package providers

import (
	"context"
	"fmt"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	k8s "k8s.io/client-go/kubernetes"
)

type ClusterNameProviderAPI interface {
	GetClusterName(ctx context.Context, cfg *config.Config, log *logp.Logger) (string, error)
}

type ClusterNameProvider struct {
	KubernetesClusterNameProvider KubernetesClusterNameProviderApi
	EKSMetadataProvider           awslib.MetadataProvider
	EKSClusterNameProvider        awslib.ClusterNameProvider
	KubeClient                    k8s.Interface
	AwsConfigProvider             awslib.ConfigProviderAPI
}

func (provider ClusterNameProvider) GetClusterName(ctx context.Context, cfg *config.Config, log *logp.Logger) (string, error) {
	switch cfg.BenchmarkConfig.ID {
	case config.CIS_K8S:
		log.Debugf("Trying to identify Kubernetes Vanilla cluster name")
		return provider.KubernetesClusterNameProvider.GetClusterName(cfg, provider.KubeClient)
	case config.CIS_EKS:
		log.Debugf("Trying to identify EKS cluster name")
		awsConfig, err := provider.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.BenchmarkConfig.AWSConfig.Credentials, log)
		if err != nil {
			return "", fmt.Errorf("failed to initialize aws configuration for identifying the cluster name: %v", err)
		}
		metadata, err := provider.EKSMetadataProvider.GetMetadata(ctx, awsConfig)
		if err != nil {
			return "", fmt.Errorf("failed to get the ec2 metadata required for identifying the cluster name: %v", err)
		}
		instanceId := metadata.InstanceID
		return provider.EKSClusterNameProvider.GetClusterName(ctx, awsConfig, instanceId)
	default:
		panic(fmt.Sprintf("cluster name provider encountered an unknown cluster type: %s, please implement the relevant cluster name provider", cfg.BenchmarkConfig.ID))
	}
}
