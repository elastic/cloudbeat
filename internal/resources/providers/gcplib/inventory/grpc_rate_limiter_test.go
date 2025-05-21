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

package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type RateLimiterTestSuite struct {
	suite.Suite
	logger      *clog.Logger
	rateLimiter *AssetsInventoryRateLimiter
}

func TestInventoryRateLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}

func (s *RateLimiterTestSuite) SetupTest() {
	s.logger = clog.NewLogger("test")
	s.rateLimiter = NewAssetsInventoryRateLimiter(s.logger)
}

func (s *RateLimiterTestSuite) TestRateLimiterWait() {
	ctx := t.Context()
	duration := time.Millisecond
	s.rateLimiter.methods = map[string]*rate.Limiter{
		"someMethod": rate.NewLimiter(rate.Every(duration/1), 1), // 1 request per duration
	}

	totalRequests := 5
	startTime := time.Now()
	for i := 0; i < totalRequests; i++ {
		s.rateLimiter.Wait(ctx, "someMethod", nil)
	}
	endTime := time.Now()

	actualDuration := endTime.Sub(startTime)
	minDuration := duration * time.Duration((totalRequests - 1)) // 1st request is instant, 2nd and above wait 1duration each
	s.GreaterOrEqual(actualDuration, minDuration)
}
