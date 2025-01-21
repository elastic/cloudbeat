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

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/uniqueness"
)

type k8sbenchmark struct {
	basebenchmark
	leaderElector uniqueness.Manager
}

func (b *Builder) BuildK8s(ctx context.Context, log *clog.Logger, cfg *config.Config, resourceCh chan fetching.ResourceInfo, reg registry.Registry, k8sLeaderElector uniqueness.Manager) (Benchmark, error) {
	base, err := b.buildBase(ctx, log, cfg, resourceCh, reg)
	if err != nil {
		return nil, err
	}

	return &k8sbenchmark{
		basebenchmark: *base,
		leaderElector: k8sLeaderElector,
	}, nil
}

func (b *k8sbenchmark) Run(ctx context.Context) (<-chan []beat.Event, error) {
	err := b.leaderElector.Run(ctx)
	if err != nil {
		return nil, err
	}

	return b.basebenchmark.Run(ctx)
}

func (b *k8sbenchmark) Stop() {
	b.leaderElector.Stop()
	b.basebenchmark.Stop()
}
