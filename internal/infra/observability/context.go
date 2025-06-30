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

func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return TracerFromContext(ctx, tracerName).Start(ctx, spanName, opts...)
}

func TracerFromContext(ctx context.Context, name string, opts ...trace.TracerOption) trace.Tracer {
	return otelFromContext(ctx).TracerProvider.Tracer(name, opts...)
}

func MeterFromContext(ctx context.Context, name string, opts ...metric.MeterOption) metric.Meter {
	return otelFromContext(ctx).MeterProvider.Meter(name, opts...)
}
