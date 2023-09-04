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
	"errors"
	"fmt"

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/uniqueness"
)

type K8S struct {
	ClientProvider k8s.ClientGetterAPI

	leaderElector uniqueness.Manager
}

func (k *K8S) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := k.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	kubeClient, err := k.ClientProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create kubernetes client :%w", err)
	}

	benchmarkHelper := NewK8sBenchmarkHelper(log, cfg, kubeClient)
	k.leaderElector = uniqueness.NewLeaderElector(log, kubeClient)

	dp, err := benchmarkHelper.GetK8sDataProvider(ctx, k8s.KubernetesClusterNameProvider{KubeClient: kubeClient})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s data provider: %w", err)
	}

	idp, err := benchmarkHelper.GetK8sIdProvider(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create k8s id provider: %w", err)
	}

	return registry.NewRegistry(log, factory.NewCisK8sFactory(log, ch, k.leaderElector, kubeClient)), dp, idp, nil
}

func (k *K8S) Run(ctx context.Context) error { return k.leaderElector.Run(ctx) }
func (k *K8S) Stop()                         { k.leaderElector.Stop() }

func (k *K8S) checkDependencies() error {
	if k.ClientProvider == nil {
		return errors.New("kubernetes client provider is uninitialized")
	}
	return nil
}
