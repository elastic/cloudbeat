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
	"errors"
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/goleak"
)

type mockStatusReporter struct {
	mock.Mock
}

func (u *mockStatusReporter) UpdateStatus(s status.Status, msg string) {
	u.Called(s, msg)
}

func TestErrorHandler(t *testing.T) {
	defer goleak.VerifyNone(t)
	logger := testhelper.NewLogger(t)

	stop := func(eh *ErrorHandler) {
		// wait worker to consume the errors
		time.Sleep(5 * time.Millisecond)
		eh.Stop()
		time.Sleep(5 * time.Millisecond)
	}

	t.Run("happy path single publish", func(t *testing.T) {
		ctx, cnl := context.WithCancel(context.Background())
		defer cnl()

		m := &mockStatusReporter{}
		defer m.AssertExpectations(t)
		m.On("UpdateStatus", status.Degraded, "missing permission on cloud provider side: p1").Once()

		mockError := &MissingCSPPermissionError{Permission: "p1"}

		eh := NewErrorHandler(m, DefaultErrorHandlerBufferSize, logger)

		eh.Register(func(sr status.StatusReporter, err error) {
			assert.Equal(t, mockError, err)
			sr.UpdateStatus(status.Degraded, err.Error())
		})
		eh.Start(ctx)

		eh.Publish(ctx, mockError)

		stop(eh)
	})

	t.Run("happy path multiple publish with reset", func(t *testing.T) {
		const times = 5
		ctx, cnl := context.WithCancel(context.Background())
		defer cnl()

		m := &mockStatusReporter{}
		defer m.AssertExpectations(t)
		c1 := m.On("UpdateStatus", status.Degraded, "abc").Times(times)
		m.On("UpdateStatus", status.Running, "").Once().NotBefore(c1)

		mockError := errors.New("error")

		eh := NewErrorHandler(m, DefaultErrorHandlerBufferSize, logger)

		eh.Register(func(sr status.StatusReporter, err error) {
			assert.Equal(t, mockError, err)
			sr.UpdateStatus(status.Degraded, "abc")
		})
		eh.Start(ctx)

		for range times {
			eh.Publish(ctx, mockError)
		}

		// wait worker to consume the errors
		time.Sleep(5 * time.Millisecond)
		eh.Reset(ctx)

		stop(eh)
	})

	t.Run("stop with context cancel", func(t *testing.T) {
		ctx, cnl := context.WithCancel(context.Background())

		m := &mockStatusReporter{}
		defer m.AssertExpectations(t)

		eh := NewErrorHandler(m, DefaultErrorHandlerBufferSize, logger)

		eh.Start(ctx)

		cnl()
	})
}
