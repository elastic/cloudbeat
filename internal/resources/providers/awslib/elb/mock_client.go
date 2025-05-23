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

package elb

import (
	context "context"

	elasticloadbalancing "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	mock "github.com/stretchr/testify/mock"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// DescribeLoadBalancers provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockClient) DescribeLoadBalancers(_a0 context.Context, _a1 *elasticloadbalancing.DescribeLoadBalancersInput, _a2 ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DescribeLoadBalancers")
	}

	var r0 *elasticloadbalancing.DescribeLoadBalancersOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) *elasticloadbalancing.DescribeLoadBalancersOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*elasticloadbalancing.DescribeLoadBalancersOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_DescribeLoadBalancers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DescribeLoadBalancers'
type MockClient_DescribeLoadBalancers_Call struct {
	*mock.Call
}

// DescribeLoadBalancers is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *elasticloadbalancing.DescribeLoadBalancersInput
//   - _a2 ...func(*elasticloadbalancing.Options)
func (_e *MockClient_Expecter) DescribeLoadBalancers(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *MockClient_DescribeLoadBalancers_Call {
	return &MockClient_DescribeLoadBalancers_Call{Call: _e.mock.On("DescribeLoadBalancers",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *MockClient_DescribeLoadBalancers_Call) Run(run func(_a0 context.Context, _a1 *elasticloadbalancing.DescribeLoadBalancersInput, _a2 ...func(*elasticloadbalancing.Options))) *MockClient_DescribeLoadBalancers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*elasticloadbalancing.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*elasticloadbalancing.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*elasticloadbalancing.DescribeLoadBalancersInput), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_DescribeLoadBalancers_Call) Return(_a0 *elasticloadbalancing.DescribeLoadBalancersOutput, _a1 error) *MockClient_DescribeLoadBalancers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_DescribeLoadBalancers_Call) RunAndReturn(run func(context.Context, *elasticloadbalancing.DescribeLoadBalancersInput, ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)) *MockClient_DescribeLoadBalancers_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
