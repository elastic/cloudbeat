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
	"sync"
	"testing"

	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type ListenerTestSuite struct {
	suite.Suite

	opts goleak.Option
}

func TestListenerTestSuite(t *testing.T) {
	s := new(ListenerTestSuite)
	suite.Run(t, s)
}

func (s *ListenerTestSuite) SetupTest() {
	s.opts = goleak.IgnoreCurrent()
}

func (s *ListenerTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

var conf = config.MustNewConfigFrom(map[string]string{
	"foo": "bar",
})

func (s *ListenerTestSuite) TestReloadAndStop() {
	type configUpdate []*reload.ConfigWithMeta
	type incomingConfigs struct {
		values []configUpdate
		name   string
	}

	meta := mapstr.NewPointer(mapstr.M{})

	tests := []incomingConfigs{
		{
			name:   "no configs",
			values: []configUpdate{},
		},
		{
			name: "single empty config",
			values: []configUpdate{
				{},
			},
		},
		{
			name: "multiple empty configs",
			values: []configUpdate{
				{},
				{},
			},
		},
		{
			name: "single config",
			values: []configUpdate{
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
			},
		},
		{
			name: "single config with length",
			values: []configUpdate{
				{
					{},
					{},
					{
						Config: conf,
						Meta:   &meta,
					},
				},
			},
		},
		{
			name: "same config 3 times",
			values: []configUpdate{
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
			},
		},
		{
			name: "mixed updates",
			values: []configUpdate{
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
				{},
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
				{},
				{
					{
						Config: conf,
						Meta:   &meta,
					},
				},
			},
		},
	}

	for _, tcase := range tests {
		s.Run(tcase.name, func() {
			sut := NewListener(testhelper.NewLogger(s.T()))
			wg := sync.WaitGroup{}

			for _, val := range tcase.values {
				wg.Add(1)
				go func(listener *Listener, update configUpdate) {
					err := listener.Reload(update)
					s.Require().NoError(err)
					wg.Done()
				}(sut, val)

				if len(val) > 0 {
					re := <-sut.Channel()
					test, err := re.String("foo", -1)
					s.Require().NoError(err)
					s.Equal("bar", test)
				}
			}

			sut.Stop()
			wg.Wait()
		})
	}
}
