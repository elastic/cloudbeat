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
	"time"
)

var (
	initOnce sync.Once
	callOnce sync.Once
	manager  *Manager
)

type ElectionManager interface {
	IsLeader() bool
	Run() error
}

type Manager struct {
	log    *logp.Logger
	ctx    context.Context
	client k8s.Interface
	leader *le.LeaderElector
	block  chan bool
}

// NewLeaderElector acts as a singleton - return & run an instance of the Election manager
func NewLeaderElector(ctx context.Context, log *logp.Logger, cfg config.Config) (ElectionManager, error) {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewLeaderElector error in GetClient: %v", err)
		return nil, err
	}

	initOnce.Do(func() {
		manager = &Manager{
			log:    log,
			ctx:    ctx,
			client: kubeClient,
			leader: nil,
			block:  make(chan bool),
		}
	})

	return manager, nil
}

func GetLeaderElectorManager() ElectionManager {
	return manager
}

func (m *Manager) IsLeader() bool {
	return m.leader.IsLeader()
}

// Run leader election is blocking until a leader is being elected or timeout has reached.
func (m *Manager) Run() error {
	var err error
	leaderElectionConf, err := m.buildConfig()
	if err != nil {
		return err
	}

	m.leader, err = le.NewLeaderElector(leaderElectionConf)
	if err != nil {
		return err
	}

	go m.leader.Run(m.ctx)
	m.log.Infof("started leader election, description: %s, id: %s ", leaderElectionConf.Lock.Describe(), leaderElectionConf.Lock.Identity())

	select {
	case <-m.block:
		m.log.Infof("new leader has been elected")
	case <-time.After(FirstLeaderDeadline):
		m.log.Warnf("timeout - no leader has been elected for %s seconds", FirstLeaderDeadline.String())
	}

	return nil
}

func (m *Manager) buildConfig() (le.LeaderElectionConfig, error) {
	podId, err := m.currentPodID()
	if err != nil {
		return le.LeaderElectionConfig{}, err
	}

	id := fmt.Sprintf("%s_%s", LeaderLeaseName, podId)
	ns, err := kubernetes.InClusterNamespace()
	if err != nil {
		return le.LeaderElectionConfig{}, err
	}

	lease := metav1.ObjectMeta{
		Name:      LeaderLeaseName,
		Namespace: ns,
	}

	return le.LeaderElectionConfig{
		Lock: &rl.LeaseLock{
			LeaseMeta: lease,
			Client:    m.client.CoordinationV1(),
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
				m.log.Infof("leader election lock LOST, id: %v", id)
				// leaderelection.Run stops in case the lease was released, we re-run the manager
				// to keep following leader status, even though, pod has lost the lease.
				go m.leader.Run(m.ctx)
				m.log.Infof("re-running leader elector")
			},
			OnNewLeader: func(identity string) {
				m.log.Infof("leader election lock has been acquired, id %v", identity)
				callOnce.Do(func() {
					m.block <- false
				})
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
	m.log.Warnf("Generating uuid as an identifier, UUID: ", uuid.String())
	return uuid.String(), err
}
