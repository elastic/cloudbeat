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
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type HealthTestSuite struct {
	suite.Suite

	opts goleak.Option
}

func TestHealthTestSuite(t *testing.T) {
	s := new(HealthTestSuite)

	s.opts = goleak.IgnoreCurrent()
	suite.Run(t, s)
}

func (s *HealthTestSuite) TearDownTest() {
	// Verify no goroutines are leaking. Safest to keep this on top of the function.
	// Go defers are implemented as a LIFO stack. This should be the last one to run.
	goleak.VerifyNone(s.T(), s.opts)
}

func (s *HealthTestSuite) TestNewHealth() {
	r := &reporter{
		ch:     make(chan error, 1),
		errors: make(map[string]error),
	}

	events := []struct {
		component string
		err       error
		wantErr   bool
	}{
		{
			component: "component1",
			err:       nil,
			wantErr:   false,
		},
		{
			component: "component1",
			err:       errors.New("component1 went wrong"),
			wantErr:   true,
		},
		{
			component: "component2",
			err:       errors.New("component2 went wrong"),
			wantErr:   true,
		},
		{
			component: "component2",
			err:       nil,
			wantErr:   true,
		},
		{
			component: "component1",
			err:       nil,
			wantErr:   false,
		},
	}

	for _, e := range events {
		r.NewHealth(e.component, e.err)
		<-r.ch
		err := r.getHealth()
		if e.wantErr {
			s.Error(err)
			fmt.Println(err)
		} else {
			s.NoError(err)
		}
	}
}

func (s *HealthTestSuite) TestParallelNewHealth() {
	r := &reporter{
		ch:     make(chan error),
		errors: make(map[string]error),
	}

	events := []struct {
		component string
		err       error
	}{
		{
			component: "component1",
			err:       nil,
		},
		{
			component: "component1",
			err:       errors.New("went wrong"),
		},
		{
			component: "component1",
			err:       errors.New("component went wrong"),
		},
		{
			component: "component1",
			err:       nil,
		},
		{
			component: "component1",
			err:       errors.New("some error"),
		},
	}

	for _, e := range events {
		go r.NewHealth(e.component, e.err)
	}

	for i := 0; i < len(events); i++ {
		<-r.ch
	}
}
