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

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_gokubernetes "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	k8sprovider "github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/cloudbeat/version"
)

type K8S struct {
	ClientProvider k8sprovider.ClientGetterAPI

	leaderElector uniqueness.Manager
}

func (k *K8S) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, error) {
	if err := k.checkDependencies(); err != nil {
		return nil, nil, err
	}

	kubeClient, err := k.ClientProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes client :%w", err)
	}
	k.leaderElector = uniqueness.NewLeaderElector(log, kubeClient)

	dp, err := getK8sDataProvider(ctx, log, *cfg, kubeClient, k8s.KubernetesClusterNameProvider{KubeClient: kubeClient})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create k8s data provider: %w", err)
	}

	return registry.NewRegistry(log, factory.NewCisK8sFactory(log, ch, k.leaderElector, kubeClient)), dp, nil
}

func (k *K8S) Run(ctx context.Context) error { return k.leaderElector.Run(ctx) }
func (k *K8S) Stop()                         { k.leaderElector.Stop() }

func getK8sDataProvider(
	ctx context.Context,
	log *logp.Logger,
	cfg config.Config,
	kubeClient client_gokubernetes.Interface,
	clusterNameProvider k8s.ClusterNameProviderAPI,
) (dataprovider.CommonDataProvider, error) {
	clusterName, err := clusterNameProvider.GetClusterName(ctx, &cfg)
	if err != nil {
		log.Errorf("failed to get cluster name: %v", err)
	}

	serverVersion, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	nodeName, err := kubernetes.DiscoverKubernetesNode(log, &kubernetes.DiscoverKubernetesNodeParams{
		ConfigHost:  "",
		Client:      kubeClient,
		IsInCluster: true,
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get node name: %w", err)
	}

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node data for node '%s': %w", nodeName, err)
	}

	namespace, err := kubeClient.CoreV1().Namespaces().Get(ctx, "kube-system", v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace data: %w", err)
	}

	options := []k8s.Option{
		k8s.WithConfig(&cfg),
		k8s.WithLogger(log),
		k8s.WithClusterName(clusterName),
		k8s.WithClusterID(string(namespace.ObjectMeta.UID)),
		k8s.WithNodeID(string(node.ObjectMeta.UID)),
		k8s.WithVersionInfo(version.CloudbeatVersionInfo{
			Version: version.CloudbeatVersion(),
			Policy:  version.PolicyVersion(),
			Kubernetes: version.Version{
				Version: serverVersion.Major + "." + serverVersion.Minor,
			},
		}),
	}
	return k8s.New(options...), nil
}

func (k *K8S) checkDependencies() error {
	if k.ClientProvider == nil {
		return errors.New("kubernetes client provider is uninitialized")
	}
	return nil
}
