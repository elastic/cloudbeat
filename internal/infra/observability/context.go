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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type contextKeyType int

const contextKey contextKeyType = iota

func otelFromContext(ctx context.Context) otelProviders {
	if ctx != nil {
		if otl, ok := ctx.Value(contextKey).(otelProviders); ok {
			return otl
		}
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			return otelProviders{
				traceProvider: tracerNoShutdown{span.TracerProvider()},
				meterProvider: meterNoShutdown{otel.GetMeterProvider()},
			}
		}
	}
	return otelProviders{
		traceProvider: tracerNoShutdown{otel.GetTracerProvider()},
		meterProvider: meterNoShutdown{otel.GetMeterProvider()},
	}
}

func contextWithOTel(ctx context.Context, otl otelProviders) context.Context {
	return context.WithValue(ctx, contextKey, otl)
}

func TracerFromContext(ctx context.Context, name string, opts ...trace.TracerOption) trace.Tracer {
	return otelFromContext(ctx).traceProvider.Tracer(name, opts...)
}

func MeterFromContext(ctx context.Context, name string, opts ...metric.MeterOption) metric.Meter {
	return otelFromContext(ctx).meterProvider.Meter(name, opts...)
}

// meterNoShutdown and tracerNoShutdown patch the metric.MeterProvider and trace.TracerProvider interfaces with Shutdown
// and Flush operations.
// The trace.TracerProvider interface does not provide Shutdown or Flush but sdktrace.TracerProvider (returned by
// newTracerProvider) does. otel.GetTracerProvider() returns the first instead of the second. In real life, this will be
// the case when using a no-op tracer (e.g. on-prem with no APM server set up).
type (
	meterNoShutdown struct {
		metric.MeterProvider
	}
	tracerNoShutdown struct {
		trace.TracerProvider
	}
)

func (m meterNoShutdown) ForceFlush(context.Context) error  { return nil }
func (m meterNoShutdown) Shutdown(context.Context) error    { return nil }
func (t tracerNoShutdown) ForceFlush(context.Context) error { return nil }
func (t tracerNoShutdown) Shutdown(context.Context) error   { return nil }
