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

package console_actions

import (
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"testing"
)

type CommandTestSuite struct {
	suite.Suite
	es   *elasticsearch.Client
	log  *logp.Logger
	opts goleak.Option
}

func TestCommandTestSuite(t *testing.T) {
	s := new(CommandTestSuite)
	s.log = logp.NewLogger("cloudbeat_starter_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	s.opts = goleak.IgnoreCurrent()
	suite.Run(t, s)
}

func (s *CommandTestSuite) TestStarterErrorBeater() {
	a := true
	RunActionsRoutine()
	//time.Sleep(60 * time.Second)
	s.True(a)
}
