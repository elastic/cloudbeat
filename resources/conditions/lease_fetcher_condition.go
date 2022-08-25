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
	"github.com/elastic/cloudbeat/leaderelection"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
)

type LeaseFetcherCondition struct {
	log      *logp.Logger
	provider leaderelection.ElectionManager
}

func NewLeaseFetcherCondition(log *logp.Logger, le leaderelection.ElectionManager) fetching.Condition {
	return &LeaseFetcherCondition{
		log:      log,
		provider: le,
	}
}

func (c *LeaseFetcherCondition) Condition() bool {
	return c.provider.IsLeader()
}

func (c *LeaseFetcherCondition) Name() string {
	return "leader_election_conditional_fetcher"
}
