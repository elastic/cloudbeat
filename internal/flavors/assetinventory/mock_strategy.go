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

package assetinventory

import (
	context "context"

	beat "github.com/elastic/beats/v7/libbeat/beat"

	inventory "github.com/elastic/cloudbeat/internal/inventory"

	mock "github.com/stretchr/testify/mock"
)

// MockStrategy is an autogenerated mock type for the Strategy type
type MockStrategy struct {
	mock.Mock
}

type MockStrategy_Expecter struct {
	mock *mock.Mock
}

func (_m *MockStrategy) EXPECT() *MockStrategy_Expecter {
	return &MockStrategy_Expecter{mock: &_m.Mock}
}

// NewAssetInventory provides a mock function with given fields: ctx, client
func (_m *MockStrategy) NewAssetInventory(ctx context.Context, client beat.Client) (inventory.AssetInventory, error) {
	ret := _m.Called(ctx, client)

	var r0 inventory.AssetInventory
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, beat.Client) (inventory.AssetInventory, error)); ok {
		return rf(ctx, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, beat.Client) inventory.AssetInventory); ok {
		r0 = rf(ctx, client)
	} else {
		r0 = ret.Get(0).(inventory.AssetInventory)
	}

	if rf, ok := ret.Get(1).(func(context.Context, beat.Client) error); ok {
		r1 = rf(ctx, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockStrategy_NewAssetInventory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewAssetInventory'
type MockStrategy_NewAssetInventory_Call struct {
	*mock.Call
}

// NewAssetInventory is a helper method to define mock.On call
//   - ctx context.Context
//   - client beat.Client
func (_e *MockStrategy_Expecter) NewAssetInventory(ctx interface{}, client interface{}) *MockStrategy_NewAssetInventory_Call {
	return &MockStrategy_NewAssetInventory_Call{Call: _e.mock.On("NewAssetInventory", ctx, client)}
}

func (_c *MockStrategy_NewAssetInventory_Call) Run(run func(ctx context.Context, client beat.Client)) *MockStrategy_NewAssetInventory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(beat.Client))
	})
	return _c
}

func (_c *MockStrategy_NewAssetInventory_Call) Return(_a0 inventory.AssetInventory, _a1 error) *MockStrategy_NewAssetInventory_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockStrategy_NewAssetInventory_Call) RunAndReturn(run func(context.Context, beat.Client) (inventory.AssetInventory, error)) *MockStrategy_NewAssetInventory_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockStrategy creates a new instance of MockStrategy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockStrategy(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStrategy {
	mock := &MockStrategy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}