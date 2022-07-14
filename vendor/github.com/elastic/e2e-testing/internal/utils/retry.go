// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package utils

import (
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/elastic/e2e-testing/internal/shell"
)

// TimeoutFactor a multiplier for the max timeout when doing backoff retries.
// It can be overriden by TIMEOUT_FACTOR env var
var TimeoutFactor = 3

func init() {
	TimeoutFactor = shell.GetEnvInteger("TIMEOUT_FACTOR", TimeoutFactor)
}

// GetExponentialBackOff returns a preconfigured exponential backoff instance
func GetExponentialBackOff(elapsedTime time.Duration) *backoff.ExponentialBackOff {
	var (
		initialInterval     = 500 * time.Millisecond
		randomizationFactor = 0.5
		multiplier          = 2.0
		maxInterval         = 5 * time.Second
		maxElapsedTime      = elapsedTime
	)

	exp := backoff.NewExponentialBackOff()
	exp.InitialInterval = initialInterval
	exp.RandomizationFactor = randomizationFactor
	exp.Multiplier = multiplier
	exp.MaxInterval = maxInterval
	exp.MaxElapsedTime = maxElapsedTime

	return exp
}
