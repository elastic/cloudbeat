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

package providers

import (
	"fmt"

	"github.com/elastic/elastic-agent-autodiscover/kubernetes/metadata"
	agentcfg "github.com/elastic/elastic-agent-libs/config"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/config"
)

type KubernetesClusterNameProviderApi interface {
	GetClusterName(cfg *config.Config, client k8s.Interface) (string, error)
}
type KubernetesClusterNameProvider struct {
}

func (provider KubernetesClusterNameProvider) GetClusterName(cfg *config.Config, client k8s.Interface) (string, error) {
	agentConfig, err := agentcfg.NewConfigFrom(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create agent config: %v", err)
	}
	clusterIdentifier, err := metadata.GetKubernetesClusterIdentifier(agentConfig, client)
	if err != nil {
		return "", fmt.Errorf("fail to resolve the name of the cluster: %v", err)
	}

	return clusterIdentifier.Name, nil
}
