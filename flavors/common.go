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

package flavors

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	aws_dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/aws"
	k8s_dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigureProcessors configure processors to be used by the beat
func ConfigureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}

func GetCommonDataProvider(ctx context.Context, log *logp.Logger, cfg config.Config) (dataprovider.CommonDataProvider, error) {
	log.Infof("Flavors.GetCommonDataProvider, constructing common data provider for benchmark: %s", cfg.Benchmark)
	switch cfg.Benchmark {
	case config.CIS_EKS, config.CIS_K8S:
		return getK8sDataProvider(ctx, log, cfg)
	default:
		return getAWSDataProvider(ctx, log, cfg)
	}
}

func getAWSDataProvider(ctx context.Context, log *logp.Logger, cfg config.Config) (dataprovider.CommonDataProvider, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, fmt.Errorf("Common.getAWSDataProvider failed to initialize AWS credentials: %w", err)
	}

	identityClient := awslib.GetIdentityClient(awsConfig)
	iamProvider := iam.NewIAMProvider(log, awsConfig)

	identity, err := identityClient.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("Common.getAWSDataProvider failed to get AWS identity: %w", err)
	}

	alias, err := iamProvider.GetAccountAlias(ctx)
	if err != nil {
		return nil, fmt.Errorf("Common.getAWSDataProvider failed to get AWS account alias: %w", err)
	}

	return aws_dataprovider.New(
		aws_dataprovider.WithLogger(log),
		aws_dataprovider.WithAccount(alias, *identity.Account),
	), nil
}

func getK8sDataProvider(ctx context.Context, log *logp.Logger, cfg config.Config) (dataprovider.CommonDataProvider, error) {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, err
	}

	clusterNameProvider := providers.ClusterNameProvider{
		KubernetesClusterNameProvider: providers.KubernetesClusterNameProvider{},
		EKSMetadataProvider:           awslib.Ec2MetadataProvider{},
		EKSClusterNameProvider:        awslib.EKSClusterNameProvider{},
		KubeClient:                    kubeClient,
		AwsConfigProvider: awslib.ConfigProvider{
			MetadataProvider: awslib.Ec2MetadataProvider{},
		},
	}
	name, err := clusterNameProvider.GetClusterName(ctx, &cfg, log)
	if err != nil {
		log.Errorf("Common.getK8sDataProvider failed to get cluster name: %v", err)
	}
	v, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("Common.getK8sDataProvider failed to get server version: %w", err)
	}
	node, err := kubernetes.DiscoverKubernetesNode(log, &kubernetes.DiscoverKubernetesNodeParams{
		ConfigHost:  "",
		Client:      kubeClient,
		IsInCluster: true,
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	})
	if err != nil {
		return nil, err
	}
	n, err := kubeClient.CoreV1().Nodes().Get(ctx, node, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Common.getK8sDataProvider failed to get node data: %w", err)
	}

	ns, err := kubeClient.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Common.getK8sDataProvider failed to get namespace data: %w", err)
	}
	options := []k8s_dataprovider.Option{
		k8s_dataprovider.WithConfig(&cfg),
		k8s_dataprovider.WithLogger(log),
		k8s_dataprovider.WithClusterName(name),
		k8s_dataprovider.WithClusterID(string(ns.ObjectMeta.UID)),
		k8s_dataprovider.WithNodeID(string(n.ObjectMeta.UID)),
		k8s_dataprovider.WithVersionInfo(version.CloudbeatVersionInfo{
			Version: version.CloudbeatVersion(),
			Policy:  version.PolicyVersion(),
			Kubernetes: version.Version{
				Version: v.Major + "." + v.Minor,
			},
		}),
	}
	return k8s_dataprovider.New(options...), nil
}
