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

package conditions

import (
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetchers"
)

type LeaderLeaseProvider interface {
	IsLeader() (bool, error)
}

type LeaseFetcherCondition struct {
	provider LeaderLeaseProvider
}

func NewLeaseFetcherCondition(provider LeaderLeaseProvider) fetchers.FetcherCondition {
	return &LeaseFetcherCondition{
		provider: provider,
	}
}

func (c *LeaseFetcherCondition) Condition() bool {
	l, err := c.provider.IsLeader()
	if err != nil {
		logp.L().Errorf("could not read leader value, using default value %v: %v", l, err)
	}
	return l
}

func (c *LeaseFetcherCondition) Name() string {
	return "leader_election_conditional_fetcher"
}
