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
	"strings"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestOtel(t *testing.T) {
	ctx := t.Context()

	_, span := StartSpan(ctx, "test-tracer", "test-span")
	assert.False(t, span.IsRecording())

	log := testhelper.NewObserverLogger(t)
	ctx, err := SetUpOtel(ctx, log.Logger)
	require.NoError(t, err)

	ctx, span = StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()
	require.NotNil(t, ctx)
	assert.True(t, span.IsRecording())
	assert.True(t, span.SpanContext().IsValid())

	spanID := span.SpanContext().SpanID().String()
	traceID := span.SpanContext().TraceID().String()

	log.WithSpanContext(span.SpanContext()).Info("test span logging")

	logs := logp.ObserverLogs().TakeAll()
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
}
