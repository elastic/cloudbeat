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

package k8s

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/elastic/cloudbeat/internal/config"
)

type KubernetesClusterNameProviderTestSuite struct {
	suite.Suite
}

func TestKubernetesClusterNameProviderTestSuite(t *testing.T) {
	s := new(KubernetesClusterNameProviderTestSuite)

	suite.Run(t, s)
}

func (s *KubernetesClusterNameProviderTestSuite) TestGetClusterName() {
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
	cfg := &config.Config{KubeConfig: "/dev/null"}
	client := fake.NewSimpleClientset(ns, cfgMap)
	provider := KubernetesClusterNameProvider{KubeClient: client}

	res, err := provider.GetClusterName(t.Context(), cfg)
	s.Require().NoError(err)
	s.Equal(clusterName, res)
}

func (s *KubernetesClusterNameProviderTestSuite) TestGetClusterMetadataNoClusterName() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	cfg := &config.Config{KubeConfig: "/dev/null"}
	client := fake.NewSimpleClientset(ns)
	provider := KubernetesClusterNameProvider{KubeClient: client}

	res, err := provider.GetClusterName(t.Context(), cfg)
	s.Empty(res)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "fail to resolve the name of the cluster")
}
