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

package benchmark

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/uniqueness"
)

type K8S struct {
	leaderElector uniqueness.Manager
}

func (k *K8S) Run(ctx context.Context) error {
	return k.leaderElector.Run(ctx)
}

func (k *K8S) InitRegistry(
	_ context.Context,
	log *logp.Logger,
	cfg *config.Config,
	ch chan fetching.ResourceInfo,
	dependencies *Dependencies,
) (registry.Registry, error) {
	kubeClient, err := dependencies.KubernetesClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client :%w", err)
	}
	k.leaderElector = uniqueness.NewLeaderElector(log, kubeClient)

	return registry.NewRegistry(log, factory.NewCisK8sFactory(log, ch, k.leaderElector, kubeClient)), nil
}

func (k *K8S) Stop() {
	k.leaderElector.Stop()
}
