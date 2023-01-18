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

package cloudtrail

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mocks                 []any
	cloudtrailClientMocks map[string][2]mocks
)

func TestProvider_DescribeCloudTrails(t *testing.T) {
	tests := []struct {
		name                           string
		want                           []TrailInfo
		wantErr                        bool
		cloudtrailClientMockReturnVals cloudtrailClientMocks
	}{
		{
			cloudtrailClientMockReturnVals: cloudtrailClientMocks{
				"DescribeTrails": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.DescribeTrailsOutput{
						TrailList: []types.Trail{
							{
								Name: aws.String("trail"),
							},
						},
					}, nil},
				},
				"GetTrailStatus": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.GetTrailStatusOutput{}, nil},
				},
				"GetEventSelectors": [2]mocks{
					{mock.Anything, mock.Anything},
					{&cloudtrail.GetEventSelectorsOutput{}, nil},
				},
			},
			want: []TrailInfo{
				{
					trail: types.Trail{
						Name: aws.String("trail"),
					},
					status:        &cloudtrail.GetTrailStatusOutput{},
					eventSelector: &cloudtrail.GetEventSelectorsOutput{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{}
			for name, call := range tt.cloudtrailClientMockReturnVals {
				mock.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				log:    logp.NewLogger("TestProvider_DescribeCloudTrails"),
				client: mock,
			}
			got, err := p.DescribeCloudTrails(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			for i, g := range got {
				assert.Equal(t, tt.want[i], g)
			}
		})
	}
}
