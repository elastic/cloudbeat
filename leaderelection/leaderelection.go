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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package leaderelection

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
	"os"
	"strings"
	"sync"
)

var (
	initOnce sync.Once
	callOnce sync.Once
	manager  *Manager
)

type ElectionManager interface {
	IsLeader() bool
	Run(ctx context.Context) error
}

type Manager struct {
	log    *logp.Logger
	leader *le.LeaderElector
	block  chan bool
}

// NewLeaderElector acts as a singleton
func NewLeaderElector(ctx context.Context, log *logp.Logger, cfg config.Config) (ElectionManager, error) {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewLeaderElector error in GetClient: %v", err)
		return nil, err
	}

	initOnce.Do(func() {
		var leConfig le.LeaderElectionConfig
		var leader *le.LeaderElector

		block := make(chan bool)
		leConfig, err = buildConfig(ctx, log, kubeClient, block)
		leader, err = le.NewLeaderElector(leConfig)

		manager = &Manager{
			log:    log,
			leader: leader,
			block:  block,
		}
	})

	return manager, err
}

func GetLeaderElectorManager() ElectionManager {
	return manager
}

func (m *Manager) IsLeader() bool {
	return m.leader.IsLeader()
}

// Run leader election is blocking until a leader is being elected or timeout has reached.
func (m *Manager) Run(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, FirstLeaderDeadline)
	defer cancel()

	go m.leader.Run(ctx)
	m.log.Infof("started leader election")

	select {
	case <-m.block:
		m.log.Infof("new leader has been elected")
	case <-timeoutCtx.Done():
		m.log.Warnf("timeout - no leader has been elected for %s", FirstLeaderDeadline.String())
	}

	return nil
}

func buildConfig(ctx context.Context, log *logp.Logger, client k8s.Interface, block chan bool) (le.LeaderElectionConfig, error) {
	podId, err := currentPodID(log)
	if err != nil {
		return le.LeaderElectionConfig{}, err
	}

	id := fmt.Sprintf("%s_%s", LeaderLeaseName, podId)
	ns, err := kubernetes.InClusterNamespace()
	if err != nil {
		ns = "default"
	}

	lease := metav1.ObjectMeta{
		Name:      LeaderLeaseName,
		Namespace: ns,
	}

	return le.LeaderElectionConfig{
		Lock: &rl.LeaseLock{
			LeaseMeta: lease,
			Client:    client.CoordinationV1(),
			LockConfig: rl.ResourceLockConfig{
				Identity: id,
			},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   LeaseDuration,
		RenewDeadline:   RenewDeadline,
		RetryPeriod:     RetryPeriod,
		Callbacks: le.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Infof("leader election lock GAINED, id: %v", id)
			},
			OnStoppedLeading: func() {
				log.Infof("leader election lock LOST, id: %v", id)
				// leaderelection.Run stops in case it has stopped holding the leader lease, we re-run the manager
				// to keep following leader status.
				go manager.leader.Run(ctx)
				log.Infof("re-running leader elector")
			},
			OnNewLeader: func(identity string) {
				log.Infof("leader election lock has been acquired, id: %v", identity)
				callOnce.Do(func() {
					defer close(block)
					block <- false
				})
			},
		},
	}, nil
}

func currentPodID(log *logp.Logger) (string, error) {
	pod, found := os.LookupEnv(PodNameEnvar)
	if !found {
		log.Warnf("Env var %s wasn't found", PodNameEnvar)
		return generateUUID(log)
	}

	return lastPart(log, pod)
}

func lastPart(log *logp.Logger, s string) (string, error) {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		log.Warnf("failed to find id for pod_name: %s", s)
		return generateUUID(log)
	}

	return parts[len(parts)-1], nil
}

func generateUUID(log *logp.Logger) (string, error) {
	uuid, err := uuid.NewV4()
	log.Warnf("Generating uuid as an identifier, UUID: ", uuid.String())
	return uuid.String(), err
}
