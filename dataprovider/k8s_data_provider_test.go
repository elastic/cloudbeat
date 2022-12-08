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
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"os"
	"testing"
)

var (
	ctxBg  = context.Background()
	logger = logp.NewLogger("cloudbeat_k8s_common_data_provider_test")
	cfg    = &config.Config{}
)

type clusterNameProviderMock struct {
	clusterName string
}

func (c clusterNameProviderMock) GetClusterName(_ context.Context, _ *config.Config) (string, error) {
	if c.clusterName == "error" {
		return "", os.ErrNotExist
	}
	return c.clusterName, nil
}

func Test_k8sDataCollector_CollectK8sData(t *testing.T) {
	tests := []struct {
		collector k8sDataCollector
		name      string
		want      *CommonK8sData
	}{
		{
			name: "test k8s common data",
			want: &CommonK8sData{
				clusterId: "testing_namespace_uid",
				nodeId:    "testing_node_uid",
				serverVersion: version.Version{
					Version: ".",
				},
				clusterName: "cluster_name",
			},
			collector: k8sDataCollector{
				kubeClient:          k8sFake.NewSimpleClientset(),
				log:                 logger,
				cfg:                 cfg,
				clusterNameProvider: clusterNameProviderMock{"cluster_name"},
			},
		},
		{
			name: "test k8s common data - error providing cluster name",
			want: &CommonK8sData{
				clusterId: "testing_namespace_uid",
				nodeId:    "testing_node_uid",
				serverVersion: version.Version{
					Version: ".",
				},
				clusterName: "",
			},
			collector: k8sDataCollector{
				kubeClient:          k8sFake.NewSimpleClientset(),
				log:                 logger,
				cfg:                 cfg,
				clusterNameProvider: clusterNameProviderMock{"error"},
			},
		},
		{
			name: "test k8s common data with no k8s connection",
			want: nil,
			collector: k8sDataCollector{
				kubeClient:          nil,
				log:                 logger,
				cfg:                 cfg,
				clusterNameProvider: clusterNameProviderMock{"some_name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustK8sCluster(t, &tt.collector)
			assert.Equal(t, tt.want, tt.collector.CollectK8sData(ctxBg))
		})
	}
}

func adjustK8sCluster(t *testing.T, k8sDataCollector *k8sDataCollector) {
	if k8sDataCollector.kubeClient == nil {
		return
	}

	// libbeat DiscoverKubernetesNode performs a fallback to environment variable NODE_NAME
	os.Setenv("NODE_NAME", "testing_node")
	// Need to add services
	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  "testing_namespace_uid",
		},
	}

	node := &v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "testing_node",
			UID:  "testing_node_uid",
		},
	}

	_, err := k8sDataCollector.kubeClient.CoreV1().Namespaces().Create(ctxBg, namespace, metav1.CreateOptions{})
	assert.NoError(t, err)

	_, err = k8sDataCollector.kubeClient.CoreV1().Nodes().Create(ctxBg, node, metav1.CreateOptions{})
	assert.NoError(t, err)
}
