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

package beater

import (
	"fmt"

	"github.com/elastic/cloudbeat/config"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
)

type validator struct{}

func (v *validator) Validate(cfg *agentconfig.C) error {
	c, err := config.New(cfg)
	if err != nil {
		return fmt.Errorf("could not parse reconfiguration %v, skipping with error: %w", cfg.FlattenedKeys(), err)
	}

	if c.RuntimeCfg == nil {
		return fmt.Errorf("runtime configuration didn't exist in new configuration")
	}

	return nil
}
