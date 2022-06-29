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
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type ResoladerTestSuite struct {
	suite.Suite

	log    *logp.Logger
	ctx    context.Context
	cancel context.CancelFunc
	sut    *Listener
	opts   goleak.Option
}

func TestResoladerTestSuite(t *testing.T) {
	s := new(ResoladerTestSuite)
	suite.Run(t, s)
}

func (s *ResoladerTestSuite) SetupTest() {
	s.log = logp.NewLogger("cloudbeat_listener_test_suite")
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.opts = goleak.IgnoreCurrent()

	s.sut = NewListener(s.ctx, s.log)
}

func (s *ResoladerTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *ResoladerTestSuite) TestEmptyReload() {
	go func() {
		s.sut.Reload([]*reload.ConfigWithMeta{})
	}()
	var re *config.C
	select {
	case <-time.After(time.Second):
	case re = <-s.sut.Channel():
	}

	s.Nil(re)
}

func (s *ResoladerTestSuite) TestCancelBeforeReload() {
	meta := mapstr.NewPointer(mapstr.M{})
	conf, err := config.NewConfigFrom(map[string]string{
		"test": "test",
	})
	s.NoError(err)

	s.cancel()
	go func() {
		s.sut.Reload([]*reload.ConfigWithMeta{
			{
				Config: conf,
				Meta:   &meta,
			},
		})
	}()
}

func (s *ResoladerTestSuite) TestCancelAfterReload() {
	meta := mapstr.NewPointer(mapstr.M{})
	conf, err := config.NewConfigFrom(map[string]string{
		"test": "test",
	})
	s.NoError(err)

	go func() {
		s.sut.Reload([]*reload.ConfigWithMeta{
			{
				Config: conf,
				Meta:   &meta,
			},
		})
	}()
	s.cancel()
}

func (s *ResoladerTestSuite) TestSingleReload() {
	meta := mapstr.NewPointer(mapstr.M{})
	conf, err := config.NewConfigFrom(map[string]string{
		"test": "test",
	})
	s.NoError(err)

	values := []*reload.ConfigWithMeta{
		{
			Config: conf,
			Meta:   &meta,
		},
	}
	go func() {
		s.sut.Reload(values)
	}()

	re := <-s.sut.Channel()
	test, err := re.String("test", -1)
	s.NoError(err)
	s.Equal("test", test)
}
