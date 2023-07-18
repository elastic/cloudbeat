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

package awslib

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type apiResult struct {
	output *organizations.ListAccountsOutput
	err    error
}

type mockClient struct {
	resultMap map[string]apiResult
}

func (m *mockClient) ListAccounts(_ context.Context, input *organizations.ListAccountsInput, _ ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	token := strings.Dereference(input.NextToken)
	result, ok := m.resultMap[token]
	err := result.err
	if !ok {
		err = fmt.Errorf("could not find token: %s", token)
	}
	if err != nil {
		return nil, err
	}

	return result.output, nil
}

func Test_listAccounts(t *testing.T) {
	account1 := types.Account{
		Id:     aws.String("1"),
		Name:   aws.String("name"),
		Status: types.AccountStatusActive,
	}

	tests := []struct {
		name    string
		client  *mockClient
		want    []dataprovider.Identity
		wantErr string
	}{
		{
			name: "sanity check error",
			client: &mockClient{
				resultMap: map[string]apiResult{},
			},
			wantErr: "could not find token",
		},
		{
			name: "api error in first call",
			client: &mockClient{
				resultMap: map[string]apiResult{
					"": {
						err: errors.New("some error"),
					},
				},
			},
			wantErr: "some error",
		},
		{
			name: "api error in second call",
			client: &mockClient{
				resultMap: map[string]apiResult{
					"": {
						output: &organizations.ListAccountsOutput{
							Accounts:  []types.Account{account1},
							NextToken: aws.String("second"),
						},
						err: nil,
					},
					"second": {
						err: errors.New("some error"),
					},
				},
			},
			wantErr: "some error",
		},
		{
			name: "single account",
			client: &mockClient{
				resultMap: map[string]apiResult{
					"": {
						output: &organizations.ListAccountsOutput{
							Accounts:  []types.Account{account1},
							NextToken: nil,
						},
						err: nil,
					},
				},
			},
			want: []dataprovider.Identity{
				{
					Account:      "1",
					AccountAlias: "name",
				},
			},
		},
		{
			name: "many accounts",
			client: &mockClient{
				resultMap: map[string]apiResult{
					"": {
						output: &organizations.ListAccountsOutput{
							Accounts:  []types.Account{account1},
							NextToken: aws.String("second"),
						},
					},
					"second": {
						output: &organizations.ListAccountsOutput{
							Accounts: []types.Account{
								{
									Id:     aws.String("123"),
									Status: types.AccountStatusActive,
								},
								{
									Id:     aws.String("456"),
									Name:   aws.String("suspended"),
									Status: types.AccountStatusSuspended,
								},
								{
									Id:     aws.String("567"),
									Status: types.AccountStatusPendingClosure,
								},
							},
							NextToken: aws.String("third"),
						},
					},
					"third": {
						output: &organizations.ListAccountsOutput{
							Accounts: []types.Account{
								{
									Id:     aws.String("1000"),
									Name:   aws.String("some name"),
									Status: types.AccountStatusActive,
								},
								{
									Id:     nil, // shouldn't really happen
									Status: types.AccountStatusActive,
								},
							},
							NextToken: nil,
						},
						err: nil,
					},
				},
			},
			want: []dataprovider.Identity{
				{
					Account:      "1",
					AccountAlias: "name",
				},
				{
					Account: "123",
				},
				{
					Account:      "1000",
					AccountAlias: "some name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listAccounts(context.Background(), tt.client)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
