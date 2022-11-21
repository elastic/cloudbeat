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

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "kube-system"
)

var uuidNamespace = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

func NewCommonDataProvider(log *logp.Logger, cfg *config.Config) CommonDataProvider {
	k8sAvailable := true
	KubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		k8sAvailable = false
		log.Errorf("NewCommonDataProvider error in GetClient: %v", err)
	}

	return CommonDataProvider{
		log:          log,
		kubeClient:   KubeClient,
		cfg:          cfg,
		k8sAvailable: k8sAvailable,
	}
}

// FetchCommonData fetches cluster and node id and version info
// Note: As of today Kubernetes is the only environment supported by CommonDataProvider
func (c CommonDataProvider) FetchCommonData(ctx context.Context) (CommonDataInterface, error) {
	cm := CommonData{}
	versionInfo, err := c.FetchVersionInfo()
	if err != nil {
		c.log.Errorf("fetchCommonData error in FetchKubernetesVersion: %v", err)
	}
	cm.versionInfo = versionInfo

	if c.kubeClient == nil {
		c.log.Warn("k8s is unavailable")
		return cm, nil
	}

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

// FetchKubernetesVersion returns the version of the Kubernetes server
func (c CommonDataProvider) FetchKubernetesVersion() (version.Version, error) {
	serverVersion, err := c.kubeClient.Discovery().ServerVersion()
	if err != nil {
		return version.Version{}, err
	}
	return version.Version{
		Version: serverVersion.Major + "." + serverVersion.Minor,
	}, nil
}

func (c CommonDataProvider) FetchVersionInfo() (version.CloudbeatVersionInfo, error) {
	cloudbeatVersion := version.CloudbeatVersion()
	policyVersion := version.PolicyVersion()
	if !c.k8sAvailable {
		c.log.Warn("K8s is unavailable")
		return version.CloudbeatVersionInfo{
			Version: cloudbeatVersion,
			Policy:  policyVersion,
		}, nil
	}

	serverVersion, err := c.FetchKubernetesVersion()
	return version.CloudbeatVersionInfo{
		Version:    cloudbeatVersion,
		Policy:     policyVersion,
		Kubernetes: serverVersion,
	}, err
}

func (cd CommonData) GetResourceId(metadata fetching.ResourceMetadata) string {
	switch metadata.Type {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		return uuid.NewV5(uuidNamespace, cd.clusterId+cd.nodeId+metadata.ID).String()
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		return uuid.NewV5(uuidNamespace, cd.clusterId+metadata.ID).String()
	default:
		return metadata.ID
	}
}

func (cd CommonData) GetData() CommonData {
	return cd
}

func (cd CommonData) GetVersionInfo() version.CloudbeatVersionInfo {
	return cd.versionInfo
}
