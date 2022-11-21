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
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
)

type ElectionManager interface {
	IsLeader() bool
	Run(ctx context.Context) error
	Stop()
}

type Manager struct {
	log        *logp.Logger
	leader     *le.LeaderElector
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
	kubeClient k8s.Interface
}

func NewLeaderElector(log *logp.Logger, cfg *config.Config) ElectionManager {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewLeaderElector error in GetClient: %v", err)
	}
	wg := &sync.WaitGroup{}

	return &Manager{
		log:        log,
		kubeClient: kubeClient,
		leader:     nil,
		cancelFunc: nil,
		wg:         wg,
	}
}

func (m *Manager) IsLeader() bool {
	return m.leader.IsLeader()
}

// Run leader election is blocking until a FirstLeaderDeadline timeout has reached.
func (m *Manager) Run(ctx context.Context) error {
	if m.kubeClient == nil {
		m.log.Warnf("K8s is unavailable")
		return nil
	}

	newCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	leConfig, err := m.buildConfig(newCtx)
	if err != nil {
		m.log.Errorf("Fail building leader election config: %v", err)
		return err
	}

	m.leader, err = le.NewLeaderElector(leConfig)
	if err != nil {
		m.log.Errorf("Fail to create a new leader elector: %v", err)
		return err
	}

	go m.leader.Run(newCtx)
	m.wg.Add(1)
	m.log.Infof("started leader election")

	time.Sleep(FirstLeaderDeadline)
	m.log.Infof("stop waiting after %s for a leader to be elected", FirstLeaderDeadline)

	return nil
}

func (m *Manager) Stop() {
	if m.cancelFunc != nil {
		m.log.Info("Stopping leader election manager")
		m.cancelFunc()
		m.wg.Wait()
		return
	}

	m.log.Warnf("cancelFunc is not set")
}

func (m *Manager) buildConfig(ctx context.Context) (le.LeaderElectionConfig, error) {
	podId, err := m.currentPodID()
	if err != nil {
		return le.LeaderElectionConfig{}, err
	}

	id := fmt.Sprintf("%s_%s", LeaderLeaseName, podId)
	ns, err := kubernetes.InClusterNamespace()
	if err != nil {
		ns = v1.NamespaceDefault
	}

	lease := metav1.ObjectMeta{
		Name:      LeaderLeaseName,
		Namespace: ns,
	}

	return le.LeaderElectionConfig{
		Lock: &rl.LeaseLock{
			LeaseMeta: lease,
			Client:    m.kubeClient.CoordinationV1(),
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
				m.log.Infof("leader election lock GAINED, id: %v", id)
			},
			OnStoppedLeading: func() {
				// OnStoppedLeading gets called even if cloudbeat wasn't the leader, for example, if the context is cancelled due to reconfiguration from fleet.
				// We re-run the manager to keep following leader status except for context cancellation events.
				m.log.Infof("leader election lock LOST, id: %v", id)
				defer m.wg.Done()

				select {
				case <-ctx.Done():
					m.log.Info("Context is cancelled - should not re-run leader election")
					return
				default:
					go m.leader.Run(ctx)
					m.wg.Add(1)
					m.log.Infof("re-running leader elector")
				}
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					m.log.Infof("leader election lock has been acquired by this pod, id: %v", identity)
				} else {
					m.log.Infof("leader election lock has been acquired by another pod, id: %v", identity)
				}
			},
		},
	}, nil
}

func (m *Manager) currentPodID() (string, error) {
	pod, found := os.LookupEnv(PodNameEnvar)
	if !found {
		m.log.Warnf("Env var %s wasn't found", PodNameEnvar)
		return m.generateUUID()
	}

	return m.lastPart(pod)
}

func (m *Manager) lastPart(s string) (string, error) {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		m.log.Warnf("failed to find id for pod_name: %s", s)
		return m.generateUUID()
	}

	return parts[len(parts)-1], nil
}

func (m *Manager) generateUUID() (string, error) {
	uuid, err := uuid.NewV4()
	m.log.Warnf("Generating uuid as an identifier, UUID: %s", uuid.String())
	return uuid.String(), err
}
