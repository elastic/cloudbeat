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

package beater

import (
	"fmt"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/flavors"
	"github.com/elastic/cloudbeat/internal/launcher"
)

func New(_ *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	log := logp.NewLogger("launcher")
	reloader := launcher.NewListener(log)
	validator := &validator{}

	s := launcher.New(log, "Cloudbeat", reloader, validator, NewBeater, cfg)

	reload.RegisterV2.MustRegisterInput(reloader)
	return s, nil
}

// NewBeater creates an instance of beater.
func NewBeater(b *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	c, err := config.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("NewBeater: could not parse configuration %v, skipping with error: %w", cfg.FlattenedKeys(), err)
	}

	switch c.Type {
	case config.VulnerabilityType:
		return flavors.NewVulnerability(b, cfg)
	case config.AssetInventoryType:
		return flavors.NewAssetInventory(b, cfg)
	default:
		return flavors.NewPosture(b, cfg)
	}
}
