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
	"github.com/elastic/cloudbeat/pipeline"
	"github.com/elastic/cloudbeat/resources/fetching"
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
)

const (
	reconfigureWaitTimeout = 10 * time.Minute
	flushInterval          = 10 * time.Second
	eventsThreshold        = 75
	resourceChBuffer       = 50
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
	resourceCh    chan fetching.ResourceInfo
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

	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)
	fetchersRegistry, err := initRegistry(log, c, resourceCh)
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

	t := transformer.NewTransformer(log, commonData, resultsIndex)

	bt := &cloudbeat{
		ctx:           ctx,
		cancel:        cancel,
		config:        c,
		configUpdates: config.Updates(ctx, log),
		evaluator:     eval,
		data:          data,
		transformer:   t,
		log:           log,
		resourceCh:    resourceCh,
	}
	return bt, nil
}

// Run starts cloudbeat.
func (bt *cloudbeat) Run(b *beat.Beat) error {
	bt.log.Info("cloudbeat is running! Hit CTRL-C to stop it.")

	// Configure the beats Manager to start after all the reloadable hooks are initialized
	// and shutdown when the function returns.
	if err := b.Manager.Start(); err != nil {
		return err
	}
	defer b.Manager.Stop()

	// Wait for Fleet-side reconfiguration only if cloudbeat is running in Agent-managed mode.
	if b.Manager.Enabled() {
		bt.log.Infof("Waiting for initial reconfiguration from Fleet server...")
		update, err := bt.reconfigureWait(reconfigureWaitTimeout)
		if err != nil {
			bt.log.Errorf("Failed while waiting for initial reconfiguraiton from Fleet server: %v", err)
			return err
		}

		if err := bt.configUpdate(update); err != nil {
			return fmt.Errorf("failed to update with initial reconfiguration from Fleet server: %w", err)
		}
	}

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

	// Creating the data pipeline
	findingsCh := pipeline.Step(bt.ctx, bt.resourceCh, bt.evaluator.Eval)
	eventsCh := pipeline.Step(bt.ctx, findingsCh, bt.transformer.CreateBeatEvents)

	var eventsToSend []beat.Event
	for {
		select {
		case <-bt.ctx.Done():
			return nil

		case update := <-bt.configUpdates:
			if err := bt.configUpdate(update); err != nil {
				bt.log.Errorf("Failed to update cloudbeat config: %v", err)
			}

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-time.Tick(flushInterval):
			if len(eventsToSend) == 0 {
				continue
			}

			bt.log.Infof("Flushing %d cloudbeat events to elasticsearch", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil

		// Flush events to ES when reaching a certain threshold
		case events := <-eventsCh:
			eventsToSend = append(eventsToSend, events...)
			if len(eventsToSend) < eventsThreshold {
				continue
			}

			bt.log.Infof("Publish %d cloudbeat events to elasticsearch", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil
		}
	}
}

// reconfigureWait will wait for and consume incoming reconfuration from the Fleet server, and keep
// discarding them until the incoming config contains the necessary information to start cloudbeat
// properly, thereafter returning the valid config.
func (bt *cloudbeat) reconfigureWait(timeout time.Duration) (*common.Config, error) {
	start := time.Now()
	timer := time.After(timeout)

	for {
		select {
		case <-bt.ctx.Done():
			return nil, fmt.Errorf("cancelled via context")

		case <-timer:
			return nil, fmt.Errorf("timed out waiting for reconfiguration after %s", time.Since(start))

		case update, ok := <-bt.configUpdates:
			if !ok {
				return nil, fmt.Errorf("reconfiguration channel is closed")
			}

			c, err := config.New(update)
			if err != nil {
				bt.log.Errorf("Could not parse reconfiguration %v, skipping with error: %v", update.FlattenedKeys(), err)
				continue
			}

			if len(c.Streams) == 0 {
				bt.log.Warnf("No streams received in reconfiguration %v", update.FlattenedKeys())
				continue
			}

			if c.Streams[0].DataYaml == nil {
				bt.log.Warnf("data_yaml not present in reconfiguration %v", update.FlattenedKeys())
				continue
			}

			bt.log.Infof("Received valid reconfiguration after waiting for %s", time.Since(start))
			return update, nil
		}
	}
}

// configUpdate applies incoming reconfiguration from the Fleet server to the cloudbeat config,
// and updates the hosted bundle with the new values.
func (bt *cloudbeat) configUpdate(update *common.Config) error {
	if err := bt.config.Update(bt.log, update); err != nil {
		return err
	}

	policies, err := csppolicies.CISKubernetes()
	if err != nil {
		return fmt.Errorf("could not load CIS Kubernetes policies: %w", err)
	}

	if len(bt.config.Streams) == 0 {
		bt.log.Infof("Did not receive any input stream from incoming config, skipping.")
		return nil
	}

	y, err := bt.config.DataYaml()
	if err != nil {
		return fmt.Errorf("could not marshal to YAML: %w", err)
	}

	if err := csppolicies.HostBundleWithDataYaml("bundle.tar.gz", policies, y); err != nil {
		return fmt.Errorf("could not update bundle with dataYaml: %w", err)
	}

	bt.log.Infof("Bundle updated with dataYaml: %s", y)
	return nil
}

func initRegistry(log *logp.Logger, c config.Config, ch chan fetching.ResourceInfo) (manager.FetchersRegistry, error) {
	registry := manager.NewFetcherRegistry(log)

	if err := manager.Factories.RegisterFetchers(log, registry, c, ch); err != nil {
		return nil, err
	}

	return registry, nil
}

// Stop stops cloudbeat.
func (bt *cloudbeat) Stop() {
	bt.cancel()
	bt.data.Stop()
	bt.evaluator.Stop(bt.ctx)
	close(bt.resourceCh)

	bt.client.Close()
}

// configureProcessors configure processors to be used by the beat
func (bt *cloudbeat) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}
