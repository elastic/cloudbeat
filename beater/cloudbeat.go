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
	"context"
	"fmt"

	"github.com/elastic/cloudbeat/evaluator"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/cloudbeat/config"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/elastic/cloudbeat/transformer"

	"github.com/gofrs/uuid"
)

// cloudbeat configuration.
type cloudbeat struct {
	ctx    context.Context
	cancel context.CancelFunc

	config      config.Config
	client      beat.Client
	data        *manager.Data
	evaluator   evaluator.Evaluator
	transformer transformer.Transformer
	log         *logp.Logger
}

// New creates an instance of cloudbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	log := logp.NewLogger("cloudbeat")

	ctx, cancel := context.WithCancel(context.Background())

	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	log.Info("Config initiated.")

	fetchersRegistry, err := InitRegistry(c)
	if err != nil {
		cancel()
		return nil, err
	}

	data, err := manager.NewData(c.Period, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	// namespace will be passed as param from fleet on https://github.com/elastic/security-team/issues/2383 and it's user configurable
	resultsIndex := config.Datastream("", config.ResultsDatastreamIndexPrefix)
	if err != nil {
		cancel()
		return nil, err
	}

	cdp, err := transformer.NewCommonDataProvider(c)
	if err != nil {
		cancel()
		return nil, err
	}

	t, err := transformer.NewTransformer(ctx, eval, cdp, resultsIndex)
	if err != nil {
		cancel()
		return nil, err
	}

	bt := &cloudbeat{
		ctx:         ctx,
		cancel:      cancel,
		config:      c,
		evaluator:   eval,
		data:        data,
		transformer: t,
		log:         log,
	}
	return bt, nil
}

// Run starts cloudbeat.
func (bt *cloudbeat) Run(b *beat.Beat) error {
	bt.log.Info("cloudbeat is running! Hit CTRL-C to stop it.")

	// Configure the beats Manager to start after all the reloadable hooks are initialized
	// and shutdown when the function return.
	if err := b.Manager.Start(); err != nil {
		return err
	}
	defer b.Manager.Stop()

	if err := bt.data.Run(bt.ctx); err != nil {
		return err
	}

	procs, err := bt.configureProcessors(bt.config.Processors)
	if err != nil {
		return err
	}

	// Connect publisher (with beat's processors)
	if bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			Processor: procs,
		},
	}); err != nil {
		return err
	}

	output := bt.data.Output()

	for {
		select {
		case <-bt.ctx.Done():
			return nil
		case fetchedResources := <-output:
			cycleId, _ := uuid.NewV4()
			bt.log.Debugf("Cycle % has started", cycleId)
			cycleMetadata := transformer.CycleMetadata{CycleId: cycleId}
			// TODO: send events through a channel and publish them by a configured threshold & time
			events := bt.transformer.ProcessAggregatedResources(fetchedResources, cycleMetadata)
			bt.client.PublishAll(events)
			bt.log.Debugf("Cycle % has ended", cycleId)
		}
	}
}

func InitRegistry(c config.Config) (manager.FetchersRegistry, error) {
	registry := manager.NewFetcherRegistry()
	err := manager.Factories.RegisterFetchers(registry, c)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// Stop stops cloudbeat.
func (bt *cloudbeat) Stop() {
	bt.data.Stop(bt.ctx, bt.cancel)
	bt.evaluator.Stop(bt.ctx)

	bt.client.Close()
}

// configureProcessors configure processors to be used by the beat
func (bt *cloudbeat) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}
