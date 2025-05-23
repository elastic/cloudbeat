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

package s3

import (
	context "context"

	s3control "github.com/aws/aws-sdk-go-v2/service/s3control"
	mock "github.com/stretchr/testify/mock"
)

// MockControlClient is an autogenerated mock type for the ControlClient type
type MockControlClient struct {
	mock.Mock
}

type MockControlClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockControlClient) EXPECT() *MockControlClient_Expecter {
	return &MockControlClient_Expecter{mock: &_m.Mock}
}

// GetPublicAccessBlock provides a mock function with given fields: ctx, params, optFns
func (_m *MockControlClient) GetPublicAccessBlock(ctx context.Context, params *s3control.GetPublicAccessBlockInput, optFns ...func(*s3control.Options)) (*s3control.GetPublicAccessBlockOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetPublicAccessBlock")
	}

	var r0 *s3control.GetPublicAccessBlockOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *s3control.GetPublicAccessBlockInput, ...func(*s3control.Options)) (*s3control.GetPublicAccessBlockOutput, error)); ok {
		return rf(ctx, params, optFns...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *s3control.GetPublicAccessBlockInput, ...func(*s3control.Options)) *s3control.GetPublicAccessBlockOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*s3control.GetPublicAccessBlockOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *s3control.GetPublicAccessBlockInput, ...func(*s3control.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockControlClient_GetPublicAccessBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPublicAccessBlock'
type MockControlClient_GetPublicAccessBlock_Call struct {
	*mock.Call
}

// GetPublicAccessBlock is a helper method to define mock.On call
//   - ctx context.Context
//   - params *s3control.GetPublicAccessBlockInput
//   - optFns ...func(*s3control.Options)
func (_e *MockControlClient_Expecter) GetPublicAccessBlock(ctx interface{}, params interface{}, optFns ...interface{}) *MockControlClient_GetPublicAccessBlock_Call {
	return &MockControlClient_GetPublicAccessBlock_Call{Call: _e.mock.On("GetPublicAccessBlock",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *MockControlClient_GetPublicAccessBlock_Call) Run(run func(ctx context.Context, params *s3control.GetPublicAccessBlockInput, optFns ...func(*s3control.Options))) *MockControlClient_GetPublicAccessBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*s3control.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*s3control.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*s3control.GetPublicAccessBlockInput), variadicArgs...)
	})
	return _c
}

func (_c *MockControlClient_GetPublicAccessBlock_Call) Return(_a0 *s3control.GetPublicAccessBlockOutput, _a1 error) *MockControlClient_GetPublicAccessBlock_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockControlClient_GetPublicAccessBlock_Call) RunAndReturn(run func(context.Context, *s3control.GetPublicAccessBlockInput, ...func(*s3control.Options)) (*s3control.GetPublicAccessBlockOutput, error)) *MockControlClient_GetPublicAccessBlock_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockControlClient creates a new instance of MockControlClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockControlClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockControlClient {
	mock := &MockControlClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
