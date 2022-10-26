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

package add_cluster_id

import (
	"context"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes/metadata"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s "k8s.io/client-go/kubernetes"
)

type ClusterHelper interface {
	GetClusterMetadata() ClusterMetadata
}

type ClusterMetadataProvider struct {
	metadata ClusterMetadata
	logger   *logp.Logger
}

type ClusterMetadata struct {
	clusterId   string
	clusterName string
}

func newClusterMetadataProvider(client k8s.Interface, cfg *agentconfig.C, logger *logp.Logger) (ClusterHelper, error) {
	clusterId, err := getClusterIdFromClient(client)
	if err != nil {
		return nil, err
	}

	clusterIdentifier, err := metadata.GetKubernetesClusterIdentifier(cfg, client)
	if err != nil {
		logger.Errorf("fail to resolve the name of the cluster, error %v", err)
	}
	return &ClusterMetadataProvider{metadata: ClusterMetadata{clusterId: clusterId, clusterName: clusterIdentifier.Name}, logger: logger}, nil
}

func (c *ClusterMetadataProvider) GetClusterMetadata() ClusterMetadata {
	return ClusterMetadata{clusterName: c.metadata.clusterName, clusterId: c.metadata.clusterId}
}

func getClusterIdFromClient(client k8s.Interface) (string, error) {
	n, err := client.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}
