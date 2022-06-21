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

func (o *OpaEvaluator) Decision(ctx context.Context, input interface{}) (interface{}, error) {
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

func (o *OpaEvaluator) Stop(ctx context.Context) {
	o.opa.Stop(ctx)
	err := o.bundleServer.Shutdown(ctx)
	if err != nil {
		o.log.Errorf("Could not stop OPA evaluator: %v", err)
	}
}

func (o *OpaEvaluator) Decode(result interface{}) ([]Finding, error) {
	var opaResult RuleResult
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &opaResult})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(result)
	return opaResult.Findings, err
}

func newEvaluatorLogger() logging.Logger {
	opaLogger := logging.New()
	opaLogger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "log.level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFile:  "log.origin",
		},
	})
	return opaLogger.WithFields(map[string]interface{}{
		"log.logger": "opa",
	})
}
