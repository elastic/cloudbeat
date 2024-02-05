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

package cloud

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

var (
	accountName    = "accountName"
	accountId      = "accountId"
	awsProvider    = "aws"
	gcpProvider    = "gcp"
	gcpProjectName = "projectName"
	gcpProjectId   = "projectId"
	someRegion     = "eu-west-1"
)

func TestDataProvider_EnrichEvent(t *testing.T) {
	tests := []struct {
		name           string
		resMetadata    fetching.ResourceMetadata
		identity       Identity
		expectedFields map[string]string
	}{
		{
			name: "no replacement",
			resMetadata: fetching.ResourceMetadata{
				Region: someRegion,
			},
			identity: Identity{
				Account:      accountId,
				AccountAlias: accountName,
				Provider:     awsProvider,
			},
			expectedFields: map[string]string{
				cloudAccountIdField:   accountId,
				cloudAccountNameField: accountName,
				cloudProviderField:    awsProvider,
				cloudRegionField:      someRegion,
			},
		},
		{
			name: "replace alias",
			resMetadata: fetching.ResourceMetadata{
				Region: someRegion,
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:   "",
					AccountName: "some alias",
				},
			},
			identity: Identity{
				Account:      accountId,
				AccountAlias: accountName,
				Provider:     awsProvider,
			},
			expectedFields: map[string]string{
				cloudAccountIdField:   accountId,
				cloudAccountNameField: "some alias",
				cloudProviderField:    awsProvider,
				cloudRegionField:      someRegion,
			},
		},
		{
			name: "replace both",
			resMetadata: fetching.ResourceMetadata{
				Region: someRegion,
				CloudAccountMetadata: fetching.CloudAccountMetadata{
					AccountId:   "12345654321",
					AccountName: "some alias",
				},
			},
			identity: Identity{
				Account:      accountId,
				AccountAlias: accountName,
				Provider:     awsProvider,
			},
			expectedFields: map[string]string{
				cloudAccountIdField:   "12345654321",
				cloudAccountNameField: "some alias",
				cloudProviderField:    awsProvider,
				cloudRegionField:      someRegion,
			},
		},
		{
			name: "enrich a gcp event",
			resMetadata: fetching.ResourceMetadata{
				Region: someRegion,
			},
			identity: Identity{
				Provider:     gcpProvider,
				Account:      gcpProjectId,
				AccountAlias: gcpProjectName,
			},
			expectedFields: map[string]string{
				cloudProviderField:    gcpProvider,
				cloudRegionField:      someRegion,
				cloudAccountIdField:   gcpProjectId,
				cloudAccountNameField: gcpProjectName,
			},
		},
		{
			name: "missing fields",
			resMetadata: fetching.ResourceMetadata{
				Region: someRegion,
			},
			identity: Identity{
				Provider: awsProvider,
			},
			expectedFields: map[string]string{
				cloudProviderField:  awsProvider,
				cloudAccountIdField: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDataProvider(WithAccount(tt.identity))
			e := &beat.Event{
				Fields: mapstr.M{},
			}

			err := p.EnrichEvent(e, tt.resMetadata)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedFields {
				assertField(t, e.Fields, key, expectedValue)
			}
		})
	}
}

func assertField(t *testing.T, fields mapstr.M, key string, expectedValue string) {
	t.Helper()

	got, err := fields.GetValue(key)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, got)
}
