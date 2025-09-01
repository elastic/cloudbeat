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

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1" // revive:disable-line
	client_gokubernetes "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type K8SBenchmarkHelper struct {
	log    *clog.Logger
	cfg    *config.Config
	client client_gokubernetes.Interface
}

func NewK8sBenchmarkHelper(log *clog.Logger, cfg *config.Config, client client_gokubernetes.Interface) *K8SBenchmarkHelper {
	return &K8SBenchmarkHelper{
		log:    log,
		cfg:    cfg,
		client: client,
	}
}

func (h *K8SBenchmarkHelper) GetK8sDataProvider(ctx context.Context, clusterNameProvider k8s.ClusterNameProviderAPI) (dataprovider.CommonDataProvider, error) {
	clusterName, err := clusterNameProvider.GetClusterName(ctx, h.cfg)
	if err != nil {
		h.log.Errorf(ctx, "failed to get cluster name: %v", err)
	}

	serverVersion, err := h.client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	clusterId, err := h.getK8sClusterId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster id: %w", err)
	}

	options := []k8s.Option{
		k8s.WithClusterName(clusterName),
		k8s.WithClusterID(clusterId),
		k8s.WithClusterVersion(serverVersion.String()),
	}
	return k8s.New(options...), nil
}

func (h *K8SBenchmarkHelper) GetK8sIdProvider(ctx context.Context) (dataprovider.IdProvider, error) {
	nodeId, err := h.getK8sNodeId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node id: %w", err)
	}

	clusterId, err := h.getK8sClusterId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster id: %w", err)
	}

	return k8s.NewIdProvider(clusterId, nodeId), nil
}

func (h *K8SBenchmarkHelper) getK8sClusterId(ctx context.Context) (string, error) {
	namespace, err := h.client.CoreV1().Namespaces().Get(ctx, "kube-system", v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get namespace data: %w", err)
	}

	return string(namespace.ObjectMeta.UID), nil
}

func (h *K8SBenchmarkHelper) getK8sNodeId(ctx context.Context) (string, error) {
	nodeName, err := kubernetes.DiscoverKubernetesNode(h.log.Logger, &kubernetes.DiscoverKubernetesNodeParams{
		ConfigHost:  "",
		Client:      h.client,
		IsInCluster: true,
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get node name: %w", err)
	}

	node, err := h.client.CoreV1().Nodes().Get(ctx, nodeName, v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get node data for node '%s': %w", nodeName, err)
	}

	return string(node.ObjectMeta.UID), nil
}
