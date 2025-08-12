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
	"errors"
	"testing"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type RateLimiterTestSuite struct {
	suite.Suite
	rateLimiter *AssetsInventoryRateLimiter
}

func TestInventoryRateLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}

func (s *RateLimiterTestSuite) SetupTest() {
	logger := testhelper.NewLogger(s.T())
	s.rateLimiter = NewAssetsInventoryRateLimiter(logger)
}

func (s *RateLimiterTestSuite) TestRateLimiterWait() {
	t := s.T()
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

func TestGAXCallOptionRetrier(t *testing.T) {
	log := testhelper.NewObserverLogger(t)
	r := GAXCallOptionRetrier(log)
	settings := gax.CallSettings{}
	r.Resolve(&settings)

	c := []codes.Code{
		codes.ResourceExhausted,
		codes.DeadlineExceeded,
		codes.Unavailable,
	}

	for _, code := range c {
		pause, shouldRetry := settings.Retry().Retry(status.New(code, "error").Err())
		require.True(t, shouldRetry)
		require.True(t, pause < 1*time.Minute && pause > 0)
	}

	// should not retry
	errs := []error{status.New(codes.Aborted, "error").Err(), status.New(codes.Canceled, "error").Err(), errors.New("error")}
	for _, err := range errs {
		pause, shouldRetry := settings.Retry().Retry(err)
		require.False(t, shouldRetry)
		require.Equal(t, time.Duration(0), pause)
	}

	logs := logp.ObserverLogs().FilterMessageSnippet("gax retryer attempt").All()
	assert.Len(t, logs, len(c))
}
