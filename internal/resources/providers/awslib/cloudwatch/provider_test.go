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

package cloudwatch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

var regions = []string{"us-east-1"}

func TestProvider_DescribeAlarms(t *testing.T) {
	tests := []struct {
		name    string
		mocks   clientMocks
		filters []string
		want    []types.MetricAlarm
		wantErr bool
	}{
		{
			name: "with results",
			mocks: clientMocks{
				"DescribeAlarms": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudwatch.DescribeAlarmsOutput{
						MetricAlarms: []types.MetricAlarm{
							{
								MetricName: aws.String("metric-name"),
							},
						},
					}, nil},
				},
			},
			filters: []string{"metric-name"},
			want: []types.MetricAlarm{
				{
					MetricName: aws.String("metric-name"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockClient{}
			for name, call := range tt.mocks {
				c.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				clients: createMockClients(c, regions),
			}
			got, err := p.DescribeAlarms(t.Context(), &regions[0], tt.filters)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func createMockClients(c Client, regions []string) map[string]Client {
	m := make(map[string]Client)
	for _, clientRegion := range regions {
		m[clientRegion] = c
	}

	return m
}
