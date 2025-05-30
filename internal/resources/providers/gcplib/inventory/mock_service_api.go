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

package inventory

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockServiceAPI is an autogenerated mock type for the ServiceAPI type
type MockServiceAPI struct {
	mock.Mock
}

type MockServiceAPI_Expecter struct {
	mock *mock.Mock
}

func (_m *MockServiceAPI) EXPECT() *MockServiceAPI_Expecter {
	return &MockServiceAPI_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with no fields
func (_m *MockServiceAPI) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockServiceAPI_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockServiceAPI_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockServiceAPI_Expecter) Close() *MockServiceAPI_Close_Call {
	return &MockServiceAPI_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockServiceAPI_Close_Call) Run(run func()) *MockServiceAPI_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockServiceAPI_Close_Call) Return(_a0 error) *MockServiceAPI_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockServiceAPI_Close_Call) RunAndReturn(run func() error) *MockServiceAPI_Close_Call {
	_c.Call.Return(run)
	return _c
}

// ListAllAssetTypesByName provides a mock function with given fields: ctx, assets
func (_m *MockServiceAPI) ListAllAssetTypesByName(ctx context.Context, assets []string) ([]*ExtendedGcpAsset, error) {
	ret := _m.Called(ctx, assets)

	if len(ret) == 0 {
		panic("no return value specified for ListAllAssetTypesByName")
	}

	var r0 []*ExtendedGcpAsset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*ExtendedGcpAsset, error)); ok {
		return rf(ctx, assets)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*ExtendedGcpAsset); ok {
		r0 = rf(ctx, assets)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ExtendedGcpAsset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, assets)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServiceAPI_ListAllAssetTypesByName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAllAssetTypesByName'
type MockServiceAPI_ListAllAssetTypesByName_Call struct {
	*mock.Call
}

// ListAllAssetTypesByName is a helper method to define mock.On call
//   - ctx context.Context
//   - assets []string
func (_e *MockServiceAPI_Expecter) ListAllAssetTypesByName(ctx interface{}, assets interface{}) *MockServiceAPI_ListAllAssetTypesByName_Call {
	return &MockServiceAPI_ListAllAssetTypesByName_Call{Call: _e.mock.On("ListAllAssetTypesByName", ctx, assets)}
}

func (_c *MockServiceAPI_ListAllAssetTypesByName_Call) Run(run func(ctx context.Context, assets []string)) *MockServiceAPI_ListAllAssetTypesByName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *MockServiceAPI_ListAllAssetTypesByName_Call) Return(_a0 []*ExtendedGcpAsset, _a1 error) *MockServiceAPI_ListAllAssetTypesByName_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServiceAPI_ListAllAssetTypesByName_Call) RunAndReturn(run func(context.Context, []string) ([]*ExtendedGcpAsset, error)) *MockServiceAPI_ListAllAssetTypesByName_Call {
	_c.Call.Return(run)
	return _c
}

// ListLoggingAssets provides a mock function with given fields: ctx
func (_m *MockServiceAPI) ListLoggingAssets(ctx context.Context) ([]*LoggingAsset, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListLoggingAssets")
	}

	var r0 []*LoggingAsset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*LoggingAsset, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*LoggingAsset); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*LoggingAsset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServiceAPI_ListLoggingAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListLoggingAssets'
type MockServiceAPI_ListLoggingAssets_Call struct {
	*mock.Call
}

// ListLoggingAssets is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockServiceAPI_Expecter) ListLoggingAssets(ctx interface{}) *MockServiceAPI_ListLoggingAssets_Call {
	return &MockServiceAPI_ListLoggingAssets_Call{Call: _e.mock.On("ListLoggingAssets", ctx)}
}

func (_c *MockServiceAPI_ListLoggingAssets_Call) Run(run func(ctx context.Context)) *MockServiceAPI_ListLoggingAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockServiceAPI_ListLoggingAssets_Call) Return(_a0 []*LoggingAsset, _a1 error) *MockServiceAPI_ListLoggingAssets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServiceAPI_ListLoggingAssets_Call) RunAndReturn(run func(context.Context) ([]*LoggingAsset, error)) *MockServiceAPI_ListLoggingAssets_Call {
	_c.Call.Return(run)
	return _c
}

