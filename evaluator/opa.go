package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/beater/bundle"
	"github.com/mitchellh/mapstructure"
	"github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/sdk"
	"github.com/sirupsen/logrus"
	"net/http"
)

type OpaEvaluator struct {
	opa          *sdk.OPA
	bundleServer *http.Server
}

func NewOpaEvaluator(ctx context.Context) (Evaluator, error) {
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
	o.bundleServer.Shutdown(ctx)
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
	opaLogger.SetFormatter(&logrus.JSONFormatter{})
	return opaLogger.WithFields(map[string]interface{}{"goroutine": "opa"})
}
