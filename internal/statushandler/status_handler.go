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
	"slices"
	"strings"
	"sync"

	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/samber/lo"
)

const messagesBufferSize = 10

type StatusHandler struct {
	statusReporter status.StatusReporter

	m        sync.Mutex
	messages map[string]struct{}
}

func NewStatusHandler(statusReporter status.StatusReporter) *StatusHandler {
	s := &StatusHandler{
		statusReporter: statusReporter,
		messages:       make(map[string]struct{}, messagesBufferSize),
	}

	return s
}

func (s *StatusHandler) refreshBuffer() {
	if len(s.messages) > messagesBufferSize {
		s.messages = make(map[string]struct{}, messagesBufferSize)
		return
	}

	// Just clear existing map — keeps allocated buckets
	clear(s.messages)
}

func (s *StatusHandler) Degraded(message string) {
	s.m.Lock()
	defer s.m.Unlock()

	// if the same message has already been reported, do nothing
	if _, exists := s.messages[message]; exists {
		return
	}
	s.messages[message] = struct{}{}

	sl := lo.Keys(s.messages)
	slices.Sort(sl)
	s.statusReporter.UpdateStatus(status.Degraded, strings.Join(sl, "\n"))
}

func (s *StatusHandler) Reset() {
	s.m.Lock()
	defer s.m.Unlock()

	s.statusReporter.UpdateStatus(status.Running, "")
	s.refreshBuffer()
}

type StatusHandlerAPI interface {
	Degraded(message string)
	Reset()
}

var _ (StatusHandlerAPI) = (*StatusHandler)(nil)

type NOOP struct{}

func (NOOP) Degraded(_ string) {}
func (NOOP) Reset()            {}

var _ (StatusHandlerAPI) = NOOP{}
