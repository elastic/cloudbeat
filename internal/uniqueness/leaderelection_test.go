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

package uniqueness

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"k8s.io/api/coordination/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type LeaderElectionTestSuite struct {
	suite.Suite
	wg         *sync.WaitGroup
	manager    *LeaderelectionManager
	opts       goleak.Option
	kubeClient *k8sFake.Clientset
}

func TestLeaderElectionTestSuite(t *testing.T) {
	testhelper.SkipLong(t)

	s := new(LeaderElectionTestSuite)
	suite.Run(t, s)
}

func (s *LeaderElectionTestSuite) SetupTest() {
	s.wg = &sync.WaitGroup{}
	s.opts = goleak.IgnoreCurrent()
	s.kubeClient = k8sFake.NewSimpleClientset()
	s.manager = &LeaderelectionManager{
		log:        testhelper.NewLogger(s.T()),
		leader:     nil,
		wg:         s.wg,
		cancelFunc: nil,
		kubeClient: s.kubeClient,
	}
}

func (s *LeaderElectionTestSuite) TearDownTest() {
	// Stop is blocking until all go routines are finished,
	// we verify there is no running leader-election managers after calling stop.
	s.manager.Stop()

	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *LeaderElectionTestSuite) TestManager_RunWaitForLeader() {
	sTime := time.Now()
	t := s.T()
	err := s.manager.Run(t.Context())
	elapsed := time.Since(sTime)

	s.Require().NoError(err)
	s.GreaterOrEqual(elapsed, FirstLeaderDeadline, "run did not wait a sufficient time to acquire the lease")
	s.True(s.manager.IsLeader())
}

// Verify that when a pre-configured lease exists, eventually, the leader-manager will try to gain control if the
// lease is not being renewed.
func (s *LeaderElectionTestSuite) TestManager_RunWithExistingLease() {
	podId := "this_pod"
	s.Require().NoError(os.Setenv(PodNameEnvar, podId))

	holderIdentity := LeaderLeaseName + "_another_pod"
	lease := generateLease(&holderIdentity)
	s.manager.kubeClient = k8sFake.NewSimpleClientset(lease)
	t := s.T()
	err := s.manager.Run(t.Context())
	s.Require().NoError(err)

	updatedLease, err := s.manager.kubeClient.CoordinationV1().Leases(core.NamespaceDefault).Get(
		t.Context(),
		LeaderLeaseName,
		metav1.GetOptions{},
	)

	s.Require().NoError(err)
	s.Contains(*updatedLease.Spec.HolderIdentity, podId)
}

// Verify that after the lease is lost we re-run the leader-election manager.
// After waiting for a FirstLeaderDeadline seconds we should be holding the lease again as it has not been renewed.
func (s *LeaderElectionTestSuite) TestManager_ReRun() {
	podId := "this_pod"
	s.Require().NoError(os.Setenv(PodNameEnvar, podId))

	s.manager.kubeClient = k8sFake.NewSimpleClientset()
	t := s.T()
	err := s.manager.Run(t.Context())
	s.Require().NoError(err)

	holderIdentity := LeaderLeaseName + "_another_pod"
	lease := generateLease(&holderIdentity)
	_, err = s.manager.kubeClient.CoordinationV1().Leases(core.NamespaceDefault).Update(
		t.Context(),
		lease,
		metav1.UpdateOptions{},
	)

	time.Sleep(FirstLeaderDeadline)
	s.Require().NoError(err)
	s.True(s.manager.IsLeader())
}

func (s *LeaderElectionTestSuite) TestManager_buildConfig() {
	const podId = "my_cloudbeat"

	tests := []struct {
		name           string
		want           le.LeaderElectionConfig
		shouldSetEnvar bool
		wantErr        bool
	}{
		{
			name: "Leader election config created as expected",
			want: le.LeaderElectionConfig{
				Lock: &rl.LeaseLock{
					LockConfig: rl.ResourceLockConfig{
						Identity:      fmt.Sprintf("%s_%s", LeaderLeaseName, podId),
						EventRecorder: nil,
					},
				},
			},
			shouldSetEnvar: true,
			wantErr:        false,
		},
		{
			name:           "No POD_NAME env var was set using uuid as an identifier",
			want:           le.LeaderElectionConfig{},
			shouldSetEnvar: false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		if tt.shouldSetEnvar {
			s.Require().NoError(os.Setenv(PodNameEnvar, podId))
		}

		got, err := s.manager.buildConfig(s.T().Context())
		if (err != nil) != tt.wantErr {
			s.FailNow("unexpected error", "error: %v", err)
		}

		if !tt.shouldSetEnvar {
			// verify that the lock identity has been constructed with uuid
			err := parseUUID(got)
			s.Require().NoError(err)
		} else {
			s.Equal(got.Lock.Identity(), tt.want.Lock.Identity(), "buildConfig() got = %v, want %v", got, tt.want)
		}

		s.Require().NoError(os.Unsetenv(PodNameEnvar))
	}
}

func parseUUID(cfg le.LeaderElectionConfig) error {
	id := cfg.Lock.Identity()
	parts := strings.Split(id, "_")
	uuidAsString := parts[len(parts)-1]
	_, err := uuid.ParseUUID(uuidAsString)

	return err
}

func generateLease(holderIdentity *string) *v1.Lease {
	return &v1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      LeaderLeaseName,
			Namespace: core.NamespaceDefault,
		},
		Spec: v1.LeaseSpec{
			HolderIdentity: holderIdentity,
		},
	}
}
