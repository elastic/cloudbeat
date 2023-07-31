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

package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/flavors/benchmark"
	"github.com/elastic/cloudbeat/pipeline"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/manager"
	"github.com/elastic/cloudbeat/transformer"
)

// posture configuration.
type posture struct {
	flavorBase
	fetcherManager *manager.Manager
	evaluator      evaluator.Evaluator
	resourceCh     chan fetching.ResourceInfo
	benchmark      benchmark.Benchmark
}

// NewPosture creates an instance of posture.
func NewPosture(_ *beat.Beat, agentConfig *agentconfig.C) (beat.Beater, error) {
	cfg, err := config.New(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return newPostureFromCfg(cfg)
}

// NewPosture creates an instance of posture.
func newPostureFromCfg(cfg *config.Config) (*posture, error) {
	log := logp.NewLogger("posture")
	log.Info("Config initiated with cycle period of ", cfg.Period)
	ctx, cancel := context.WithCancel(context.Background())

	b, err := benchmark.NewBenchmark(cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)

	fetchersRegistry, cdp, err := b.Initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		cancel()
		return nil, err
	}

	// TODO: timeout should be configurable and not hard-coded. Setting to 10 minutes for now to account for CSPM fetchers
	// 	https://github.com/elastic/cloudbeat/issues/653
	fetcherManager, err := manager.NewManager(ctx, log, cfg.Period, time.Minute*10, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log, cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	// namespace will be passed as param from fleet on https://github.com/elastic/security-team/issues/2383 and it's user configurable
	resultsIndex := config.Datastream("", config.ResultsDatastreamIndexPrefix)

	t := transformer.NewTransformer(log, cdp, resultsIndex)

	return &posture{
		flavorBase: flavorBase{
			ctx:         ctx,
			cancel:      cancel,
			config:      cfg,
			transformer: t,
			log:         log,
		},
		fetcherManager: fetcherManager,
		evaluator:      eval,
		resourceCh:     resourceCh,
		benchmark:      b,
	}, nil
}

// Run starts posture.
func (bt *posture) Run(b *beat.Beat) error {
	bt.log.Info("posture is running! Hit CTRL-C to stop it")

	if err := bt.benchmark.Run(bt.ctx); err != nil {
		return err
	}

	bt.fetcherManager.Run()

	procs, err := ConfigureProcessors(bt.config.Processors)
	if err != nil {
		return err
	}
	bt.log.Debugf("posture configured %d processors", len(bt.config.Processors))

	// Connect publisher (with beat's processors)
	if bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			Processor: procs,
		},
	}); err != nil {
		return err
	}

	// Creating the data pipeline
	findingsCh := pipeline.Step(bt.ctx, bt.log, bt.resourceCh, bt.evaluator.Eval)
	eventsCh := pipeline.Step(bt.ctx, bt.log, findingsCh, bt.transformer.CreateBeatEvents)

	var eventsToSend []beat.Event
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-bt.ctx.Done():
			bt.log.Warn("Posture context is done")
			return nil

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-ticker.C:
			if len(eventsToSend) == 0 {
				continue
			}

			bt.log.Infof("Publishing %d posture events to elasticsearch, time interval reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil

		// Flush events to ES when reaching a certain threshold
		case events := <-eventsCh:
			eventsToSend = append(eventsToSend, events...)
			if len(eventsToSend) < eventsThreshold {
				continue
			}

			bt.log.Infof("Publishing %d posture events to elasticsearch, buffer threshold reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil
		}
	}
}

// Stop stops posture.
func (bt *posture) Stop() {
	bt.fetcherManager.Stop()
	bt.evaluator.Stop(bt.ctx)
	close(bt.resourceCh)
	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	bt.cancel()
}
