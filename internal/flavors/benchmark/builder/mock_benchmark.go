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

// Code generated by mockery v2.53.3. DO NOT EDIT.

package builder

import (
	context "context"

	beat "github.com/elastic/beats/v7/libbeat/beat"

	mock "github.com/stretchr/testify/mock"
)

// MockBenchmark is an autogenerated mock type for the Benchmark type
type MockBenchmark struct {
	mock.Mock
}

type MockBenchmark_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBenchmark) EXPECT() *MockBenchmark_Expecter {
	return &MockBenchmark_Expecter{mock: &_m.Mock}
}

// Run provides a mock function with given fields: ctx
func (_m *MockBenchmark) Run(ctx context.Context) (<-chan []beat.Event, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Run")
	}

	var r0 <-chan []beat.Event
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (<-chan []beat.Event, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) <-chan []beat.Event); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan []beat.Event)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBenchmark_Run_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Run'
type MockBenchmark_Run_Call struct {
	*mock.Call
}

// Run is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockBenchmark_Expecter) Run(ctx interface{}) *MockBenchmark_Run_Call {
	return &MockBenchmark_Run_Call{Call: _e.mock.On("Run", ctx)}
}

func (_c *MockBenchmark_Run_Call) Run(run func(ctx context.Context)) *MockBenchmark_Run_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockBenchmark_Run_Call) Return(_a0 <-chan []beat.Event, _a1 error) *MockBenchmark_Run_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockBenchmark_Run_Call) RunAndReturn(run func(context.Context) (<-chan []beat.Event, error)) *MockBenchmark_Run_Call {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with no fields
func (_m *MockBenchmark) Stop() {
	_m.Called()
}

// MockBenchmark_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type MockBenchmark_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *MockBenchmark_Expecter) Stop() *MockBenchmark_Stop_Call {
	return &MockBenchmark_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *MockBenchmark_Stop_Call) Run(run func()) *MockBenchmark_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockBenchmark_Stop_Call) Return() *MockBenchmark_Stop_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockBenchmark_Stop_Call) RunAndReturn(run func()) *MockBenchmark_Stop_Call {
	_c.Run(run)
	return _c
}

// NewMockBenchmark creates a new instance of MockBenchmark. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBenchmark(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBenchmark {
	mock := &MockBenchmark{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
