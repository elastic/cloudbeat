package leaderelection

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
	"os"
	"strings"
	"sync"
)

var (
	callOnce sync.Once
)

const ()

type ElectionManager interface {
	Run(onNewLeader func(ctx context.Context) error) error
	IsLeader() bool
}

type Manager struct {
	log    *logp.Logger
	ctx    context.Context
	client k8s.Interface
	leader *le.LeaderElector
}

func NewLeaderElector(ctx context.Context, log *logp.Logger, cfg config.Config) (ElectionManager, error) {
	kubeClient, err := providers.KubernetesProvider{}.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("NewLeaderElector error in GetClient: %v", err)
		return nil, err
	}

	return &Manager{
		ctx:    ctx,
		log:    log,
		client: kubeClient,
		leader: nil,
	}, nil
}

// Run runs leader election and calls onStarted when lease has been acquired
func (m *Manager) Run(onNewLeader func(ctx context.Context) error) error {
	var err error
	leaderElectionConf := m.getConfig(onNewLeader)

	m.leader, err = le.NewLeaderElector(leaderElectionConf)
	if err != nil {
		return err
	}

	go m.leader.Run(m.ctx)
	m.log.Infof("started leader election", "for", leaderElectionConf.Lock.Describe(), "id", leaderElectionConf.Lock.Identity())

	return nil
}

func (m *Manager) IsLeader() bool {
	return m.leader.IsLeader()
}

func (m *Manager) getConfig(onNewLeaderCb func(ctx context.Context) error) le.LeaderElectionConfig {
	id := fmt.Sprintf("%s_%s", LeaderLeaseName, currentPodID())
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
				m.log.Infof("leader election lock GAINED, id %v", id)

			},
			OnStoppedLeading: func() {
				m.log.Infof("leader election lock LOST, id %v", id)
				// leaderelection.Run stops in case the lease was released, we re-run the manager
				// to keep following leader status, even though, we lost the lease for some reason.
				if err := m.Run(onNewLeaderCb); err != nil {
					m.log.Errorf("failed to re-run the leader elector, Error: %v", err)
				}
			},
			OnNewLeader: func(identity string) {
				m.log.Infof("leader election lock has been acquired, id %v", identity)
				callOnce.Do(func() {
					if err := onNewLeaderCb(m.ctx); err != nil {
						m.log.Error(err)
					}
				})
			},
		},
	}
}

func currentPodID() string {
	pod := os.Getenv(PodNameEnvar)

	return lastPart(pod)
}

func lastPart(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
