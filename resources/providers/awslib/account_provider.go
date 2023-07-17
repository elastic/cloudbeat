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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"

	"github.com/elastic/cloudbeat/resources/utils/strings"
)

type AccountProviderAPI interface {
	ListAccounts(ctx context.Context, cfg aws.Config) ([]Identity, error)
}

type AccountProvider struct{}

func (a AccountProvider) ListAccounts(ctx context.Context, cfg aws.Config) ([]Identity, error) {
	return listAccounts(ctx, organizations.NewFromConfig(cfg))
}

func listAccounts(ctx context.Context, client organizations.ListAccountsAPIClient) ([]Identity, error) {
	input := organizations.ListAccountsInput{}
	var accounts []Identity
	for {
		o, err := client.ListAccounts(ctx, &input)
		if err != nil {
			return nil, err
		}

		for _, account := range o.Accounts {
			if account.Status != types.AccountStatusActive || account.Id == nil {
				continue
			}

			accounts = append(accounts, Identity{
				Account: *account.Id,
				Alias:   strings.Dereference(account.Name),
			})
		}

		if o.NextToken == nil {
			break
		}
		input.NextToken = o.NextToken
	}
	return accounts, nil
}
