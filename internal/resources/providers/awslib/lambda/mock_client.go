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

// Code generated by mockery v2.37.1. DO NOT EDIT.

package lambda

import (
	context "context"

	servicelambda "github.com/aws/aws-sdk-go-v2/service/lambda"
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

// ListAliases provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockClient) ListAliases(_a0 context.Context, _a1 *servicelambda.ListAliasesInput, _a2 ...func(*servicelambda.Options)) (*servicelambda.ListAliasesOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *servicelambda.ListAliasesOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListAliasesInput, ...func(*servicelambda.Options)) (*servicelambda.ListAliasesOutput, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListAliasesInput, ...func(*servicelambda.Options)) *servicelambda.ListAliasesOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*servicelambda.ListAliasesOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *servicelambda.ListAliasesInput, ...func(*servicelambda.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ListAliases_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAliases'
type MockClient_ListAliases_Call struct {
	*mock.Call
}

// ListAliases is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *servicelambda.ListAliasesInput
//   - _a2 ...func(*servicelambda.Options)
func (_e *MockClient_Expecter) ListAliases(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *MockClient_ListAliases_Call {
	return &MockClient_ListAliases_Call{Call: _e.mock.On("ListAliases",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *MockClient_ListAliases_Call) Run(run func(_a0 context.Context, _a1 *servicelambda.ListAliasesInput, _a2 ...func(*servicelambda.Options))) *MockClient_ListAliases_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*servicelambda.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*servicelambda.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*servicelambda.ListAliasesInput), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_ListAliases_Call) Return(_a0 *servicelambda.ListAliasesOutput, _a1 error) *MockClient_ListAliases_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ListAliases_Call) RunAndReturn(run func(context.Context, *servicelambda.ListAliasesInput, ...func(*servicelambda.Options)) (*servicelambda.ListAliasesOutput, error)) *MockClient_ListAliases_Call {
	_c.Call.Return(run)
	return _c
}

// ListEventSourceMappings provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockClient) ListEventSourceMappings(_a0 context.Context, _a1 *servicelambda.ListEventSourceMappingsInput, _a2 ...func(*servicelambda.Options)) (*servicelambda.ListEventSourceMappingsOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *servicelambda.ListEventSourceMappingsOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListEventSourceMappingsInput, ...func(*servicelambda.Options)) (*servicelambda.ListEventSourceMappingsOutput, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListEventSourceMappingsInput, ...func(*servicelambda.Options)) *servicelambda.ListEventSourceMappingsOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*servicelambda.ListEventSourceMappingsOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *servicelambda.ListEventSourceMappingsInput, ...func(*servicelambda.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ListEventSourceMappings_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListEventSourceMappings'
type MockClient_ListEventSourceMappings_Call struct {
	*mock.Call
}

// ListEventSourceMappings is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *servicelambda.ListEventSourceMappingsInput
//   - _a2 ...func(*servicelambda.Options)
func (_e *MockClient_Expecter) ListEventSourceMappings(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *MockClient_ListEventSourceMappings_Call {
	return &MockClient_ListEventSourceMappings_Call{Call: _e.mock.On("ListEventSourceMappings",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *MockClient_ListEventSourceMappings_Call) Run(run func(_a0 context.Context, _a1 *servicelambda.ListEventSourceMappingsInput, _a2 ...func(*servicelambda.Options))) *MockClient_ListEventSourceMappings_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*servicelambda.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*servicelambda.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*servicelambda.ListEventSourceMappingsInput), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_ListEventSourceMappings_Call) Return(_a0 *servicelambda.ListEventSourceMappingsOutput, _a1 error) *MockClient_ListEventSourceMappings_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ListEventSourceMappings_Call) RunAndReturn(run func(context.Context, *servicelambda.ListEventSourceMappingsInput, ...func(*servicelambda.Options)) (*servicelambda.ListEventSourceMappingsOutput, error)) *MockClient_ListEventSourceMappings_Call {
	_c.Call.Return(run)
	return _c
}

// ListFunctions provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockClient) ListFunctions(_a0 context.Context, _a1 *servicelambda.ListFunctionsInput, _a2 ...func(*servicelambda.Options)) (*servicelambda.ListFunctionsOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *servicelambda.ListFunctionsOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListFunctionsInput, ...func(*servicelambda.Options)) (*servicelambda.ListFunctionsOutput, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListFunctionsInput, ...func(*servicelambda.Options)) *servicelambda.ListFunctionsOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*servicelambda.ListFunctionsOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *servicelambda.ListFunctionsInput, ...func(*servicelambda.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ListFunctions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListFunctions'
type MockClient_ListFunctions_Call struct {
	*mock.Call
}

// ListFunctions is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *servicelambda.ListFunctionsInput
//   - _a2 ...func(*servicelambda.Options)
func (_e *MockClient_Expecter) ListFunctions(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *MockClient_ListFunctions_Call {
	return &MockClient_ListFunctions_Call{Call: _e.mock.On("ListFunctions",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *MockClient_ListFunctions_Call) Run(run func(_a0 context.Context, _a1 *servicelambda.ListFunctionsInput, _a2 ...func(*servicelambda.Options))) *MockClient_ListFunctions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*servicelambda.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*servicelambda.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*servicelambda.ListFunctionsInput), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_ListFunctions_Call) Return(_a0 *servicelambda.ListFunctionsOutput, _a1 error) *MockClient_ListFunctions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ListFunctions_Call) RunAndReturn(run func(context.Context, *servicelambda.ListFunctionsInput, ...func(*servicelambda.Options)) (*servicelambda.ListFunctionsOutput, error)) *MockClient_ListFunctions_Call {
	_c.Call.Return(run)
	return _c
}

// ListLayers provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockClient) ListLayers(_a0 context.Context, _a1 *servicelambda.ListLayersInput, _a2 ...func(*servicelambda.Options)) (*servicelambda.ListLayersOutput, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0, _a1)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *servicelambda.ListLayersOutput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListLayersInput, ...func(*servicelambda.Options)) (*servicelambda.ListLayersOutput, error)); ok {
		return rf(_a0, _a1, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *servicelambda.ListLayersInput, ...func(*servicelambda.Options)) *servicelambda.ListLayersOutput); ok {
		r0 = rf(_a0, _a1, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*servicelambda.ListLayersOutput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *servicelambda.ListLayersInput, ...func(*servicelambda.Options)) error); ok {
		r1 = rf(_a0, _a1, _a2...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ListLayers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListLayers'
type MockClient_ListLayers_Call struct {
	*mock.Call
}

// ListLayers is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *servicelambda.ListLayersInput
//   - _a2 ...func(*servicelambda.Options)
func (_e *MockClient_Expecter) ListLayers(_a0 interface{}, _a1 interface{}, _a2 ...interface{}) *MockClient_ListLayers_Call {
	return &MockClient_ListLayers_Call{Call: _e.mock.On("ListLayers",
		append([]interface{}{_a0, _a1}, _a2...)...)}
}

func (_c *MockClient_ListLayers_Call) Run(run func(_a0 context.Context, _a1 *servicelambda.ListLayersInput, _a2 ...func(*servicelambda.Options))) *MockClient_ListLayers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*servicelambda.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*servicelambda.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*servicelambda.ListLayersInput), variadicArgs...)
	})
	return _c
}

func (_c *MockClient_ListLayers_Call) Return(_a0 *servicelambda.ListLayersOutput, _a1 error) *MockClient_ListLayers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ListLayers_Call) RunAndReturn(run func(context.Context, *servicelambda.ListLayersInput, ...func(*servicelambda.Options)) (*servicelambda.ListLayersOutput, error)) *MockClient_ListLayers_Call {
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