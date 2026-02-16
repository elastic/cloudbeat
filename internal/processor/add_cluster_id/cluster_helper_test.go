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
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

type ClusterHelperTestSuite struct {
	suite.Suite
}

func TestClusterHelperTestSuite(t *testing.T) {
	s := new(ClusterHelperTestSuite)

	suite.Run(t, s)
}

func (s *ClusterHelperTestSuite) TestClusterId() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	client := fake.NewSimpleClientset(ns)
	sut, err := newClusterHelper(client)
	s.Require().NoError(err)

	s.Equal(kubeSystemNamespaceId, sut.ClusterId())
}

func (s *ClusterHelperTestSuite) TestClusterIdNotFound() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-sys",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	client := fake.NewSimpleClientset(ns)
	_, err := newClusterHelper(client)
	s.Require().Error(err)
}
