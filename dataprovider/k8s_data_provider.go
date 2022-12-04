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

package dataprovider

import (
	"context"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

type k8sDataCollector struct {
	kubeClient k8s.Interface
	log        *logp.Logger
	cfg        *config.Config
}

type k8sDataProvider interface {
	CollectK8sData(ctx context.Context) *CommonK8sData
}

func NewK8sDataProvider(log *logp.Logger, cfg *config.Config) k8sDataProvider {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewK8sDataProvider error in GetClient: %v", err)
	}

	return k8sDataCollector{
		kubeClient: kubeClient,
		log:        log,
		cfg:        cfg,
	}
}

func (k k8sDataCollector) CollectK8sData(ctx context.Context) *CommonK8sData {
	if k.kubeClient == nil {
		k.log.Warn("k8s in unavailable")
		return nil
	}

	return &CommonK8sData{
		clusterId:     k.getClusterId(ctx),
		nodeId:        k.getNodeId(ctx),
		serverVersion: k.fetchKubernetesVersion(),
	}
}

func (k k8sDataCollector) getClusterId(ctx context.Context) string {
	n, err := k.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		k.log.Errorf("getClusterId error in Namespaces get: %v", err)
		return ""
	}
	return string(n.ObjectMeta.UID)
}

func (k k8sDataCollector) getNodeId(ctx context.Context) string {
	nName, err := k.getNodeName()
	if err != nil {
		k.log.Errorf("getNodeId error in getNodeName: %v", err)
		return ""
	}
	n, err := k.kubeClient.CoreV1().Nodes().Get(ctx, nName, metav1.GetOptions{})
	if err != nil {
		k.log.Errorf("getClusterId error in Nodes get: %v", err)
		return ""
	}
	return string(n.ObjectMeta.UID)
}

func (k k8sDataCollector) getNodeName() (string, error) {
	nd := &kubernetes.DiscoverKubernetesNodeParams{
		// TODO: Add Host capability to Config
		ConfigHost:  "",
		Client:      k.kubeClient,
		IsInCluster: kubernetes.IsInCluster(k.cfg.KubeConfig),
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	}

	nName, err := kubernetes.DiscoverKubernetesNode(k.log, nd)
	if err != nil {
		k.log.Errorf("getNodeName error in DiscoverKubernetesNode: %v", err)
		return "", err
	}

	return nName, nil
}

// FetchKubernetesVersion returns the version of the Kubernetes server
func (k k8sDataCollector) fetchKubernetesVersion() version.Version {
	serverVersion, err := k.kubeClient.Discovery().ServerVersion()
	if err != nil {
		k.log.Errorf("fetchKubernetesVersion error in DiscoverK8sServerVersion: %v", err)
		return version.Version{}
	}
	return version.Version{
		Version: serverVersion.Major + "." + serverVersion.Minor,
	}
}
