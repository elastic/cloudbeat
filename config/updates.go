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

package config

import (
	"context"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/beats/v7/libbeat/logp"
)

type reloader struct {
	ctx context.Context
	log *logp.Logger
	ch  chan<- *common.Config
}

func (r *reloader) Reload(configs []*reload.ConfigWithMeta) error {
	if len(configs) == 0 {
		return nil
	}

	r.log.Infof("Received %v new configs for reload.", len(configs))

	select {
	case <-r.ctx.Done():
	default:
		// TODO(yashtewari): Based on limitations elsewhere, such as the CSP integration,
		// don't think we should receive more than one Config here. Need to confirm and handle.
		r.ch <- configs[len(configs)-1].Config
	}

	return nil
}

func Updates(ctx context.Context, log *logp.Logger) <-chan *common.Config {
	ch := make(chan *common.Config)
	r := &reloader{
		ctx: ctx,
		log: log,
		ch:  ch,
	}

	reload.Register.MustRegisterList("inputs", r)

	return ch
}
