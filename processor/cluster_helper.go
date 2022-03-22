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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
)

type ClusterHelper struct {
	clusterId string
}

func newClusterHelper() (*ClusterHelper, error) {
	clusterId, err := getClusterIdFromClient()
	if err != nil {
		return nil, err
	}
	return &ClusterHelper{clusterId: clusterId}, nil
}

func (c ClusterHelper) ClusterId() string {
	return c.clusterId
}

func getClusterIdFromClient() (string, error) {
	client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
	if err != nil {
		return "", err
	}
	n, err := client.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}
