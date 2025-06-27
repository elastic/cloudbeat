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
	"github.com/elastic/cloudbeat/internal/infra/observability"
	"github.com/elastic/cloudbeat/version"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/monitoring"
	"github.com/goforj/godump"
	"github.com/prometheus/client_golang/prometheus"
	"go.elastic.co/apm/module/apmprometheus/v2"
	"go.elastic.co/apm/v2"
	"os"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	_ "github.com/elastic/cloudbeat/internal/processor" // Add cloudbeat default processors.
)

func init() {
	if err := os.Setenv("ELASTIC_APM_ACTIVE", "true"); err != nil {
		panic(fmt.Sprintf("failed to set ELASTIC_APM_ACTIVE environment variable: %v", err))
	}
}

type posture struct {
	flavorBase
	benchmark  builder.Benchmark
	tracerDone func()
	tracer     *apm.Tracer
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

	reg := newMonitoringRegistry(b, "cloudbeat")
	tracer, err := newTracer("cloudbeat", version.CloudbeatSemanticVersion(), log.Logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	log.Infof("posture configured with tracer %s", godump.DumpStr(tracer))
	log.Infof("posture configured with registry %s", godump.DumpStr(reg))

	publisher := NewPublisher(log, flushInterval, eventsThreshold, client, reg, tracer)

	return &posture{
		flavorBase: flavorBase{
			ctx:                ctx,
			cancel:             cancel,
			publisher:          publisher,
			monitoringRegistry: reg,
			config:             cfg,
			log:                log,
			client:             client,
		},
		benchmark: bench,
		tracer:    tracer,
	}, nil
}

func newMonitoringRegistry(b *beat.Beat, name string) *monitoring.Registry {
	ns := b.Info.Monitoring.Namespace
	if ns == nil {
		return nil // No monitoring namespace, no registry
	}
	reg := ns.GetRegistry().GetRegistry(name)
	if reg == nil {
		reg = ns.GetRegistry().NewRegistry(name)
	}
	return reg
}

func newTracer(serviceName, serviceVersion string, logger *logp.Logger) (*apm.Tracer, error) {
	//tr, err := transport.NewHTTPTransport(transport.HTTPTransportOptions{})
	//if err != nil {
	//	return nil, err
	//}
	//
	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:        serviceName,
		ServiceVersion:     serviceVersion,
		ServiceEnvironment: "",
		//Transport:          tr,
	})
	if err != nil {
		return nil, err
	}

	tracer.SetSpanStackTraceMinDuration(-1)              // always include stacktrace
	tracer.SetLogger(warningLogger{logger.Named("apm")}) // Set logger for APM tracer
	apm.SetDefaultTracer(tracer)

	return tracer, nil
}

// warningLogger wraps logp.Logger to allow to be set in the apm.Tracer.
type warningLogger struct {
	logp *logp.Logger
}

// Warningf logs a message at warning level.
func (l warningLogger) Warningf(format string, args ...interface{}) {
	l.logp.Warnf("GREPME: "+format, args...)
}

func (l warningLogger) Errorf(format string, args ...interface{}) {
	l.logp.Errorf("GREPME: "+format, args...)
}

func (l warningLogger) Debugf(format string, args ...interface{}) {
	l.logp.Infof("GREPME: "+format, args...)
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

	//prometheusRegistry := prometheus.NewRegistry()
	//prometheusRegistry.MustRegister(observability.All...)
	//bt.tracerDone = bt.tracer.RegisterMetricsGatherer(apmprometheus.Wrap(prometheusRegistry))
	prometheus.MustRegister(observability.EventsPublished)
	bt.tracerDone = bt.tracer.RegisterMetricsGatherer(apmprometheus.Wrap(prometheus.DefaultGatherer))

	return nil
}

// Stop stops posture.
func (bt *posture) Stop() {
	defer bt.cancel() // context cancellation should be the last action
	bt.benchmark.Stop()

	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	if bt.tracer != nil {
		bt.tracer.Flush(nil)
	}
	if bt.tracerDone != nil {
		bt.tracerDone()
		bt.tracerDone = nil
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
