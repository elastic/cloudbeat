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

package observability_test

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	metricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	otlpmetricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	otlptracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/elastic/cloudbeat/internal/infra/observability"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var tracer = otel.Tracer("test-scope")
var meter = otel.Meter("test-scope")

func TestOtel(t *testing.T) {
	testhelper.SkipLong(t)
	ctx := t.Context()

	t.Run("No OTel Setup", func(t *testing.T) {
		spanCtx, span := tracer.Start(ctx, "test-span")
		require.NotNil(t, spanCtx)
		assert.False(t, span.IsRecording())
	})

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:12345") // enable OTel exporter
	endpoint := "localhost:12345"
	server, traceService, metricService := startMockOtlpServer(t, endpoint)
	t.Cleanup(server.Stop)

	log, observedLogs := testhelper.NewObserverLogger(t)
	err := observability.SetUpOtel(ctx, log.Logger)
	require.NoError(t, err)

	var spanID, traceID string
	t.Run("Start Span", func(t *testing.T) {
		spanCtx, span := tracer.Start(ctx, "test-span")
		require.NotNil(t, spanCtx)
		assert.True(t, span.IsRecording())
		assert.True(t, span.SpanContext().IsValid())

		spanID = span.SpanContext().SpanID().String()
		traceID = span.SpanContext().TraceID().String()

		log.WithSpanContext(span.SpanContext()).Info("test span logging")

		span.End() // End the span to ensure it is recorded
	})

	counter, err := meter.Int64Counter("test-counter")
	require.NoError(t, err)
	counter.Add(ctx, 10)

	t.Run("Check observed logs", func(t *testing.T) {
		logs := observedLogs.TakeAll()
		require.NotEmpty(t, logs)
		seen := false
		for _, entry := range logs {
			assert.GreaterOrEqual(t, entry.Level, zap.InfoLevel)
			if strings.HasSuffix(entry.LoggerName, ".otel") {
				// Logs coming from OTel
				continue
			}

			assert.False(t, seen)
			seen = true

			assert.Equal(t, zap.InfoLevel, entry.Level)
			assert.Equal(t, "test span logging", entry.Message)
			assert.Contains(t, entry.Context, zap.String("span.id", spanID))
			assert.Contains(t, entry.Context, zap.String("trace.id", traceID))
		}
	})

	t.Run("Check spans recorded", func(t *testing.T) {
		require.NoError(t, observability.ShutdownOtel(ctx))

		// Check that the mock server received the span.
		var spans []*otlptracepb.Span
		require.Eventually(
			t,
			func() bool {
				spans = traceService.getSpans()
				return len(spans) > 0
			},
			5*time.Second,
			100*time.Millisecond,
			"Spans were not recorded in time",
		)
		assert.Equal(t, "test-span", spans[0].Name)

		// Check that the mock server received the metric.
		metrics := metricService.getMetrics()
		require.NotEmpty(t, metrics)
		assert.Equal(t, "test-counter", metrics[0].Name)
	})
}

type mockTracerService struct {
	tracepb.UnimplementedTraceServiceServer
	spans []*otlptracepb.Span
	lock  sync.Mutex
}

// Export implements the Export method of the OTLP trace service.
func (s *mockTracerService) Export(_ context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, resourceSpans := range req.GetResourceSpans() {
		for _, scopeSpans := range resourceSpans.GetScopeSpans() {
			s.spans = append(s.spans, scopeSpans.GetSpans()...)
		}
	}
	return &tracepb.ExportTraceServiceResponse{}, nil
}

func (s *mockTracerService) getSpans() []*otlptracepb.Span {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.spans
}

// mockMetricService implements the OTLP metric service.
type mockMetricService struct {
	metricpb.UnimplementedMetricsServiceServer
	metrics []*otlpmetricpb.Metric
	lock    sync.Mutex
}

// Export implements the Export method of the OTLP metric service.
func (s *mockMetricService) Export(_ context.Context, req *metricpb.ExportMetricsServiceRequest) (*metricpb.ExportMetricsServiceResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, resourceMetrics := range req.GetResourceMetrics() {
		for _, scopeMetrics := range resourceMetrics.GetScopeMetrics() {
			s.metrics = append(s.metrics, scopeMetrics.GetMetrics()...)
		}
	}
	return &metricpb.ExportMetricsServiceResponse{}, nil
}

func (s *mockMetricService) getMetrics() []*otlpmetricpb.Metric {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.metrics
}

// startMockOtlpServer starts a gRPC server with the mock OTLP trace and metric services.
func startMockOtlpServer(t *testing.T, endpoint string) (*grpc.Server, *mockTracerService, *mockMetricService) {
	lis, err := net.Listen("tcp", endpoint)
	require.NoError(t, err)

	server := grpc.NewServer()
	traceService := &mockTracerService{}
	metricService := &mockMetricService{}
	tracepb.RegisterTraceServiceServer(server, traceService)
	metricpb.RegisterMetricsServiceServer(server, metricService)

	go func() {
		err := server.Serve(lis)
		if err != nil {
			t.Logf("Mock OTLP server failed: %v", err)
			t.Fail()
		}
	}()

	return server, traceService, metricService
}
