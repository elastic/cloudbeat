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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// TestNewResourceSurvivesSemconvUpgrade guards against reintroducing a hard-coded
// resource.WithSchemaURL in newResource. Pinning our own schema URL alongside the
// SDK's detectors makes resource.New fail with ErrSchemaURLConflict the moment
// go.opentelemetry.io/otel/sdk is built against a newer semconv, which silently
// breaks every dependency bump that pulls the SDK forward.
func TestNewResourceSurvivesSemconvUpgrade(t *testing.T) {
	res, err := newResource(t.Context())
	require.NoError(t, err, "newResource must not assert its own schema URL; see ErrSchemaURLConflict")
	require.NotNil(t, res)

	// The schema URL has to come from the SDK's detectors rather than from our
	// semconv import, otherwise the conflict above is only a bump away.
	assert.NotEmpty(t, res.SchemaURL(), "expected the SDK detectors to supply a schema URL")
	assert.Equal(t, resource.Default().SchemaURL(), res.SchemaURL())

	attrs := res.Attributes()
	assert.Contains(t, attrs, semconv.ServiceNameKey.String(serviceName))
}
