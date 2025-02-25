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
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

func TestProvider_Describe(t *testing.T) {
	accountId := "dummy-id"
	tests := []struct {
		name    string
		want    []SecurityHub
		wantErr bool
		mocks   clientMocks
		regions []string
	}{
		{
			name: "enabled",
			want: []SecurityHub{{
				Enabled:           true,
				DescribeHubOutput: &securityhub.DescribeHubOutput{},
				Region:            awslib.DefaultRegion,
				AccountId:         accountId,
			}},
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{&securityhub.DescribeHubOutput{}, nil},
				},
			},
			regions: []string{awslib.DefaultRegion},
		},
		{
			name: "disabled",
			want: []SecurityHub{{
				Enabled:   false,
				Region:    awslib.DefaultRegion,
				AccountId: accountId,
			}},
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, errors.New("is not subscribed to AWS Security Hub")},
				},
			},
			regions: []string{awslib.DefaultRegion},
		},
		{
			name: "multi region",
			want: []SecurityHub{{
				Enabled:           true,
				DescribeHubOutput: &securityhub.DescribeHubOutput{},
				Region:            awslib.DefaultRegion,
				AccountId:         accountId,
			}, {
				Enabled:           true,
				DescribeHubOutput: &securityhub.DescribeHubOutput{},
				Region:            "eu-west-1",
				AccountId:         accountId,
			}},
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{&securityhub.DescribeHubOutput{}, nil},
				},
			},
			regions: []string{awslib.DefaultRegion, "eu-west-1"},
		},
		{
			name:    "with error",
			wantErr: true,
			mocks: clientMocks{
				"DescribeHub": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, errors.New("error")},
				},
			},
			regions: []string{awslib.DefaultRegion},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockClient{}
			for name, call := range tt.mocks {
				c.On(name, call[0]...).Return(call[1]...)
			}
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = c
			}
			p := &Provider{
				accountId: accountId,
				clients:   clients,
			}
			got, err := p.Describe(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
			var names []string
			for _, r := range tt.regions {
				names = append(names, fmt.Sprintf("securityhub-%s-%s", r, p.accountId))
			}
			for _, s := range got {
				assert.Contains(t, names, s.GetResourceName())
			}
		})
	}
}
