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

package factory

import (
	"context"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
)

var vanillaFsPatterns = []string{
	"/hostfs/etc/kubernetes/scheduler.conf",
	"/hostfs/etc/kubernetes/controller-manager.conf",
	"/hostfs/etc/kubernetes/admin.conf",
	"/hostfs/etc/kubernetes/kubelet.conf",
	"/hostfs/etc/kubernetes/manifests/etcd.yaml",
	"/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
	"/hostfs/etc/kubernetes/manifests/kube-controller-manager.yaml",
	"/hostfs/etc/kubernetes/manifests/kube-scheduler.yaml",
	"/hostfs/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
	"/hostfs/etc/kubernetes/pki/*",
	"/hostfs/var/lib/kubelet/config.yaml",
	"/hostfs/var/lib/etcd",
	"/hostfs/etc/kubernetes/pki",
}

var vanillaRequiredProcesses = fetchers.ProcessesConfigMap{
	"etcd":            {},
	"kube-apiserver":  {},
	"kube-controller": {},
	"kube-scheduler":  {},
	"kubelet":         {ConfigFileArguments: []string{"config"}},
}

func NewCisK8sFactory(_ context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (FetchersMap, error) {
	log.Infof("Initializing K8s fetchers")
	m := make(FetchersMap)
	fsFetcher := fetchers.NewFsFetcher(log, ch, vanillaFsPatterns)
	m[fetching.FileSystemType] = fsFetcher

	procFetcher := fetchers.NewProcessFetcher(log, ch, vanillaRequiredProcesses)
	m[fetching.ProcessType] = procFetcher

	k8sProvider, err := providers.GetK8sClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, err
	}

	kubeFetcher := fetchers.NewKubeFetcher(log, ch, k8sProvider)
	m[fetching.KubeAPIType] = kubeFetcher

	return m, nil
}
