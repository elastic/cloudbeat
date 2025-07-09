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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func AppendAWSMiddlewares(awsConfig *aws.Config) {
	otelaws.AppendMiddlewares(
		&awsConfig.APIOptions,
		otelaws.WithAttributeBuilder(otelaws.DefaultAttributeBuilder, ensureSpanName),
	)
}

func ensureSpanName(ctx context.Context, _ middleware.InitializeInput, _ middleware.InitializeOutput) []attribute.KeyValue {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		if awsmiddleware.GetServiceID(ctx) == "" && awsmiddleware.GetOperationName(ctx) == "" {
			// The OTel AWS instrumentation uses the service ID and operation name to set the span name.
			// If those are not set for some reason, we must set a default name to avoid having an empty span name which
			// causes Elastic APM to throw an error when showing the span in the UI.
			span.SetName("Unknown AWS API Call")
		}
	}
	return []attribute.KeyValue{}
}
