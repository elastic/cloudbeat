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

package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/sdk"
	"go.opentelemetry.io/otel/attribute"

	"github.com/elastic/cloudbeat/internal/config"
	dlogger "github.com/elastic/cloudbeat/internal/evaluator/debug_logger"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

var now = func() time.Time { return time.Now().UTC() }

const scopeName = "github.com/elastic/cloudbeat/internal/evaluator"

type OpaEvaluator struct {
	log       *clog.Logger
	opa       *sdk.OPA
	benchmark string
}

type OpaInput struct {
	fetching.Result
	Benchmark string `json:"benchmark,omitempty"`
}

var opaConfig = `{
	"bundles": {
		"CSP": {
			"resource": "file://%s"
		}
	},
%s
}`

var logPlugin = `
	"decision_logs": {
		"plugin": "%s"
	},
	"plugins": {
		"%s": {}
	}`

func NewOpaEvaluator(ctx context.Context, log *clog.Logger, cfg *config.Config) (*OpaEvaluator, error) {
	// provide the OPA configuration which specifies
	// fetching policy bundle and logging decisions locally to the console

	log.Infof("OPA bundle path: %s", cfg.BundlePath)

	plugin := fmt.Sprintf(logPlugin, dlogger.PluginName, dlogger.PluginName)
	opaCfg := fmt.Sprintf(opaConfig, cfg.BundlePath, plugin)

	// create an instance of the OPA object
	opa, err := sdk.New(ctx, sdk.Options{
		Config:        bytes.NewReader([]byte(opaCfg)),
		Logger:        newLogger(),
		ConsoleLogger: newLogger(),
		Plugins: map[string]plugins.Factory{
			dlogger.PluginName: &dlogger.Factory{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("fail to init opa: %s", err.Error())
	}

	// Newer cloudbeat versions shouldn't look at deprecated runtime config values
	// and should always get benchmark values because the integration is auto updated with the stack
	var benchmark string
	if cfg.Benchmark != "" {
		// Assume that isSupportedBenchmark ran in config creation
		benchmark = cfg.Benchmark
	} else {
		log.Warn("no benchmark supplied")
	}

	log.Info("Successfully initiated OPA")
	return &OpaEvaluator{
		log:       log,
		opa:       opa,
		benchmark: benchmark,
	}, nil
}

func (o *OpaEvaluator) Eval(ctx context.Context, resourceInfo fetching.ResourceInfo) (EventData, error) {
	ctx, span := observability.StartSpan(ctx, scopeName, "OPA Eval")
	defer span.End()

	resMetadata, err := resourceInfo.GetMetadata()
	if err != nil {
		return EventData{}, observability.FailSpan(span, "failed to get resource metadata", err)
	}

	fetcherResult := fetching.Result{
		Type:     resMetadata.Type,
		SubType:  resMetadata.SubType,
		Resource: resourceInfo.GetData(),
	}

	result, err := o.decision(ctx, OpaInput{
		Result:    fetcherResult,
		Benchmark: o.benchmark,
	})
	if err != nil {
		return EventData{}, observability.FailSpan(span, "error running the policy", err)
	}

	ruleResults, err := o.decode(result)
	if err != nil {
		return EventData{}, observability.FailSpan(span, "error decoding findings", err)
	}

	span.SetAttributes(
		attribute.Int("findings.count", len(ruleResults.Findings)),
		attribute.String("resource.type", resMetadata.Type),
		attribute.String("resource.sub_type", resMetadata.SubType),
		attribute.String("resource.name", resMetadata.Name),
	)
	o.log.Debugf("Created %d findings for input: %v", len(ruleResults.Findings), fetcherResult)
	return EventData{ruleResults, resourceInfo}, nil
}

func (o *OpaEvaluator) Stop(ctx context.Context) {
	o.opa.Stop(ctx)
}

func (o *OpaEvaluator) decision(ctx context.Context, input OpaInput) (any, error) {
	// get the named policy decision for the specified input
	result, err := o.opa.Decision(ctx, sdk.DecisionOptions{
		Path:  "main",
		Input: input,
	})
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (o *OpaEvaluator) decode(result any) (RuleResult, error) {
	var opaResult RuleResult
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &opaResult})
	if err != nil {
		return RuleResult{}, err
	}

	err = decoder.Decode(result)
	opaResult.Metadata.CreatedAt = now()
	return opaResult, err
}
