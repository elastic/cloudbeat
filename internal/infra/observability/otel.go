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
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/elastic/cloudbeat/version"
)

const serviceName = "cloudbeat"

type OTel struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

func SetUpOtel(ctx context.Context, logger *logp.Logger) (context.Context, error) {
	otel.SetLogger(logr.New(logWrapper{logger.Named("otel")}))

	res, err := newResource(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to create resource: %w", err)
	}

	mp, err := newMetricsProvider(ctx, res)
	if err != nil {
		return ctx, fmt.Errorf("failed to create metrics provider: %w", err)
	}

	tp, err := newTracerProvider(ctx, res)
	if err != nil {
		return ctx, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	return contextWithOTel(ctx, OTel{
		TracerProvider: tp,
		MeterProvider:  mp,
	}), nil
}

func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return TracerFromContext(ctx, tracerName).Start(ctx, spanName, opts...)
}

func newMetricsProvider(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			metricExporter,
		)),
	)

	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.New(
		ctx,
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
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		return nil, fmt.Errorf("failed to merge resource: %w", err)
	}
	return res, nil
}

func newTracerProvider(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// This uses environment variables like OTEL_EXPORTER_OTLP_ENDPOINT
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create a new TracerProvider with the exporter and resource.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	// Set the global TracerProvider.
	otel.SetTracerProvider(tp)
	return tp, nil
}

// TODO: figure out

type logWrapper struct {
	logp *logp.Logger
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
