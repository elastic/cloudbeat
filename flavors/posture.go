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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/fetchersManager/factory"
	"github.com/elastic/cloudbeat/resources/fetchersManager/manager"
	"github.com/elastic/cloudbeat/resources/fetchersManager/registry"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	k8s "k8s.io/client-go/kubernetes"
	"time"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/pipeline"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/transformer"
	"github.com/elastic/cloudbeat/uniqueness"

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

// posture configuration.
type posture struct {
	flavorBase
	data       *manager.Manager
	evaluator  evaluator.Evaluator
	resourceCh chan fetching.ResourceInfo
	leader     uniqueness.Manager
}

// NewPosture creates an instance of posture.
func NewPosture(_ *beat.Beat, cfg *agentconfig.C) (*posture, error) {
	log := logp.NewLogger("posture")

	ctx, cancel := context.WithCancel(context.Background())

	c, err := config.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	log.Info("Config initiated with cycle period of ", c.Period)

	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)

	kubeClient, err := providers.GetK8sClient(log, c.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		log.Errorf("failed to create kubernetes client: %v", err)
	}
	le := uniqueness.NewLeaderElector(log, kubeClient)

	awsConfigProvider := awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}}

	fetchersRegistry, err := initRegistry(ctx, log, c, resourceCh, le, kubeClient, awslib.GetIdentityClient, awsConfigProvider)
	if err != nil {
		cancel()
		return nil, err
	}

	// TODO: timeout should be configurable and not hard-coded. Setting to 10 minutes for now to account for CSPM fetchers
	// 	https://github.com/elastic/cloudbeat/issues/653
	newData, err := manager.NewManager(ctx, log, c.Period, time.Minute*10, fetchersRegistry)
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

	cdp, err := GetCommonDataProvider(ctx, log, *c)
	if err != nil {
		cancel()
		return nil, err
	}

	t := transformer.NewTransformer(log, cdp, resultsIndex)

	base := flavorBase{
		ctx:         ctx,
		cancel:      cancel,
		config:      c,
		transformer: t,
		log:         log,
	}

	bt := &posture{
		flavorBase: base,
		evaluator:  eval,
		data:       newData,
		resourceCh: resourceCh,
		leader:     le,
	}
	return bt, nil
}

// Run starts posture.
func (bt *posture) Run(b *beat.Beat) error {
	bt.log.Info("posture is running! Hit CTRL-C to stop it")

	if err := bt.leader.Run(bt.ctx); err != nil {
		return err
	}

	bt.data.Run()

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

func initRegistry(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager, k8sClient k8s.Interface, identityProvider func(cfg awssdk.Config) awslib.IdentityProviderGetter, awsConfigProvider awslib.ConfigProviderAPI) (registry.FetchersRegistry, error) {
	f, err := factory.NewFactory(ctx, log, cfg, ch, le, k8sClient, identityProvider, awsConfigProvider)
	if err != nil {
		return nil, err
	}

	return registry.NewFetcherRegistry(log, f), nil
}

// Stop stops posture.
func (bt *posture) Stop() {
	bt.data.Stop()
	bt.evaluator.Stop(bt.ctx)
	bt.leader.Stop()
	close(bt.resourceCh)
	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	bt.cancel()
}
