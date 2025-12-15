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

package observability

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/elastic/cloudbeat/version"
)

const (
	serviceName    = "cloudbeat"
	endpointEnvVar = "OTEL_EXPORTER_OTLP_ENDPOINT"
)

var (
	traceProvider *sdktrace.TracerProvider
	meterProvider *sdkmetric.MeterProvider
	once          sync.Once
)

// SetUpOtel initializes OpenTelemetry logging, tracing, and metrics providers.
// It configures OTLP exporters that send data to an OTLP endpoint
// (e.g., APM Server) configured via environment variables.
func SetUpOtel(ctx context.Context, logger *logp.Logger) error {
	logger = logger.Named("otel")
	if os.Getenv(endpointEnvVar) == "" {
		logger.Infof("%s is not set, skipping OpenTelemetry setup", endpointEnvVar)
		return nil
	}

	var err error
	once.Do(func() {
		err = setUpOnce(ctx, logger)
	})
	return err
}

func setUpOnce(ctx context.Context, logger *logp.Logger) error {
	wrap := loggerWrapper{l: logger}
	otel.SetLogger(logr.New(&wrap))
	otel.SetErrorHandler(&wrap)

	res, err := newResource(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTel resource: %w", err)
	}

	err = newDefaultMeterProvider(ctx, res)
	if err != nil {
		return fmt.Errorf("failed to create metrics provider: %w", err)
	}

	err = newDefaultTracerProvider(ctx, res)
	if err != nil {
		return fmt.Errorf("failed to create tracer provider: %w", err)
	}

	return nil
}

// FailSpan records an error in the span and sets its status to Error.
// It returns an error that includes the original error message.
// Note: If you want to record an error in a span but not mark the span as failed, use `span.RecordError(err)` instead.
func FailSpan(span trace.Span, msg string, err error) error {
	err = fmt.Errorf("%s: %w", msg, err)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

// ShutdownOtel flushes and shuts down the registered OpenTelemetry providers.
func ShutdownOtel(ctx context.Context) error {
	var err error
	if meterProvider != nil {
		err = errors.Join(
			meterProvider.ForceFlush(ctx),
			meterProvider.Shutdown(ctx),
		)
	}
	if traceProvider != nil {
		err = errors.Join(
			err,
			traceProvider.ForceFlush(ctx),
			traceProvider.Shutdown(ctx),
		)
	}
	return err
}

func newDefaultMeterProvider(ctx context.Context, res *resource.Resource) error {
	// The OTLP gRPC exporter will be configured using environment variables (e.g., OTEL_EXPORTER_OTLP_ENDPOINT).
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)
	otel.SetMeterProvider(meterProvider)
	return nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersion(version.CloudbeatSemanticVersion()),
		),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithProcess(),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create application resource: %w", err)
	}

	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		return nil, fmt.Errorf("failed to merge OTel resources: %w", err)
	}
	return res, nil
}

func newDefaultTracerProvider(ctx context.Context, res *resource.Resource) error {
	// The APM server supports OTLP over gRPC, so we use the gRPC exporter.
	// The OTLP gRPC exporter uses environment variables for configuration (e.g., OTEL_EXPORTER_OTLP_ENDPOINT).
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter), // Batches spans for better performance.
		sdktrace.WithSpanProcessor(ensureSpanNameProcessor{}),
	)
	// Set the global TracerProvider to allow instrumentation libraries to use it.
	otel.SetTracerProvider(traceProvider)
	return nil
}

// loggerWrapper is a wrapper around logp.Logger that implements the logr.LogSink and otel.ErrorHandler interfaces.
type loggerWrapper struct {
	l *logp.Logger
}

// Handle handles any error deemed irremediable by an OpenTelemetry component.
// Implements the otel.ErrorHandler interface.
func (w *loggerWrapper) Handle(err error) {
	w.Error(err, "otel error")
}

func (w *loggerWrapper) Init(ri logr.RuntimeInfo) {
	w.l = w.l.WithOptions(zap.AddCallerSkip(ri.CallDepth))
}

func (w *loggerWrapper) Enabled(level int) bool {
	// The higher the level, the more verbose the logging. E.g. warn is 1, Info is 4, Debug is 8.
	// From the OTel documentation:
	// To see Warn messages use a logger with `l.V(1).Enabled() == true`
	// To see Info messages use a logger with `l.V(4).Enabled() == true`
	// To see Debug messages use a logger with `l.V(8).Enabled() == true`.
	return level <= 4
}

func (w *loggerWrapper) Info(level int, msg string, keysAndValues ...any) {
	if !w.Enabled(level) {
		return
	}
	w.l.Infow(msg, keysAndValues...)
}

func (w *loggerWrapper) Error(err error, msg string, keysAndValues ...any) {
	w.l.With(logp.Error(err)).Errorw(msg, keysAndValues...)
}

func (w *loggerWrapper) WithValues(keysAndValues ...any) logr.LogSink {
	return &loggerWrapper{l: w.l.With(keysAndValues...)}
}

func (w *loggerWrapper) WithName(name string) logr.LogSink {
	return &loggerWrapper{l: w.l.Named(name)}
}
