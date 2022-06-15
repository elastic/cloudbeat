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
	"github.com/elastic/cloudbeat/resources/fetching"
	"net/http"

	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/sirupsen/logrus"

	"github.com/elastic/cloudbeat/beater/bundle"
)

type OpaEvaluator struct {
	log          *logp.Logger
	opa          *sdk.OPA
	bundleServer *http.Server
}

func NewOpaEvaluator(ctx context.Context, log *logp.Logger) (Evaluator, error) {
	server, err := bundle.StartServer()
	if err != nil {
		return nil, err
	}

	// provide the OPA configuration which specifies
	// fetching policy bundles from the mock bundleServer
	// and logging decisions locally to the console
	config := []byte(fmt.Sprintf(bundle.Config, bundle.ServerAddress))

	// create an instance of the OPA object
	opaLogger := newEvaluatorLogger()
	opa, err := sdk.New(ctx, sdk.Options{
		Config: bytes.NewReader(config),
		Logger: opaLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to init opa: %s", err.Error())
	}

	return &OpaEvaluator{
		log:          log,
		opa:          opa,
		bundleServer: server,
	}, nil
}

func (o *OpaEvaluator) Eval(ctx context.Context, resourceInfo fetching.ResourceInfo) (EventData, error) {
	fetcherResult := fetching.Result{
		Type:     resourceInfo.GetMetadata().Type,
		Resource: resourceInfo.GetData(),
	}

	result, err := o.decision(ctx, fetcherResult)
	if err != nil {
		return EventData{}, fmt.Errorf("error running the policy: %v", err)
	}

	o.log.Debugf("Eval decision for input: %v -- %v", fetcherResult, result)
	ruleResults, err := o.decode(result)
	if err != nil {
		return EventData{}, fmt.Errorf("error decoding findings: %v", err)
	}

	o.log.Debugf("Created %d findings for input: %v", len(ruleResults.Findings), fetcherResult)
	return EventData{ruleResults, resourceInfo}, nil
}

func (o *OpaEvaluator) Stop(ctx context.Context) {
	o.opa.Stop(ctx)
	err := o.bundleServer.Shutdown(ctx)
	if err != nil {
		o.log.Errorf("Could not stop OPA evaluator: %v", err)
	}
}

func (o *OpaEvaluator) decision(ctx context.Context, input interface{}) (interface{}, error) {
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
	return opaResult, err
}

func newEvaluatorLogger() logging.Logger {
	opaLogger := logging.New()
	opaLogger.SetFormatter(&logrus.JSONFormatter{})
	return opaLogger.WithFields(map[string]interface{}{"goroutine": "opa"})
}
