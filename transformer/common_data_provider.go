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
	"io/fs"
	"os"
	"strings"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ( 
	hostNamePath = "/etc/"
	hostNameFile = "hostname"
	namespace = "kube-system"
)

func NewCommonDataProvider(cfg config.Config) (CommonDataProvider, error) {
	KubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("NewCommonDataProvider error in GetClient: %w", err))
		return CommonDataProvider{}, err
	}

	return CommonDataProvider{
		kubeClient: KubeClient,
		fsys: os.DirFS(hostNamePath),
	}, nil
}

// TODO: Support environments besides K8S
func (c CommonDataProvider) fetchCommonData(ctx context.Context) (CommonDataInterface, error) {
	cm := CommonData{}
	ClusterId, err := c.getClusterId(ctx)
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getClusterId: %w", err))
		return CommonData{}, err
	}
	cm.clusterId = ClusterId
	NodeId, err := c.getNodeId(ctx)
	if err != nil {
		logp.Error(fmt.Errorf("fetchCommonData error in getNodeId: %w", err))
		return CommonData{}, err
	}
	cm.nodeId = NodeId
	return cm, nil
}

func (c CommonDataProvider) getClusterId(ctx context.Context) (string, error) {
	n, err := c.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Namespaces get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func (c CommonDataProvider) getNodeId(ctx context.Context) (string, error) {
	hName, err := c.getHostName()
	if err != nil {
		logp.Error(fmt.Errorf("getNodeId error in getHostName: %w", err))
		return "", err
	}
	n, err := c.kubeClient.CoreV1().Nodes().Get(ctx, hName, metav1.GetOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("getClusterId error in Nodes get: %w", err))
		return "", err
	}
	return string(n.ObjectMeta.UID), nil
}

func (c CommonDataProvider) getHostName() (string, error) {
    hName, err := fs.ReadFile(c.fsys, hostNameFile)
	if err != nil {
		logp.Error(fmt.Errorf("getHostName error in ReadFile: %w", err))
		return "", err
	}
	return strings.TrimSpace(string(hName)), nil
}
