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
	"io"
	"os"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/go-logr/logr"
	"github.com/goforj/godump"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	_ "github.com/elastic/cloudbeat/internal/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/version"
)

func init() {
	if err := os.Setenv("ELASTIC_APM_ACTIVE", "true"); err != nil {
		panic(fmt.Sprintf("failed to set ELASTIC_APM_ACTIVE environment variable: %v", err))
	}
}

type posture struct {
	flavorBase
	benchmark builder.Benchmark
	tracer    *sdktrace.TracerProvider
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

	tracerProvider, err := newTracerProvider(ctx, "cloudbeat", version.CloudbeatSemanticVersion(), log.Logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create tracerProvider: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String("cloudbeat")))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	meterProvider, err := initMeterProvider(res)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create meter provider: %w", err)
	}

	log.Infof("GREPME posture configured with tracerProvider %s", godump.DumpStr(tracerProvider))
	log.Infof("GREPME posture configured with metricsProvider %s", godump.DumpStr(meterProvider))

	publisher := NewPublisher(
		log,
		flushInterval,
		eventsThreshold,
		client,
		tracerProvider.Tracer("orestis"),
		meterProvider.Meter("orestis"),
	)

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
		tracer:    tracerProvider,
	}, nil
}

// initMeterProvider creates and registers a basic OpenTelemetry MeterProvider that prints to stdout.
func initMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(
			metricExporter,
			metric.WithInterval(3*time.Second),
		)),
	)

	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func newTracerProvider(ctx context.Context, name string, version string, log *logp.Logger) (*sdktrace.TracerProvider, error) {
	log.Info("GREPME: Initializing OpenTelemetry TracerProvider")

	// Create a new stdout exporter.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(newTmpWritter(log)))
	if err != nil {
		return nil, err
	}

	// Create a new resource with the service name.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create a new TracerProvider with the exporter and resource.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global TracerProvider.
	otel.SetTracerProvider(tp)
	otel.SetLogger(logr.New(logWrapper{log.Named("GREPME")}))
	return tp, nil
}

func newTmpWritter(log *logp.Logger) io.Writer {
	return logWrapper{log.Named("GREPME")}
}

type logWrapper struct {
	logp *logp.Logger
}

func (l logWrapper) Write(p []byte) (n int, err error) {
	//TODO delete this
	l.logp.Info("GREPME: Write: " + string(p))
	return len(p), nil
}

func (l logWrapper) Init(info logr.RuntimeInfo) {
	l.logp.Info("GREPME: Initializing logr wrapper", "info", info)
}

func (l logWrapper) Enabled(int) bool {
	return true
}

func (l logWrapper) Info(_ int, msg string, keysAndValues ...any) {
	l.logp.Info("GREPME: "+msg, "keysAndValues", keysAndValues)
}

func (l logWrapper) Error(err error, msg string, keysAndValues ...any) {
	l.logp.Error("GREPME: "+msg, "err", err, "keysAndValues", keysAndValues)
}

func (l logWrapper) WithValues(keysAndValues ...any) logr.LogSink {
	return logWrapper{l.logp.With(keysAndValues)}
}

func (l logWrapper) WithName(name string) logr.LogSink {
	return logWrapper{l.logp.With(name)}
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

	//prometheus.MustRegister(observability.EventsPublished)

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
		_ = bt.tracer.ForceFlush(bt.ctx)
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
