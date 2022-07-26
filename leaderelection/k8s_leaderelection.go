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
	le "k8s.io/client-go/tools/leaderelection"
)

type K8SLeaderElectionService interface {
	IsLeader() bool
	Run(ctx context.Context)
}

type K8sLeaderElector struct {
	leader *le.LeaderElector
}

func NewK8sLeaderElector(lec le.LeaderElectionConfig) (K8SLeaderElectionService, error) {
	leader, err := le.NewLeaderElector(lec)
	if err != nil {
		return nil, err
	}

	return &K8sLeaderElector{
		leader: leader,
	}, nil
}

func (l K8sLeaderElector) IsLeader() bool {
	return l.leader.IsLeader()
}

func (l K8sLeaderElector) Run(ctx context.Context) {
	l.leader.Run(ctx)
}
