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

package statushandler

import (
	"strconv"
	"testing"

	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatusHandler(t *testing.T) {
	m := newMockStatusReporter(t)
	m.On("UpdateStatus", status.Degraded, "abc").Once()

	sh := NewStatusHandler(m)
	sh.Degraded("abc")
	sh.Degraded("abc")
	sh.Degraded("abc")

	t.Run("same message does not trigger status call", func(t *testing.T) {
		m := newMockStatusReporter(t)
		m.On("UpdateStatus", status.Degraded, "abc").Once()

		sh := NewStatusHandler(m)
		sh.Degraded("abc")
		sh.Degraded("abc")
		sh.Degraded("abc")
	})

	t.Run("unique messages", func(t *testing.T) {
		m := newMockStatusReporter(t)
		c1 := m.On("UpdateStatus", status.Degraded, "abc").Once()
		c2 := m.On("UpdateStatus", status.Degraded, "abc\nzzz").Once().NotBefore(c1)
		m.On("UpdateStatus", status.Running, "").Once().NotBefore(c2)

		sh := NewStatusHandler(m)
		sh.Degraded("abc")
		sh.Degraded("abc")
		sh.Degraded("abc")
		sh.Degraded("zzz")
		sh.Degraded("zzz")
		sh.Degraded("zzz")
		sh.Reset()
	})

	t.Run("reset map", func(t *testing.T) {
		m := newMockStatusReporter(t)
		m.On("UpdateStatus", mock.Anything, mock.Anything).Maybe()

		sh := NewStatusHandler(m)

		for i := range 50 {
			sh.Degraded(strconv.Itoa(i))
		}

		require.Len(t, sh.messages, 50)
		sh.Reset()
		require.Empty(t, sh.messages)
	})
}

type mockStatusReporter struct {
	mock.Mock
}

func newMockStatusReporter(t *testing.T) *mockStatusReporter {
	m := &mockStatusReporter{}
	t.Cleanup(func() {
		m.AssertExpectations(t)
	})
	return m
}

func (m *mockStatusReporter) UpdateStatus(status status.Status, msg string) {
	m.Called(status, msg)
}
