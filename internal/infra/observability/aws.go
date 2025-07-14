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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ensureSpanNameProcessor is an OpenTelemetry span processor that ensures no span has an empty name.
// On AWS multi-account onboarding, otelaws has an edge case where there's no span name when the ec2imds service cannot
// be reached:
// > Region 'us-east-1' selected after failure to retrieve aws regions: operation error ec2imds:
// > GetInstanceIdentityDocument, canceled, context deadline exceeded
// Which in turn fails with `Error while fetching resource` `Missing required fields span.name in event (500)` in
// Elastic APM.
// This processor works around this by setting a default span name when the span name is empty.
type ensureSpanNameProcessor struct{}

func (a ensureSpanNameProcessor) OnStart(_ context.Context, s sdktrace.ReadWriteSpan) {
	if s.Name() == "" { // Empty span names are not allowed in Elastic APM.
		s.SetName(defaultSpanName(s))
	}
}

func defaultSpanName(s sdktrace.ReadWriteSpan) string {
	if strings.Contains(s.InstrumentationScope().Name, "otelaws") {
		return "Unknown AWS API Call"
	}
	return "Anonymous Span"
}

func (a ensureSpanNameProcessor) OnEnd(sdktrace.ReadOnlySpan)      {}
func (a ensureSpanNameProcessor) Shutdown(context.Context) error   { return nil }
func (a ensureSpanNameProcessor) ForceFlush(context.Context) error { return nil }

func AppendAWSMiddlewares(awsConfig *aws.Config) {
	otelaws.AppendMiddlewares(&awsConfig.APIOptions)
}