// ListMonitoringAssets provides a mock function with given fields: ctx, monitoringAssetTypes
func (_m *MockServiceAPI) ListMonitoringAssets(ctx context.Context, monitoringAssetTypes map[string][]string) ([]*MonitoringAsset, error) {
	ret := _m.Called(ctx, monitoringAssetTypes)

	if len(ret) == 0 {
		panic("no return value specified for ListMonitoringAssets")
	}

	var r0 []*MonitoringAsset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, map[string][]string) ([]*MonitoringAsset, error)); ok {
		return rf(ctx, monitoringAssetTypes)
	}
	if rf, ok := ret.Get(0).(func(context.Context, map[string][]string) []*MonitoringAsset); ok {
		r0 = rf(ctx, monitoringAssetTypes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*MonitoringAsset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, map[string][]string) error); ok {
		r1 = rf(ctx, monitoringAssetTypes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServiceAPI_ListMonitoringAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListMonitoringAssets'
type MockServiceAPI_ListMonitoringAssets_Call struct {
	*mock.Call
}

// ListMonitoringAssets is a helper method to define mock.On call
//   - ctx context.Context
//   - monitoringAssetTypes map[string][]string
func (_e *MockServiceAPI_Expecter) ListMonitoringAssets(ctx interface{}, monitoringAssetTypes interface{}) *MockServiceAPI_ListMonitoringAssets_Call {
	return &MockServiceAPI_ListMonitoringAssets_Call{Call: _e.mock.On("ListMonitoringAssets", ctx, monitoringAssetTypes)}
}

func (_c *MockServiceAPI_ListMonitoringAssets_Call) Run(run func(ctx context.Context, monitoringAssetTypes map[string][]string)) *MockServiceAPI_ListMonitoringAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(map[string][]string))
	})
	return _c
}

func (_c *MockServiceAPI_ListMonitoringAssets_Call) Return(_a0 []*MonitoringAsset, _a1 error) *MockServiceAPI_ListMonitoringAssets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServiceAPI_ListMonitoringAssets_Call) RunAndReturn(run func(context.Context, map[string][]string) ([]*MonitoringAsset, error)) *MockServiceAPI_ListMonitoringAssets_Call {
	_c.Call.Return(run)
	return _c
}

// ListProjectsAncestorsPolicies provides a mock function with given fields: ctx
func (_m *MockServiceAPI) ListProjectsAncestorsPolicies(ctx context.Context) ([]*ProjectPoliciesAsset, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListProjectsAncestorsPolicies")
	}

	var r0 []*ProjectPoliciesAsset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*ProjectPoliciesAsset, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*ProjectPoliciesAsset); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ProjectPoliciesAsset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServiceAPI_ListProjectsAncestorsPolicies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListProjectsAncestorsPolicies'
type MockServiceAPI_ListProjectsAncestorsPolicies_Call struct {
	*mock.Call
}

// ListProjectsAncestorsPolicies is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockServiceAPI_Expecter) ListProjectsAncestorsPolicies(ctx interface{}) *MockServiceAPI_ListProjectsAncestorsPolicies_Call {
	return &MockServiceAPI_ListProjectsAncestorsPolicies_Call{Call: _e.mock.On("ListProjectsAncestorsPolicies", ctx)}
}

func (_c *MockServiceAPI_ListProjectsAncestorsPolicies_Call) Run(run func(ctx context.Context)) *MockServiceAPI_ListProjectsAncestorsPolicies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockServiceAPI_ListProjectsAncestorsPolicies_Call) Return(_a0 []*ProjectPoliciesAsset, _a1 error) *MockServiceAPI_ListProjectsAncestorsPolicies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServiceAPI_ListProjectsAncestorsPolicies_Call) RunAndReturn(run func(context.Context) ([]*ProjectPoliciesAsset, error)) *MockServiceAPI_ListProjectsAncestorsPolicies_Call {
	_c.Call.Return(run)
	return _c
}

// ListServiceUsageAssets provides a mock function with given fields: ctx
func (_m *MockServiceAPI) ListServiceUsageAssets(ctx context.Context) ([]*ServiceUsageAsset, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListServiceUsageAssets")
	}

	var r0 []*ServiceUsageAsset
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*ServiceUsageAsset, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*ServiceUsageAsset); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ServiceUsageAsset)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockServiceAPI_ListServiceUsageAssets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListServiceUsageAssets'
type MockServiceAPI_ListServiceUsageAssets_Call struct {
	*mock.Call
}

// ListServiceUsageAssets is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockServiceAPI_Expecter) ListServiceUsageAssets(ctx interface{}) *MockServiceAPI_ListServiceUsageAssets_Call {
	return &MockServiceAPI_ListServiceUsageAssets_Call{Call: _e.mock.On("ListServiceUsageAssets", ctx)}
}

func (_c *MockServiceAPI_ListServiceUsageAssets_Call) Run(run func(ctx context.Context)) *MockServiceAPI_ListServiceUsageAssets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockServiceAPI_ListServiceUsageAssets_Call) Return(_a0 []*ServiceUsageAsset, _a1 error) *MockServiceAPI_ListServiceUsageAssets_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockServiceAPI_ListServiceUsageAssets_Call) RunAndReturn(run func(context.Context) ([]*ServiceUsageAsset, error)) *MockServiceAPI_ListServiceUsageAssets_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockServiceAPI creates a new instance of MockServiceAPI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockServiceAPI(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockServiceAPI {
	mock := &MockServiceAPI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
