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

// MockAssetFetcher is an autogenerated mock type for the AssetFetcher type
type MockAssetFetcher struct {
	mock.Mock
}

type MockAssetFetcher_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAssetFetcher) EXPECT() *MockAssetFetcher_Expecter {
	return &MockAssetFetcher_Expecter{mock: &_m.Mock}
}

// Fetch provides a mock function with given fields: ctx, assetChannel
func (_m *MockAssetFetcher) Fetch(ctx context.Context, assetChannel chan<- AssetEvent) {
	_m.Called(ctx, assetChannel)
}

// MockAssetFetcher_Fetch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fetch'
type MockAssetFetcher_Fetch_Call struct {
	*mock.Call
}

// Fetch is a helper method to define mock.On call
//   - ctx context.Context
//   - assetChannel chan<- AssetEvent
func (_e *MockAssetFetcher_Expecter) Fetch(ctx interface{}, assetChannel interface{}) *MockAssetFetcher_Fetch_Call {
	return &MockAssetFetcher_Fetch_Call{Call: _e.mock.On("Fetch", ctx, assetChannel)}
}

func (_c *MockAssetFetcher_Fetch_Call) Run(run func(ctx context.Context, assetChannel chan<- AssetEvent)) *MockAssetFetcher_Fetch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(chan<- AssetEvent))
	})
	return _c
}

func (_c *MockAssetFetcher_Fetch_Call) Return() *MockAssetFetcher_Fetch_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockAssetFetcher_Fetch_Call) RunAndReturn(run func(context.Context, chan<- AssetEvent)) *MockAssetFetcher_Fetch_Call {
	_c.Run(run)
	return _c
}

// NewMockAssetFetcher creates a new instance of MockAssetFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAssetFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAssetFetcher {
	mock := &MockAssetFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
