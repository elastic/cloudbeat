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

package transformer

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ( 
	namespace = "kube-system"
)
var uuid_namespace uuid.UUID = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

func NewCommonDataProvider(cfg config.Config) (CommonDataProvider, error) {
	KubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("NewCommonDataProvider error in GetClient: %w", err))
		return CommonDataProvider{}, err
	}

	return CommonDataProvider{
		kubeClient: KubeClient,
		cfg: cfg,
	}, nil
}

// Note: As of today Kubernetes is the only environment supported by CommonDataProvider
func (c CommonDataProvider) FetchCommonData(ctx context.Context) (CommonDataInterface, error) {
	cm := CommonData{}
	ClusterId, err := c.getClusterId(ctx)
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getClusterId: %w", err))
		return CommonData{}, err
	}
	cm.clusterId = ClusterId
	NodeId, err := c.getNodeId(ctx)
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getNodeId: %w", err))
		return CommonData{}, err
	}
	cm.nodeId = NodeId
	return cm, nil
}

func (c CommonDataProvider) getClusterId(ctx context.Context) (string, error) {
	n, err := c.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Namespaces get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func (c CommonDataProvider) getNodeId(ctx context.Context) (string, error) {
	nName, err := c.getNodeName()
	if err != nil {
		logp.Error(fmt.Errorf("getNodeId error in getNodeName: %w", err))
		return "", err
	}
	n, err := c.kubeClient.CoreV1().Nodes().Get(ctx, nName, metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Nodes get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func (c CommonDataProvider) getNodeName() (string, error) {
	nd := &kubernetes.DiscoverKubernetesNodeParams{
		// TODO: Add Host capability to Config
		ConfigHost:  "",
		Client:      c.kubeClient,
		IsInCluster: kubernetes.IsInCluster(c.cfg.KubeConfig),
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	}

	nName, err := kubernetes.DiscoverKubernetesNode(logp.L(), nd)
	if err != nil {
		logp.Error(fmt.Errorf("getNodeName error in DiscoverKubernetesNode: %w", err))
		return "", err
	}

	return nName, nil
}

func (cd CommonData) GetResourceId(id string) string {
	rid := cd.clusterId + cd.nodeId + id
	return uuid.NewV5(uuid_namespace, rid).String()
}

func (cd CommonData) GetData() CommonData {
	return cd
}
