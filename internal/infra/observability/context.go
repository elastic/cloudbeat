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

func otelFromContext(ctx context.Context) OTel {
	if ctx != nil {
		if otl, ok := ctx.Value(contextKey).(OTel); ok {
			return otl
		}
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			return OTel{
				TracerProvider: span.TracerProvider(),
				MeterProvider:  otel.GetMeterProvider(),
			}
		}
	}
	return OTel{
		TracerProvider: otel.GetTracerProvider(),
		MeterProvider:  otel.GetMeterProvider(),
	}
}

func contextWithOTel(ctx context.Context, otl OTel) context.Context {
	return context.WithValue(ctx, contextKey, otl)
}

func TracerFromContext(ctx context.Context, name string, opts ...trace.TracerOption) trace.Tracer {
	return otelFromContext(ctx).TracerProvider.Tracer(name, opts...)
}

func MeterFromContext(ctx context.Context, name string, opts ...metric.MeterOption) metric.Meter {
	return otelFromContext(ctx).MeterProvider.Meter(name, opts...)
}
