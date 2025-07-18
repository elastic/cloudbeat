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

package builder

import (
	"context"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/common"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/rule_ecs"
	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/manager"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/transformer"
	"github.com/elastic/cloudbeat/version"
)

const (
	defaultManagerTimeout = 10 * time.Minute
)

type Benchmark interface {
	Run(ctx context.Context) (<-chan []beat.Event, error)
	Stop()
}

type Builder struct {
	managerTimeout time.Duration
	idp            dataprovider.IdProvider
	bdp            dataprovider.CommonDataProvider
}

func New(options ...Option) *Builder {
	b := &Builder{
		managerTimeout: defaultManagerTimeout,
		idp:            &idProvider{},
		bdp:            &dataProvider{},
	}
	for _, fn := range options {
		fn(b)
	}
	return b
}

func (b *Builder) Build(ctx context.Context, log *clog.Logger, cfg *config.Config, resourceCh chan fetching.ResourceInfo, reg registry.Registry) (Benchmark, error) {
	return b.buildBase(ctx, log, cfg, resourceCh, reg)
}

func (b *Builder) buildBase(ctx context.Context, log *clog.Logger, cfg *config.Config, resourceCh chan fetching.ResourceInfo, reg registry.Registry) (*basebenchmark, error) {
	manager, err := manager.NewManager(ctx, log, cfg.Period, b.managerTimeout, reg)
	if err != nil {
		return nil, err
	}

	evaluator, err := evaluator.NewOpaEvaluator(ctx, log, cfg)
	if err != nil {
		return nil, err
	}

	cdp, err := common.New(version.CloudbeatVersionInfo{
		Version: version.CloudbeatVersion(),
		// Keeping Policy field for backward compatibility
		Policy: version.CloudbeatVersion(),
	}, cfg)
	if err != nil {
		return nil, err
	}

	rep := rule_ecs.NewDataProvider()

	transformer := transformer.NewTransformer(log, cfg, b.bdp, cdp, b.idp, rep)
	return &basebenchmark{
		log:         log,
		manager:     manager,
		evaluator:   evaluator,
		transformer: transformer,
		resourceCh:  resourceCh,
	}, nil
}
