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
//go:build !requirefips

package assetinventory

import (
	"testing"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestStrategyPicks_GCP(t *testing.T) {
	testhelper.SkipLong(t)

	testStrategyPicks(t, []strategyPicksCase{
		{
			name: "expected error: GCP missing account type",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderGCP,
			},
			env:         nil,
			expectedErr: "invalid gcp account type",
		},
		{
			name: "expected success: GCP",
			cfg: &config.Config{
				AssetInventoryProvider: config.ProviderGCP,
				CloudConfig: config.CloudConfig{
					Gcp: config.GcpConfig{
						AccountType:    config.SingleAccount,
						ProjectId:      "nonexistent",
						OrganizationId: "nonexistent",
						GcpClientOpt: config.GcpClientOpt{
							CredentialsJSON: "{\"type\": \"service_account\"}",
						},
					},
				},
			},
			env:         nil,
			expectedErr: "could not parse key",
		},
	})
}
