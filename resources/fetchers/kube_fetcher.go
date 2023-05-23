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

package fetchers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	allNamespaces = "" // The Kube API treats this as "all namespaces"
)

var (
	// vanillaClusterResources represents those resources that are required for a vanilla
	// Kubernetes cluster.
	vanillaClusterResources = []requiredResource{
		{
			&kubernetes.Pod{},
			allNamespaces,
		},
		{
			&kubernetes.Role{},
			allNamespaces,
		},
		{
			&kubernetes.RoleBinding{},
			allNamespaces,
		},
		{
			&kubernetes.ClusterRole{},
			allNamespaces,
		},
		{
			&kubernetes.ClusterRoleBinding{},
			allNamespaces,
		},
		{
			&kubernetes.PodSecurityPolicy{},
			allNamespaces,
		},
		{
			&kubernetes.ServiceAccount{},
			allNamespaces,
		},
		{
			&kubernetes.Node{},
			allNamespaces,
		},
		// TODO(yashtewari): Problem: github.com/elastic/beats/vendor/k8s.io/apimachinery/pkg/api/errors/errors.go#401
		// > "the server could not find the requested resource"
		// {
		// 	&kubernetes.NetworkPolicy{},
		// 	allNamespaces,
		// },
	}
)

type requiredResource struct {
	resource  kubernetes.Resource
	namespace string
}

type KubeFetcher struct {
	log        *logp.Logger
	cfg        KubeApiFetcherConfig
	resourceCh chan fetching.ResourceInfo

	watchers       []kubernetes.Watcher
	clientProvider func(string, kubernetes.KubeClientOptions) (k8s.Interface, error)
	watcherLock    *sync.Once
}

type KubeApiFetcherConfig struct {
	fetching.BaseFetcherConfig
	Interval   time.Duration `config:"interval"`
	KubeConfig string        `config:"kubeconfig"`
}

func (f *KubeFetcher) initWatcher(client k8s.Interface, r requiredResource) error {
	f.cfg.Interval = time.Duration(time.Duration.Seconds(30)) // todo: hard coded - need to get from config

	watcher, err := kubernetes.NewWatcher(client, r.resource, kubernetes.WatchOptions{
		SyncTimeout: f.cfg.Interval,
		Namespace:   r.namespace,
	}, nil)
	if err != nil {
		return fmt.Errorf("could not create watcher: %w", err)
	}

	// TODO(yashtewari): it appears that Start never returns in case of certain failures, for example
	// if the configured Client's Role does not have the necessary permissions to list the Resource
	// being watched. This needs to be handled.
	//
	// When such a failure happens, cloudbeat won't shut down gracefully, i.e. Stop will not work. This
	// happens due to a context.TODO present in the libbeat dependency. It needs to accept context
	// from the caller instead.
	if err := watcher.Start(); err != nil {
		return fmt.Errorf("could not start watcher: %w", err)
	}

	f.watchers = append(f.watchers, watcher)

	return nil
}

func (f *KubeFetcher) initWatchers() error {
	client, err := f.clientProvider(f.cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return fmt.Errorf("could not get k8s client: %w", err)
	}

	f.log.Info("Kubernetes client initiated")

	f.watchers = make([]kubernetes.Watcher, 0)

	for _, r := range vanillaClusterResources {
		err := f.initWatcher(client, r)
		if err != nil {
			return err
		}
	}

	f.log.Info("Kubernetes Watchers initiated")

	return nil
}

func (f *KubeFetcher) Fetch(_ context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting KubeFetcher.Fetch")

	var err error
	f.watcherLock.Do(func() {
		err = f.initWatchers()
	})
	if err != nil {
		// Reset watcherLock if the watchers could not be initiated.
		f.watcherLock = &sync.Once{}
		return fmt.Errorf("could not initate Kubernetes watchers: %w", err)
	}

	getKubeData(f.log, f.watchers, f.resourceCh, cMetadata)
	return nil
}

func (f *KubeFetcher) Stop() {
	for _, watcher := range f.watchers {
		watcher.Stop()
	}
}

// addTypeInformationToKubeResource adds TypeMeta information to a kubernetes.Resource based upon the loaded scheme.Scheme
// inspired by: https://github.com/kubernetes/cli-runtime/blob/v0.19.2/pkg/printers/typesetter.go#L41
func addTypeInformationToKubeResource(resource kubernetes.Resource) error {
	groupVersionKinds, _, err := scheme.Scheme.ObjectKinds(resource)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, groupVersionKind := range groupVersionKinds {
		if len(groupVersionKind.Kind) == 0 {
			continue
		}
		if len(groupVersionKind.Version) == 0 || groupVersionKind.Version == runtime.APIVersionInternal {
			continue
		}
		resource.GetObjectKind().SetGroupVersionKind(groupVersionKind)
		break
	}

	return nil
}
