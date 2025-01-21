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

	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/pipeline"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type Manager interface {
	Run()
	Stop()
}

type Evaluator interface {
	Eval(ctx context.Context, resource fetching.ResourceInfo) (evaluator.EventData, error)
}

type Transformer interface {
	CreateBeatEvents(ctx context.Context, data evaluator.EventData) ([]beat.Event, error)
}

type basebenchmark struct {
	log         *clog.Logger
	manager     Manager
	evaluator   Evaluator
	transformer Transformer
	resourceCh  chan fetching.ResourceInfo
}

func (b *basebenchmark) Run(ctx context.Context) (<-chan []beat.Event, error) {
	b.manager.Run()
	findingsCh := pipeline.Step(ctx, b.log, b.resourceCh, b.evaluator.Eval)
	eventsCh := pipeline.Step(ctx, b.log, findingsCh, b.transformer.CreateBeatEvents)
	return eventsCh, nil
}

func (b *basebenchmark) Stop() {
	b.manager.Stop()
	close(b.resourceCh)
}
