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

package k8s

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes/metadata"
	agent_config "github.com/elastic/elastic-agent-libs/config"
	"k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type ClusterNameProviderAPI interface {
	GetClusterName(ctx context.Context, cfg *config.Config) (string, error)
}

type EKSClusterNameProvider struct {
	AwsCfg              aws.Config
	EKSMetadataProvider awslib.MetadataProvider
	ClusterNameProvider awslib.EKSClusterNameProviderAPI
	KubeClient          kubernetes.Interface
}

func (provider EKSClusterNameProvider) GetClusterName(ctx context.Context, _ *config.Config) (string, error) {
	mdata, err := provider.EKSMetadataProvider.GetMetadata(ctx, provider.AwsCfg)
	if err != nil {
		return "", fmt.Errorf("failed to get the ec2 metadata required for identifying the cluster name: %v", err)
	}
	return provider.ClusterNameProvider.GetClusterName(ctx, provider.AwsCfg, mdata.InstanceID)
}

type KubernetesClusterNameProvider struct {
	KubeClient kubernetes.Interface
}

func (provider KubernetesClusterNameProvider) GetClusterName(_ context.Context, cfg *config.Config) (string, error) {
	agentConfig, err := agent_config.NewConfigFrom(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create agent config: %v", err)
	}
	clusterIdentifier, err := metadata.GetKubernetesClusterIdentifier(agentConfig, provider.KubeClient)
	if err != nil {
		return "", fmt.Errorf("fail to resolve the name of the cluster: %v", err)
	}

	return clusterIdentifier.Name, nil
}
