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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type listAccountsResult struct {
	output *organizations.ListAccountsOutput
	err    error
}

type listParentsResult struct {
	output *organizations.ListParentsOutput
	err    error
}

func Test_listAccounts(t *testing.T) {
	account1 := types.Account{
		Id:     aws.String("1"),
		Name:   aws.String("name"),
		Status: types.AccountStatusActive,
	}
	ou1Result := listParentsResult{
		output: &organizations.ListParentsOutput{
			Parents: []types.Parent{{
				Id:   aws.String("ou-1"),
				Type: types.ParentTypeOrganizationalUnit,
			}},
		},
	}

	tests := []struct {
		name                      string
		listAccountsResults       map[string]listAccountsResult
		listParentsResults        map[string]listParentsResult
		describeOrganizationError error
		want                      []cloud.Identity
		wantErr                   string
	}{
		{
			name:    "sanity check error",
			wantErr: "could not find token",
		},
		{
			name: "api error in first call",
			listAccountsResults: map[string]listAccountsResult{
				"": {
					err: errors.New("some error"),
				},
			},
			wantErr: "some error",
		},
		{
			name: "api error in second call",
			listAccountsResults: map[string]listAccountsResult{
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
			listParentsResults: map[string]listParentsResult{},
			wantErr:            "some error",
		},
		{
			name: "single account",
			listAccountsResults: map[string]listAccountsResult{
				"": {
					output: &organizations.ListAccountsOutput{
						Accounts:  []types.Account{account1},
						NextToken: nil,
					},
					err: nil,
				},
			},
			listParentsResults: map[string]listParentsResult{
				"1": ou1Result,
			},
			want: []cloud.Identity{
				{
					Provider:         "aws",
					Account:          "1",
					AccountAlias:     "name",
					OrganizationId:   "ou-1",
					OrganizationName: "ou-1-name",
				},
			},
		},
		{
			name: "many accounts",
			listAccountsResults: map[string]listAccountsResult{
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
			want: []cloud.Identity{
				{
					Provider:         "aws",
					Account:          "1",
					AccountAlias:     "name",
					OrganizationId:   "ou-1",
					OrganizationName: "ou-1-name",
				},
				{
					Provider:         "aws",
					Account:          "123",
					AccountAlias:     "",
					OrganizationId:   "ou-1",
					OrganizationName: "ou-1-name",
				},
				{
					Provider:         "aws",
					Account:          "1000",
					AccountAlias:     "some name",
					OrganizationId:   "",
					OrganizationName: "",
				},
			},
			listParentsResults: map[string]listParentsResult{
				"1":    ou1Result,
				"123":  ou1Result,
				"1000": {err: errors.New("some-ignored-error")},
			},
		},
		{
			name: "ignore describe organization error",
			listAccountsResults: map[string]listAccountsResult{
				"": {
					output: &organizations.ListAccountsOutput{
						Accounts:  []types.Account{account1},
						NextToken: nil,
					},
					err: nil,
				},
			},
			listParentsResults: map[string]listParentsResult{
				"1": ou1Result,
			},
			want: []cloud.Identity{
				{
					Provider:         "aws",
					Account:          "1",
					AccountAlias:     "name",
					OrganizationId:   "ou-1",
					OrganizationName: "",
				},
			},
			describeOrganizationError: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockFromResultMap(tt.listAccountsResults, tt.listParentsResults, tt.describeOrganizationError)
			defer m.AssertExpectations(t)

			got, err := listAccounts(context.Background(), testhelper.NewLogger(t), m)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func mockFromResultMap(
	listAccountsResults map[string]listAccountsResult,
	listParentsResults map[string]listParentsResult,
	describeOrganizationalUnitError error,
) *mockOrganizationsAPI {
	m := mockOrganizationsAPI{}
	m.EXPECT().ListAccounts(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, input *organizations.ListAccountsInput, _ ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
			token := pointers.Deref(input.NextToken)
			result, ok := listAccountsResults[token]
			err := result.err
			if !ok {
				err = fmt.Errorf("could not find token: %s", token)
			}
			if err != nil {
				return nil, err
			}

			return result.output, nil
		},
	).Times(len(listAccountsResults))

	if listParentsResults == nil {
		return &m
	}

	m.EXPECT().DescribeOrganizationalUnit(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, input *organizations.DescribeOrganizationalUnitInput, _ ...func(*organizations.Options)) (*organizations.DescribeOrganizationalUnitOutput, error) {
			if input.OrganizationalUnitId == nil {
				return nil, errors.New("organizational unit id is nil")
			}
			if describeOrganizationalUnitError != nil {
				return nil, describeOrganizationalUnitError
			}
			id := *input.OrganizationalUnitId
			return &organizations.DescribeOrganizationalUnitOutput{
				OrganizationalUnit: &types.OrganizationalUnit{
					Arn:  aws.String(fmt.Sprintf("%s-arn", id)),
					Id:   &id,
					Name: aws.String(fmt.Sprintf("%s-name", id)),
				},
			}, nil
		},
	).Maybe()
	m.EXPECT().ListParents(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, input *organizations.ListParentsInput, _ ...func(*organizations.Options)) (*organizations.ListParentsOutput, error) {
			id := pointers.Deref(input.ChildId)
			result, ok := listParentsResults[id]
			err := result.err
			if !ok {
				err = fmt.Errorf("could not find child: %s", id)
			}
			if err != nil {
				return nil, err
			}

			return result.output, nil
		},
	).Times(len(listParentsResults))

	return &m
}

func Test_getOUInfoForAccount(t *testing.T) {
	ctx := context.Background()
	accountId := "123"

	t.Run("error in list", func(t *testing.T) {
		m := &mockOrganizationsAPI{}
		defer m.AssertExpectations(t)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{ChildId: &accountId},
			mock.Anything,
		).Return(nil, errors.New("some-error"))

		_, err := getOUInfoForAccount(ctx, m, nil, &accountId)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("error in list with NextToken", func(t *testing.T) {
		m := &mockOrganizationsAPI{}
		defer m.AssertExpectations(t)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{ChildId: &accountId},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{NextToken: aws.String("some-token")}, nil)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{
				ChildId:   &accountId,
				NextToken: aws.String("some-token"),
			},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{}, errors.New("some-error"))

		_, err := getOUInfoForAccount(ctx, m, nil, &accountId)
		require.ErrorContains(t, err, "some-error")
	})

	t.Run("no parents?", func(t *testing.T) {
		m := &mockOrganizationsAPI{}
		defer m.AssertExpectations(t)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{ChildId: &accountId},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{NextToken: aws.String("some-token")}, nil)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{
				ChildId:   &accountId,
				NextToken: aws.String("some-token"),
			},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{
			Parents: []types.Parent{},
		}, nil)

		_, err := getOUInfoForAccount(ctx, m, nil, &accountId)
		require.ErrorContains(t, err, "empty response")
	})

	t.Run("root ou", func(t *testing.T) {
		m := &mockOrganizationsAPI{}
		defer m.AssertExpectations(t)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{ChildId: &accountId},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{NextToken: aws.String("some-token")}, nil)
		m.EXPECT().ListParents(
			mock.Anything,
			&organizations.ListParentsInput{
				ChildId:   &accountId,
				NextToken: aws.String("some-token"),
			},
			mock.Anything,
		).Return(&organizations.ListParentsOutput{
			Parents: []types.Parent{{
				Id:   aws.String("root-id"),
				Type: types.ParentTypeRoot,
			}},
		}, nil)

		got, err := getOUInfoForAccount(ctx, m, nil, &accountId)
		require.NoError(t, err)
		assert.Equal(t, organizationalUnitInfo{
			id:   "root-id",
			name: "Root",
		}, got)
	})
}

