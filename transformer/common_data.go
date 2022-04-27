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
	"io/ioutil"
	"strings"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonData struct {
	clusterId string
	nodeId string
}

// TODO: Consider moving this layer to be custom for every resource type
func NewCommonData() CommonData {
	return CommonData{
		clusterId: "",
		nodeId: "",
	}
}

// TODO: Support environments besides K8S
func (c *CommonData) fetchCommonData() error {
	clusterId, err := getClusterId()
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getClusterId: %w", err))
		return err
	}
	c.clusterId = clusterId
	nodeId, err := getNodeId()
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getNodeId: %w", err))
		return err
	}
	c.nodeId = nodeId
	return nil
}

func getClusterId() (string, error) {
	client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in GetKubernetesClient: %w", err))
		return "", err
	}
	n, err := client.CoreV1().Namespaces().Get(context.Background(), "kube-system", metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Namespaces get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func getNodeId() (string, error) {
	hName, err := getHostName()
	if err != nil {
		logp.Error(fmt.Errorf("getNodeId error in getHostName: %w", err))
		return "", err
	}
	client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getNodeId error in GetKubernetesClient: %w", err))
		return "", err
	}
	n, err := client.CoreV1().Nodes().Get(context.Background(), hName, metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Nodes get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func getHostName() (string, error) {
	hName, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		logp.Error(fmt.Errorf("getHostName error in ReadFile: %w", err))
		return "", err
	}
	return strings.TrimSpace(string(hName)), nil
}
