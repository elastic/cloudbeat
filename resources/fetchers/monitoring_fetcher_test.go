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

package fetchers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

func TestMonitoringFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name              string
		mocks             clientMocks
		wantErr           bool
		expectedResources int
	}{
		{
			name: "with resources",
			mocks: clientMocks{
				"Rules41_415": [2]mocks{
					{mock.Anything},
					{monitoring.Output{
						Items: []monitoring.Item{
							{},
							{},
						},
					}, nil},
				},
			},
			expectedResources: 1,
		},
		{
			name: "with error",
			mocks: clientMocks{
				"Rules41_415": [2]mocks{
					{mock.Anything},
					{monitoring.Output{}, fmt.Errorf("failed to run provider")},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetching.ResourceInfo, 100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			client := monitoring.MockClient{}
			for name, call := range tt.mocks {
				client.On(name, call[0]...).Return(call[1]...)
			}
			m := MonitoringFetcher{
				log:        logp.NewLogger("TestMonitoringFetcher_Fetch"),
				provider:   &client,
				cfg:        MonitoringFetcherConfig{},
				resourceCh: ch,
			}

			err := m.Fetch(ctx, fetching.CycleMetadata{})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			resources := testhelper.CollectResources(ch)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResources, len(resources))
		})
	}
}
