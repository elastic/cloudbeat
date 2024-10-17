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

package uniqueness

import (
	"time"
)

const (
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	LeaseDuration = 30 * time.Second
	// RenewDeadline is the duration that the acting manager will retry
	// refreshing leadership before giving up.
	//
	RenewDeadline = 3 * time.Second

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	RetryPeriod = 2 * time.Second

	// FirstLeaderDeadline is the duration to wait for the leader to acquire the lease for the first time.
	// Known issue: the lease is not released when we delete an agent deployment,
	// it's causing the new agents to think that the old agent still hold the lease,
	// therefore, we wait for at least a LeaseDuration + few seconds.
	FirstLeaderDeadline = LeaseDuration + 5*time.Second

	PodNameEnvar = "POD_NAME"

	LeaderLeaseName = "cloudbeat-cluster-leader"
)
