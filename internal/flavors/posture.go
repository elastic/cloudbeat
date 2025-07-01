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
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
	_ "github.com/elastic/cloudbeat/internal/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/version"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"os"
)

func init() {
	envVars := map[string]string{
		"ELASTIC_APM_ACTIVE":          "true",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://apm-server.elastic-agent:8200",
		"OTEL_LOGS_EXPORTER":          "otlp",
		"OTEL_METRICS_EXPORTER":       "otlp",
		"OTEL_TRACES_EXPORTER":        "otlp",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			panic(fmt.Sprintf("failed to set %s environment variable: %v", key, err))
		}
	}
}

type posture struct {
	flavorBase
	benchmark builder.Benchmark
}

// NewPosture creates an instance of posture.
func NewPosture(b *beat.Beat, agentConfig *agentconfig.C) (beat.Beater, error) {
	cfg, err := config.New(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return newPostureFromCfg(b, cfg)
}

// NewPosture creates an instance of posture.
func newPostureFromCfg(b *beat.Beat, cfg *config.Config) (*posture, error) {
	log := clog.NewLogger("posture")
	log.Info("Config initiated with cycle period of ", cfg.Period)
	ctx, cancel := context.WithCancel(context.Background())

	strategy, err := benchmark.GetStrategy(cfg, log)
	if err != nil {
		cancel()
		return nil, err
	}

	log.Infof("Creating benchmark %T", strategy)
	bench, err := strategy.NewBenchmark(ctx, log, cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	err = ensureHostProcessor(log, cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	client, err := NewClient(b.Publisher, cfg.Processors)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to init client: %w", err)
	}
	log.Infof("posture configured %d processors", len(cfg.Processors))

	ctx, err = observability.SetUpOtel(ctx, "orestis-cloudbeat", version.CloudbeatSemanticVersion(), log.Logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to set up OpenTelemetry: %w", err)
	}
	// TODO: these need shutdown...

	publisher := NewPublisher(log, flushInterval, eventsThreshold, client)

	return &posture{
		flavorBase: flavorBase{
			ctx:       ctx,
			cancel:    cancel,
			publisher: publisher,
			config:    cfg,
			log:       log,
			client:    client,
		},
		benchmark: bench,
	}, nil
}

// Run starts posture.
func (bt *posture) Run(*beat.Beat) error {
	bt.log.Info("posture is running! Hit CTRL-C to stop it")

	eventsCh, err := bt.benchmark.Run(bt.ctx)
	if err != nil {
		return err
	}

	bt.publisher.HandleEvents(bt.ctx, eventsCh)
	bt.log.Warn("Posture has finished running")

	return nil
}

// Stop stops posture.
func (bt *posture) Stop() {
	defer bt.cancel() // context cancellation should be the last action
	bt.benchmark.Stop()

	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

}

// ensureAdditionalProcessors modifies cfg.Processors list to ensure 'host'
// processor is present for K8s and EKS benchmarks.
func ensureHostProcessor(log *clog.Logger, cfg *config.Config) error {
	if cfg.Benchmark != config.CIS_EKS && cfg.Benchmark != config.CIS_K8S {
		return nil
	}
	log.Info("Adding host processor config")
	hostProcessor, err := agentconfig.NewConfigFrom("add_host_metadata: ~")
	if err != nil {
		return err
	}
	cfg.Processors = append(cfg.Processors, hostProcessor)
	return nil
}
