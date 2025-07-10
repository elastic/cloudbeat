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

package builder

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/manager"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/uniqueness"
)

func TestBase_Build_Success(t *testing.T) {
	testhelper.SkipLong(t)

	tests := []struct {
		name      string
		opts      []Option
		benchType any
	}{
		{
			name:      "by default create base benchmark",
			benchType: &basebenchmark{}, //nolint:exhaustruct
		},
		{
			name: "with opts create base benchmark",
			opts: []Option{
				WithIdProvider(dataprovider.NewMockIdProvider(t)),
				WithManagerTimeout(time.Minute),
				WithBenchmarkDataProvider(dataprovider.NewMockCommonDataProvider(t)),
			},
			benchType: &basebenchmark{}, //nolint:exhaustruct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			path, err := filepath.Abs("../../../../bundle.tar.gz")
			require.NoError(t, err)

			resourceCh := make(chan fetching.ResourceInfo)
			reg := registry.NewMockRegistry(t)
			mpf := manager.NewMockPreflight(t)
			mpf.EXPECT().Prepare(mock.Anything, mock.Anything).Return(nil)
			benchmark, err := New(tt.opts...).Build(t.Context(), log, &config.Config{
				BundlePath: path,
				Period:     time.Minute,
			}, resourceCh, reg, mpf)
			require.NoError(t, err)
			assert.IsType(t, tt.benchType, benchmark)

			reg.EXPECT().Keys().Return([]string{}).Twice()
			reg.EXPECT().Update().Return().Once()
			_, err = benchmark.Run(t.Context())
			time.Sleep(100 * time.Millisecond)
			require.NoError(t, err)
		})
	}
}

func TestBase_BuildK8s_Success(t *testing.T) {
	testhelper.SkipLong(t)

	tests := []struct {
		name      string
		opts      []Option
		benchType any
	}{
		{
			name:      "by default create k8s benchmark",
			benchType: &k8sbenchmark{}, //nolint:exhaustruct

		}, {
			name: "with opts create k8s benchmark",
			opts: []Option{
				WithIdProvider(dataprovider.NewMockIdProvider(t)),
				WithManagerTimeout(time.Minute),
				WithBenchmarkDataProvider(dataprovider.NewMockCommonDataProvider(t)),
			},
			benchType: &k8sbenchmark{}, //nolint:exhaustruct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			path, err := filepath.Abs("../../../../bundle.tar.gz")
			require.NoError(t, err)

			resourceCh := make(chan fetching.ResourceInfo)
			reg := registry.NewMockRegistry(t)
			le := uniqueness.NewMockManager(t)
			benchmark, err := New(tt.opts...).BuildK8s(t.Context(), log, &config.Config{
				BundlePath: path,
				Period:     time.Minute,
			}, resourceCh, reg, le)
			require.NoError(t, err)
			assert.IsType(t, tt.benchType, benchmark)

			reg.EXPECT().Keys().Return([]string{}).Twice()
			reg.EXPECT().Update().Return().Once()
			le.EXPECT().Run(mock.Anything).Return(nil).Once()
			_, err = benchmark.Run(t.Context())
			time.Sleep(100 * time.Millisecond)
			require.NoError(t, err)
		})
	}
}
