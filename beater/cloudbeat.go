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
	"time"

	"github.com/elastic/cloudbeat/leaderelection"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/launcher"
	"github.com/elastic/cloudbeat/pipeline"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/transformer"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/beats/v7/libbeat/processors"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	flushInterval    = 10 * time.Second
	eventsThreshold  = 75
	resourceChBuffer = 10000
)

// cloudbeat configuration.
type cloudbeat struct {
	ctx         context.Context
	cancel      context.CancelFunc
	config      config.Config
	client      beat.Client
	data        *fetchersManager.Data
	evaluator   evaluator.Evaluator
	transformer transformer.Transformer
	log         *logp.Logger
	resourceCh  chan fetching.ResourceInfo
	leader      leaderelection.ElectionManager
}

func New(b *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	log := logp.NewLogger("launcher")
	ctx := context.Background()
	reloader := launcher.NewListener(ctx, log)
	validator := &validator{}

	s, err := launcher.New(ctx, log, reloader, validator, NewCloudbeat, cfg)
	if err != nil {
		return nil, err
	}

	reload.Register.MustRegisterList("inputs", reloader)
	return s, nil
}

func NewCloudbeat(b *beat.Beat, cfg *agentconfig.C) (beat.Beater, error) {
	return newCloudbeat(b, cfg)
}

// NewCloudbeat creates an instance of cloudbeat.
func newCloudbeat(_ *beat.Beat, cfg *agentconfig.C) (*cloudbeat, error) {
	log := logp.NewLogger("cloudbeat")

	ctx, cancel := context.WithCancel(context.Background())

	c, err := config.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	log.Info("Config initiated with cycle period of ", c.Period)

	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)

	le, err := leaderelection.NewLeaderElector(log, c)
	if err != nil {
		cancel()
		return nil, err
	}

	fetchersRegistry, err := initRegistry(log, c, resourceCh, le)
	if err != nil {
		cancel()
		return nil, err
	}

	data, err := fetchersManager.NewData(log, c.Period, time.Minute, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log, c)
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

	cdp, err := transformer.NewCommonDataProvider(log, c)
	if err != nil {
		cancel()
		return nil, err
	}

	commonData, err := cdp.FetchCommonData(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	t := transformer.NewTransformer(log, commonData, resultsIndex)

	bt := &cloudbeat{
		ctx:         ctx,
		cancel:      cancel,
		config:      c,
		evaluator:   eval,
		data:        data,
		transformer: t,
		log:         log,
		resourceCh:  resourceCh,
		leader:      le,
	}
	return bt, nil
}

// Run starts cloudbeat.
func (bt *cloudbeat) Run(b *beat.Beat) error {
	bt.log.Info("cloudbeat is running! Hit CTRL-C to stop it.")
	if err := bt.leader.Run(bt.ctx); err != nil {
		return err
	}

	if err := bt.data.Run(bt.ctx); err != nil {
		return err
	}

	procs, err := bt.configureProcessors(bt.config.Processors)
	if err != nil {
		return err
	}
	bt.log.Debugf("cloudbeat configured %d processors", len(bt.config.Processors))

	// Connect publisher (with beat's processors)
	if bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			Processor: procs,
		},
	}); err != nil {
		return err
	}

	// Creating the data pipeline
	findingsCh := pipeline.Step(bt.log, bt.resourceCh, bt.evaluator.Eval)
	eventsCh := pipeline.Step(bt.log, findingsCh, bt.transformer.CreateBeatEvents)

	var eventsToSend []beat.Event
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-bt.ctx.Done():
			bt.log.Warn("cloudbeat context is done")
			return nil

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-ticker.C:
			if len(eventsToSend) == 0 {
				continue
			}

			bt.log.Infof("Publishing %d cloudbeat events to elasticsearch, time interval reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil

		// Flush events to ES when reaching a certain threshold
		case events := <-eventsCh:
			eventsToSend = append(eventsToSend, events...)
			if len(eventsToSend) < eventsThreshold {
				continue
			}

			bt.log.Infof("Publishing %d cloudbeat events to elasticsearch, buffer threshold reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil
		}
	}
}

func initRegistry(log *logp.Logger, cfg config.Config, ch chan fetching.ResourceInfo, le leaderelection.ElectionManager) (fetchersManager.FetchersRegistry, error) {
	registry := fetchersManager.NewFetcherRegistry(log)

	parsedList, err := fetchersManager.Factories.ParseConfigFetchers(log, cfg, ch)
	if err != nil {
		return nil, err
	}

	if err := registry.RegisterFetchers(parsedList, le); err != nil {
		return nil, err
	}

	return registry, nil
}

// Stop stops cloudbeat.
func (bt *cloudbeat) Stop() {
	bt.data.Stop()
	bt.evaluator.Stop(bt.ctx)
	bt.leader.Stop()
	close(bt.resourceCh)
	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	bt.cancel()
}

// configureProcessors configure processors to be used by the beat
func (bt *cloudbeat) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}
