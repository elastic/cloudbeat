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

package launcher

import (
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/elastic-agent-libs/config"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type Listener struct {
	log  *clog.Logger
	done chan struct{}
	ch   chan *config.C
}

func (l *Listener) Reload(configs []*reload.ConfigWithMeta) error {
	if len(configs) == 0 {
		return nil
	}

	l.log.Infof("Listener received %v new configs for reload", len(configs))

	// TODO(yashtewari): Based on limitations elsewhere, such as the CSP integration,
	// don't think we should receive more than one Config here. Need to confirm and handle.
	data := configs[len(configs)-1].Config

	select {
	case <-l.done:
	case l.ch <- data:
	}

	return nil
}

func (l *Listener) Channel() <-chan *config.C {
	return l.ch
}

func (l *Listener) Stop() {
	l.log.Info("Listener is about to stop")
	close(l.done)
	close(l.ch)
}

func NewListener(log *clog.Logger) *Listener {
	return &Listener{
		log:  log,
		done: make(chan struct{}),
		ch:   make(chan *config.C),
	}
}
