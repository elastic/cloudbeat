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

package securityhub

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

func TestProvider_Describe(t *testing.T) {
	tests := []struct {
		name    string
		want    SecurityHub
		wantErr bool
		mocks   clientMocks
	}{
		{
			name: "enabled",
			want: SecurityHub{
				Enabled:           true,
				DescribeHubOutput: &securityhub.DescribeHubOutput{},
			},
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{&securityhub.DescribeHubOutput{}, nil},
				},
			},
		},
		{
			name: "disabled",
			want: SecurityHub{
				Enabled: false,
			},
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, fmt.Errorf("is not subscribed to AWS Security Hub")},
				},
			},
		},
		{
			name:    "with error",
			wantErr: true,
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, fmt.Errorf("error")},
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
				log:    logp.NewLogger("TestProvider_Describe"),
				client: c,
			}
			got, err := p.Describe(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
