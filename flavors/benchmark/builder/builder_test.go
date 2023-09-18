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
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/uniqueness"
)

func TestBase_Success(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		benchType interface{}
	}{
		{
			name:      "by default create base benchmark",
			benchType: &basebenchmark{},
		},
		{
			name: "with opts create base benchmark",
			opts: []Option{
				WithIdProvider(dataprovider.NewMockIdProvider(t)),
				WithManagerTimeout(0),
				WithBenchmarkDataProvider(dataprovider.NewMockCommonDataProvider(t)),
			},
			benchType: &basebenchmark{},
		},
		{
			name: "with leader elector create k8s benchmark",
			opts: []Option{
				WithK8sLeaderElector(uniqueness.NewMockManager(t)),
			},
			benchType: &k8sbenchmark{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			path, err := filepath.Abs("../../../bundle.tar.gz")
			assert.NoError(t, err)

			benchmark, err := New(tt.opts...).Build(context.Background(), log, &config.Config{
				BundlePath: path,
			}, nil, nil)
			assert.NoError(t, err)
			assert.IsType(t, tt.benchType, benchmark)
		})
	}
}
