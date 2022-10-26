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
	"fmt"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

type ClusterMetadataProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestClusterMetadataProviderTestSuite(t *testing.T) {
	s := new(ClusterMetadataProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_cluster_metadata_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ClusterMetadataProviderTestSuite) TestGetClusterMetadata() {
	kubeSystemNamespaceId := "123"
	clusterName := "my-cluster-name"
	configMapId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	cfgMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeadm-config",
			Namespace: "kube-system",
			UID:       types.UID(configMapId),
		},
		Immutable: nil,
		Data: map[string]string{
			"ClusterConfiguration": fmt.Sprintf("clusterName: %s", clusterName),
		},
		BinaryData: nil,
	}

	cfg := agentconfig.NewConfig()
	client := fake.NewSimpleClientset(ns, cfgMap)
	provider, err := newClusterMetadataProvider(client, cfg, s.log)
	s.NoError(err)

	res := provider.GetClusterMetadata()
	s.Equal(kubeSystemNamespaceId, res.clusterId)
	s.Equal(clusterName, res.clusterName)
}

func (s *ClusterMetadataProviderTestSuite) TestGetClusterMetadataNoClusterName() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	cfg := agentconfig.NewConfig()
	client := fake.NewSimpleClientset(ns)
	provider, err := newClusterMetadataProvider(client, cfg, s.log)
	s.NoError(err)

	res := provider.GetClusterMetadata()
	s.Equal(kubeSystemNamespaceId, res.clusterId)
	s.Equal("", res.clusterName)
}

func (s *ClusterMetadataProviderTestSuite) TestGetClusterMetadataClusterIdNotFound() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-sys",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	client := fake.NewSimpleClientset(ns)
	cfg := agentconfig.NewConfig()
	_, err := newClusterMetadataProvider(client, cfg, s.log)
	s.Error(err)
}
