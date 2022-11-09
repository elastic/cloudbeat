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
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"

	"github.com/gofrs/uuid"

	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "kube-system"
)

var uuid_namespace uuid.UUID = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

func NewCommonDataProvider(log *logp.Logger, cfg config.Config) (CommonDataProvider, error) {
	KubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewCommonDataProvider error in GetClient: %v", err)
		return CommonDataProvider{}, err
	}

	clusterNameProvider := providers.ClusterNameProvider{
		KubernetesClusterNameProvider: providers.KubernetesClusterNameProvider{},
		EKSMetadataProvider:           awslib.Ec2MetadataProvider{},
		EKSClusterNameProvider:        awslib.EKSClusterNameProvider{},
		KubeClient:                    KubeClient,
		AwsConfigProvider:             awslib.ConfigProvider{},
	}

	return CommonDataProvider{
		log:                 log,
		kubeClient:          KubeClient,
		cfg:                 cfg,
		clusterNameProvider: clusterNameProvider,
	}, nil
}

// FetchCommonData Note: As of today Kubernetes is the only environment supported by CommonDataProvider
func (c CommonDataProvider) FetchCommonData(ctx context.Context) (CommonDataInterface, error) {
	cm := CommonData{}
	ClusterId, err := c.getClusterId(ctx)
	if err != nil {
		c.log.Errorf("fetchCommonData error in getClusterId: %v", err)
		return CommonData{}, err
	}
	cm.clusterId = ClusterId
	NodeId, err := c.getNodeId(ctx)
	if err != nil {
		c.log.Errorf("fetchCommonData error in getNodeId: %v", err)
		return CommonData{}, err
	}
	cm.nodeId = NodeId

	clusterName, err := c.clusterNameProvider.GetClusterName(ctx, c.cfg)
	if err != nil {
		c.log.Errorf("could not fetch cluster name", err)
	}
	cm.clusterName = clusterName

	return cm, nil
}

func (c CommonDataProvider) getClusterId(ctx context.Context) (string, error) {
	n, err := c.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		c.log.Errorf("getClusterId error in Namespaces get: %v", err)
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func (c CommonDataProvider) getNodeId(ctx context.Context) (string, error) {
	nName, err := c.getNodeName()
	if err != nil {
		c.log.Errorf("getNodeId error in getNodeName: %v", err)
		return "", err
	}
	n, err := c.kubeClient.CoreV1().Nodes().Get(ctx, nName, metav1.GetOptions{})
	if err != nil {
		c.log.Errorf("getClusterId error in Nodes get: %v", err)
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

	nName, err := kubernetes.DiscoverKubernetesNode(c.log, nd)
	if err != nil {
		c.log.Errorf("getNodeName error in DiscoverKubernetesNode: %v", err)
		return "", err
	}

	return nName, nil
}

func (cd CommonData) GetResourceId(metadata fetching.ResourceMetadata) string {
	switch metadata.Type {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		return uuid.NewV5(uuid_namespace, cd.clusterId+cd.nodeId+metadata.ID).String()
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		return uuid.NewV5(uuid_namespace, cd.clusterId+metadata.ID).String()
	default:
		return metadata.ID
	}
}

func (cd CommonData) GetClusterName() string {
	return cd.clusterName
}

func (cd CommonData) GetData() CommonData {
	return cd
}
