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
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	namespace        = "kube-system"
	clusterNameField = "orchestrator.cluster.name"
)

var uuidNamespace = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

type k8sDataProvider struct {
	kubeClient          k8s.Interface
	log                 *logp.Logger
	cfg                 *config.Config
	clusterNameProvider providers.ClusterNameProviderAPI
}

type commonK8sData struct {
	clusterId     string
	nodeId        string
	serverVersion version.Version
	clusterName   string
}

func NewK8sDataProvider(log *logp.Logger, cfg *config.Config) EnvironmentCommonDataProvider {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Warnf("Could not create Kubernetes client to provide common data: %v", err)
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

	return k8sDataProvider{
		kubeClient:          kubeClient,
		log:                 log,
		cfg:                 cfg,
		clusterNameProvider: clusterNameProvider,
	}
}

func (k k8sDataProvider) GetData(ctx context.Context) (CommonData, error) {
	if k.kubeClient == nil {
		k.log.Debug("Could not collect Kubernetes common data as the client was not provided")
		return nil, nil
	}

	return &commonK8sData{
		clusterId:     k.getClusterId(ctx),
		nodeId:        k.getNodeId(ctx),
		serverVersion: k.fetchKubernetesVersion(),
		clusterName:   k.getClusterName(ctx),
	}, nil
}

func (k k8sDataProvider) getClusterId(ctx context.Context) string {
	n, err := k.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		k.log.Errorf("getClusterId error in Namespaces get: %v", err)
		return ""
	}
	return string(n.ObjectMeta.UID)
}

func (k k8sDataProvider) getClusterName(ctx context.Context) string {
	clusterName, err := k.clusterNameProvider.GetClusterName(ctx, k.cfg, k.log)
	if err != nil {
		k.log.Errorf("cloud not identify the cluster name: %v", err)
		return ""
	}
	return clusterName
}

func (k k8sDataProvider) getNodeId(ctx context.Context) string {
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

func (k k8sDataProvider) getNodeName() (string, error) {
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
func (k k8sDataProvider) fetchKubernetesVersion() version.Version {
	serverVersion, err := k.kubeClient.Discovery().ServerVersion()
	if err != nil {
		k.log.Errorf("fetchKubernetesVersion error in DiscoverK8sServerVersion: %v", err)
		return version.Version{}
	}
	return version.Version{
		Version: serverVersion.Major + "." + serverVersion.Minor,
	}
}

func (c commonK8sData) GetResourceId(metadata fetching.ResourceMetadata) string {
	switch metadata.Type {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		return uuid.NewV5(uuidNamespace, c.clusterId+c.nodeId+metadata.ID).String()
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		return uuid.NewV5(uuidNamespace, c.clusterId+metadata.ID).String()
	default:
		return metadata.ID
	}
}

func (c commonK8sData) GetVersionInfo() version.CloudbeatVersionInfo {
	return version.CloudbeatVersionInfo{
		Version:    version.CloudbeatVersion(),
		Policy:     version.PolicyVersion(),
		Kubernetes: c.serverVersion,
	}
}

func (c commonK8sData) EnrichEvent(event beat.Event) error {
	_, err := event.Fields.Put(clusterNameField, c.clusterName)
	return err
}