func Test_describeOU(t *testing.T) {
	ctx := context.Background()
	ouID := "123"
	ou := types.OrganizationalUnit{
		Arn:  aws.String("some-arn"),
		Id:   &ouID,
		Name: aws.String("some-name"),
	}

	m := &mockOrganizationsAPI{}
	defer m.AssertExpectations(t)
	m.EXPECT().DescribeOrganizationalUnit(mock.Anything, &organizations.DescribeOrganizationalUnitInput{OrganizationalUnitId: &ouID}).
		Return(&organizations.DescribeOrganizationalUnitOutput{
			OrganizationalUnit: &ou,
		}, nil).Times(2)

	t.Run("protect against nil cache", func(t *testing.T) {
		got, err := describeOU(ctx, m, nil, &ouID)
		assert.Equal(t, organizationalUnitInfo{
			id:   ouID,
			name: "some-name",
		}, got)
		require.NoError(t, err)
	})

	cache := map[string]string{}

	t.Run("first", func(t *testing.T) {
		got, err := describeOU(ctx, m, cache, &ouID)
		assert.Equal(t, organizationalUnitInfo{
			id:   ouID,
			name: "some-name",
		}, got)
		require.NoError(t, err)
		assert.Len(t, cache, 1)
	})

	t.Run("use cache", func(t *testing.T) {
		got, err := describeOU(ctx, m, cache, &ouID)
		assert.Equal(t, organizationalUnitInfo{
			id:   ouID,
			name: "some-name",
		}, got)
		require.NoError(t, err)
		assert.Len(t, cache, 1)
	})
}
