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

package leaderelection

import (
	"context"
	"fmt"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/suite"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	le "k8s.io/client-go/tools/leaderelection"
	rl "k8s.io/client-go/tools/leaderelection/resourcelock"
	"os"
	"strings"
	"testing"
)

type LeaderElectionTestSuite struct {
	suite.Suite
	log *logp.Logger
}

func TestLeaderElectionTestSuite(t *testing.T) {
	s := new(LeaderElectionTestSuite)
	s.log = logp.NewLogger("cloudbeat_leader_election_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
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
			os.Setenv(PodNameEnvar, podId)
		}

		got, err := buildConfig(context.TODO(), s.log, k8sfake.NewSimpleClientset(), make(chan bool), nil)
		if (err != nil) != tt.wantErr {
			s.FailNow("unexpected error: %v", err)
		}

		if !tt.shouldSetEnvar {
			// verify that the lock identity has been constructed with uuid
			err := parseUUID(got)
			s.NoError(err)
		} else {
			s.Equal(got.Lock.Identity(), tt.want.Lock.Identity(), "buildConfig() got = %v, want %v", got, tt.want)
		}

		os.Unsetenv(PodNameEnvar)
	}
}

func parseUUID(cfg le.LeaderElectionConfig) error {
	id := cfg.Lock.Identity()
	parts := strings.Split(id, "_")
	uuidAsString := parts[len(parts)-1]
	_, err := uuid.ParseUUID(uuidAsString)

	return err
}
