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

package flavors

import (
	"context"
	"time"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type RepeaterFunc func() error

func NewRepeater(log *clog.Logger, interval time.Duration) *Repeater {
	return &Repeater{
		log:      log,
		interval: interval,
	}
}

type Repeater struct {
	log      *clog.Logger
	interval time.Duration
}

func (r *Repeater) Run(ctx context.Context, fn RepeaterFunc) error {
	r.log.Warn("Repeater ticker running for period ", r.interval)
	ticker := time.NewTicker(r.interval)
	immediate := time.NewTimer(0)

	for {
		select {
		case <-ctx.Done():
			r.log.Warnf("Repeater context is done: %v", ctx.Err())
			return nil
		case <-immediate.C:
		case <-ticker.C:
		}

		r.log.Warn("Repeater cycle triggered")
		err := fn()
		if err != nil {
			return err
		}
	}
}
