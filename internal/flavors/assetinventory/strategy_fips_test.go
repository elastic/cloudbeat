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
//go:build requirefips

package assetinventory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestStrategyPicks_FIPSUnsupportedProviders(t *testing.T) {
	for provider, name := range map[string]string{
		config.ProviderAzure: "azure",
		config.ProviderGCP:   "gcp",
	} {
		t.Run(provider, func(t *testing.T) {
			cfg := &config.Config{AssetInventoryProvider: provider}
			obj, err := GetStrategy(testhelper.NewLogger(t), cfg, nil).NewAssetInventory(t.Context(), nil)
			assert.Equal(t, inventory.AssetInventory{}, obj)
			require.ErrorIs(t, err, errors.ErrUnsupported)
			assert.ErrorContains(t, err, name+" is not supported in the FIPS distribution")
		})
	}
}
