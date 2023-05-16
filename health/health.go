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

package health

import (
	"errors"
	"fmt"
	"sync"

	"github.com/elastic/beats/v7/libbeat/management"
)

type StatusReporter interface {
	UpdateStatus(status management.Status, msg string)
}

var Reporter = &reporter{
	ch:     make(chan error, 1),
	errors: map[string]error{},
	mut:    sync.RWMutex{},
}

type reporter struct {
	ch     chan error
	errors map[string]error
	mut    sync.RWMutex
}

func (r *reporter) NewHealth(component string, err error) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.errors[component] = err
	r.ch <- nil
}

func (r *reporter) Report(statusReporter StatusReporter) {
	var status management.Status
	err := r.getHealth()
	if err != nil {
		status = management.Degraded
	} else {
		status = management.Running
	}

	statusReporter.UpdateStatus(status, err.Error())
}

func (r *reporter) getHealth() error {
	r.mut.RLock()
	defer r.mut.RUnlock()
	list := make([]error, 0, len(r.errors))
	for c, err := range r.errors {
		if err != nil {
			list = append(list, fmt.Errorf("component %s is unhealthy: %w", c, err))
		}
	}

	return errors.Join(list...)
}

func (r *reporter) Channel() <-chan error {
	return r.ch
}

func (r *reporter) Close() {
	close(r.ch)
}
