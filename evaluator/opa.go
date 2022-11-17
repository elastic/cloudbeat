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

	"github.com/elastic/cloudbeat/config"
	dlogger "github.com/elastic/cloudbeat/evaluator/debug_logger"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/sdk"
)

var now = func() time.Time { return time.Now().UTC() }

type OpaEvaluator struct {
	log            *logp.Logger
	opa            *sdk.OPA
	activatedRules *config.Benchmarks
}

type OpaInput struct {
	fetching.Result
	ActivatedRules *config.Benchmarks `json:"activated_rules,omitempty"`
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

func NewOpaEvaluator(ctx context.Context, log *logp.Logger, cfg config.Config) (Evaluator, error) {
	// provide the OPA configuration which specifies
	// fetching policy bundle and logging decisions locally to the console

	log.Infof("OPA bundle path: %s", cfg.BundlePath)

	plugin := fmt.Sprintf(logPlugin, dlogger.PluginName, dlogger.PluginName)
	opaCfg := fmt.Sprintf(opaConfig, cfg.BundlePath, plugin)

	decisonLogger := newLogger()
	stdLogger := newLogger()

	// create an instance of the OPA object
	opa, err := sdk.New(ctx, sdk.Options{
		Config:        bytes.NewReader([]byte(opaCfg)),
		Logger:        stdLogger,
		ConsoleLogger: decisonLogger,
		Plugins: map[string]plugins.Factory{
			dlogger.PluginName: &dlogger.Factory{},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("fail to init opa: %s", err.Error())
	}

	var rules *config.Benchmarks
	if cfg.RuntimeCfg != nil {
		rules = cfg.RuntimeCfg.ActivatedRules
	} else {
		log.Warn("no runtime config supplied")
	}

	log.Info("Successfully initiated OPA")
	return &OpaEvaluator{
		log:            log,
		opa:            opa,
		activatedRules: rules,
	}, nil
}

func (o *OpaEvaluator) Eval(ctx context.Context, resourceInfo fetching.ResourceInfo) (EventData, error) {
	resMetadata, err := resourceInfo.GetMetadata()
	if err != nil {
		return EventData{}, fmt.Errorf("failed to get resource metadata: %v", err)
	}

	fetcherResult := fetching.Result{
		Type:     resMetadata.Type,
		SubType:  resMetadata.SubType,
		Resource: resourceInfo.GetData(),
	}

	result, err := o.decision(ctx, OpaInput{
		Result:         fetcherResult,
		ActivatedRules: o.activatedRules,
	})

	if err != nil {
		return EventData{}, fmt.Errorf("error running the policy: %v", err)
	}

	ruleResults, err := o.decode(result)
	if err != nil {
		return EventData{}, fmt.Errorf("error decoding findings: %v", err)
	}

	o.log.Debugf("Created %d findings for input: %v", len(ruleResults.Findings), fetcherResult)
	return EventData{ruleResults, resourceInfo}, nil
}

func (o *OpaEvaluator) Stop(ctx context.Context) {
	o.opa.Stop(ctx)
}

func (o *OpaEvaluator) decision(ctx context.Context, input OpaInput) (interface{}, error) {
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

func (o *OpaEvaluator) decode(result interface{}) (RuleResult, error) {
	var opaResult RuleResult
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &opaResult})
	if err != nil {
		return RuleResult{}, err
	}

	err = decoder.Decode(result)
	opaResult.Metadata.CreatedAt = now()
	return opaResult, err
}
