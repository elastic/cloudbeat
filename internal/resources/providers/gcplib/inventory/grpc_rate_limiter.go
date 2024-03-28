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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var RetryOnResourceExhausted = gax.WithRetry(func() gax.Retryer {
	return gax.OnCodes([]codes.Code{codes.ResourceExhausted}, gax.Backoff{
		Initial:    1 * time.Second,
		Max:        1 * time.Minute,
		Multiplier: 1.2,
	})
})

type AssetsInventoryRateLimiter struct {
	methods map[string]*rate.Limiter
	log     *logp.Logger
}

// https://cloud.google.com/asset-inventory/docs/quota
var methods = map[string]*rate.Limiter{
	// In both 'single-account' and 'organization-account' cases, we always need to pace by project because of the consumer project (quota_project_id)
	// Which is the one effectively consuming the quota
	// the organization quota would be relevant if we manually send multiple requests with diff quota_project_id, which we don't do
	"/google.cloud.asset.v1.AssetService/ListAssets": rate.NewLimiter(rate.Every(time.Minute/100), 1),
}

func NewAssetsInventoryRateLimiter(log *logp.Logger) *AssetsInventoryRateLimiter {
	return &AssetsInventoryRateLimiter{
		log:     log,
		methods: methods,
	}
}

func (rl *AssetsInventoryRateLimiter) Wait(ctx context.Context, method string) {
	limiter := rl.methods[method]
	if limiter != nil {
		err := limiter.Wait(ctx)
		if err != nil {
			rl.log.Errorf("Failed to wait for project quota on method %s, error: %w", method, err)
		}
	}

}

func (rl *AssetsInventoryRateLimiter) GetInterceptorDialOption() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		rl.Wait(ctx, method)
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
