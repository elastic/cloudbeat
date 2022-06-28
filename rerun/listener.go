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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package rerun

import (
	"context"

	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Listener struct {
	ctx context.Context
	log *logp.Logger
	ch  chan *config.C
}

func (r *Listener) Reload(configs []*reload.ConfigWithMeta) error {
	if len(configs) == 0 {
		return nil
	}

	r.log.Infof("Received %v new configs for reload.", len(configs))

	// TODO(yashtewari): Based on limitations elsewhere, such as the CSP integration,
	// don't think we should receive more than one Config here. Need to confirm and handle.
	data := configs[len(configs)-1].Config

	select {
	case <-r.ctx.Done():
	case r.ch <- data:
	}

	return nil
}

func (r *Listener) Channel() <-chan *config.C {
	return r.ch
}

func NewListener(ctx context.Context, log *logp.Logger) *Listener {
	ch := make(chan *config.C)
	return &Listener{
		ctx: ctx,
		log: log,
		ch:  ch,
	}
}
