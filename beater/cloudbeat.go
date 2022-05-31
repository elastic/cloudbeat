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

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/elastic/cloudbeat/transformer"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	csppolicies "github.com/elastic/csp-security-policies/bundle"

	"github.com/gofrs/uuid"
)

// cloudbeat configuration.
type cloudbeat struct {
	ctx    context.Context
	cancel context.CancelFunc

	config        config.Config
	configUpdates <-chan *common.Config
	client        beat.Client
	data          *manager.Data
	evaluator     evaluator.Evaluator
	transformer   transformer.Transformer
	log           *logp.Logger
}

// New creates an instance of cloudbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	log := logp.NewLogger("cloudbeat")

	ctx, cancel := context.WithCancel(context.Background())

	c, err := config.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	log.Info("Config initiated.")

	fetchersRegistry, err := InitRegistry(log, c)
	if err != nil {
		cancel()
		return nil, err
	}

	data, err := manager.NewData(log, c.Period, time.Minute, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log)
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

	t := transformer.NewTransformer(ctx, log, eval, commonData, resultsIndex)

	bt := &cloudbeat{
		ctx:           ctx,
		cancel:        cancel,
		config:        c,
		configUpdates: config.Updates(ctx, log),
		evaluator:     eval,
		data:          data,
		transformer:   t,
		log:           log,
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

		case update := <-bt.configUpdates:
			if err := bt.config.Update(update); err != nil {
				bt.log.Errorf("Could not update cloudbeat config: %v", err)
				break
			}

			policies, err := csppolicies.CISKubernetes()
			if err != nil {
				bt.log.Errorf("Could not load CIS Kubernetes policies: %v", err)
				break
			}

			if len(bt.config.Streams) == 0 {
				bt.log.Infof("Did not receive any input stream, skipping.")
				break
			}

			y, err := bt.config.DataYaml()
			if err != nil {
				bt.log.Errorf("Could not marshal to YAML: %v", err)
				break
			}

			if err := csppolicies.HostBundleWithDataYaml("bundle.tar.gz", policies, y); err != nil {
				bt.log.Errorf("Could not update bundle with dataYaml: %v", err)
				break
			}

			bt.log.Infof("Bundle updated with dataYaml: %s", y)

		case fetchedResources := <-output:
			cycleId, _ := uuid.NewV4()
			bt.log.Infof("Cycle %v has started", cycleId)

			cycleMetadata := transformer.CycleMetadata{CycleId: cycleId}
			// TODO: send events through a channel and publish them by a configured threshold & time
			events := bt.transformer.ProcessAggregatedResources(fetchedResources, cycleMetadata)

			bt.log.Infof("Publishing %d events to index", len(events))
			bt.client.PublishAll(events)

			bt.log.Infof("Cycle %v has ended", cycleId)
		}
	}
}

func InitRegistry(log *logp.Logger, c config.Config) (manager.FetchersRegistry, error) {
	registry := manager.NewFetcherRegistry(log)

	if err := manager.Factories.RegisterFetchers(log, registry, c); err != nil {
		return nil, err
	}

	return registry, nil
}

// Stop stops cloudbeat.
func (bt *cloudbeat) Stop() {
	bt.cancel()
	bt.data.Stop()
	bt.evaluator.Stop(bt.ctx)

	bt.client.Close()
}

// configureProcessors configure processors to be used by the beat
func (bt *cloudbeat) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}
