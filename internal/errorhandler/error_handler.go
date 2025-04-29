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

package errorhandler

import (
	"context"
	"sync"

	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/elastic/cloudbeat/internal/infra/clog"
)

const DefaultErrorHandlerBufferSize = 10

type ErrorHandler struct {
	statusReporter status.StatusReporter

	processorsMutex sync.Mutex
	processors      []func(status.StatusReporter, error)

	ch    chan error
	close chan struct{}

	log *clog.Logger
}

func NewErrorHandler(statusReporter status.StatusReporter, bufferSize int, log *clog.Logger) *ErrorHandler {
	return &ErrorHandler{
		statusReporter: statusReporter,
		ch:             make(chan error, bufferSize),
		close:          make(chan struct{}),
		log:            log,
	}
}

func (r *ErrorHandler) Register(errorProcessor func(status.StatusReporter, error)) {
	r.processorsMutex.Lock()
	defer r.processorsMutex.Unlock()
	r.processors = append(r.processors, errorProcessor)
}

func (r *ErrorHandler) Start(ctx context.Context) {
	go func() {
		r.log.Info("error handler processing started")
		for {
			select {
			case err, ok := <-r.ch:
				if !ok {
					return
				}
				r.log.Info("error handler: error received")
				r.process(ctx, err)
			case <-ctx.Done():
				return
			case <-r.close:
				return
			}
		}
	}()
}

func (r *ErrorHandler) Stop() {
	close(r.close)
}

func (r *ErrorHandler) process(ctx context.Context, err error) {
	r.processorsMutex.Lock()
	defer r.processorsMutex.Unlock()
	for _, h := range r.processors {
		if ctx.Err() != nil {
			return
		}
		h(r.statusReporter, err)
	}
}

func (r *ErrorHandler) Publish(ctx context.Context, err error) {
	select {
	case r.ch <- err:
	case <-ctx.Done():
	}
}

func (r *ErrorHandler) Reset(_ context.Context) {
	r.statusReporter.UpdateStatus(status.Running, "")
}

type ErrorPublisher interface {
	Publish(ctx context.Context, err error)
}

// StatusReporter provides a method to update current status of a unit.
type StatusReporter interface {
	// UpdateStatus updates the status of the unit.
	UpdateStatus(status status.Status, msg string)
}
