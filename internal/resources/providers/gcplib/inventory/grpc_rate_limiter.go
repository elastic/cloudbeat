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
	"time"

	"github.com/googleapis/gax-go/v2"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

// retryer wraps another gax Retryer and logs each retry attempt.
type retryer struct {
	gaxRetryer gax.Retryer
	log        *clog.Logger
}

func (r retryer) Retry(err error) (time.Duration, bool) {
	pause, shouldRetry := r.gaxRetryer.Retry(err)

	if shouldRetry {
		s := ""
		// This should always be not nil, but let's be extra cautious.
		if err != nil {
			s = err.Error()
		}
		r.log.Debugf("gax Retryer retries based on error (%s) pause %s", s, pause.String())
	}

	return pause, shouldRetry
}

func GAXCallOptionRetrier(log *clog.Logger) gax.CallOption {
	return gax.WithRetry(func() gax.Retryer {
		return retryer{
			log: log,
			gaxRetryer: gax.OnCodes([]codes.Code{
				codes.ResourceExhausted,
				codes.DeadlineExceeded,
				codes.Unavailable,
			}, gax.Backoff{
				Initial:    1 * time.Second,
				Max:        1 * time.Minute,
				Multiplier: 1.2,
			}),
		}
	})
}

type AssetsInventoryRateLimiter struct {
	methods map[string]*rate.Limiter
	log     *clog.Logger
}

// a map of asset inventory client methods and their quotas.
// see https://cloud.google.com/asset-inventory/docs/quota
var methods = map[string]*rate.Limiter{
	// using per-project quota suffices for both single-account and organization-account, because it's more restrictive.
	"/google.cloud.asset.v1.AssetService/ListAssets": rate.NewLimiter(rate.Every(time.Minute/100), 1),
}

func NewAssetsInventoryRateLimiter(log *clog.Logger) *AssetsInventoryRateLimiter {
	return &AssetsInventoryRateLimiter{
		log:     log,
		methods: methods,
	}
}

// Limits the rate of the method calls defined in the methods map.
func (rl *AssetsInventoryRateLimiter) Wait(ctx context.Context, method string, req any) {
	limiter := rl.methods[method]
	if limiter != nil {
		err := limiter.Wait(ctx)
		if err != nil {
			rl.log.Errorf("Failed to wait for project quota on method: %s, request: %v, error: %v", method, req, err)
		}
	}
}

// Returns a grpc.DialOption that intercepts the unary RPCs and limits the rate of the method calls.
func (rl *AssetsInventoryRateLimiter) GetInterceptorDialOption() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		rl.Wait(ctx, method, req)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
