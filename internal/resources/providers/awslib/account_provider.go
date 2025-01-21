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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type AccountProviderAPI interface {
	ListAccounts(ctx context.Context, log *clog.Logger, cfg aws.Config) ([]cloud.Identity, error)
}

type organizationsAPI interface {
	organizations.ListAccountsAPIClient
	organizations.ListParentsAPIClient
	DescribeOrganizationalUnit(context.Context, *organizations.DescribeOrganizationalUnitInput, ...func(*organizations.Options)) (*organizations.DescribeOrganizationalUnitOutput, error)
}

type AccountProvider struct{}

func (a AccountProvider) ListAccounts(ctx context.Context, log *clog.Logger, cfg aws.Config) ([]cloud.Identity, error) {
	return listAccounts(ctx, log, organizations.NewFromConfig(cfg))
}

func listAccounts(ctx context.Context, log *clog.Logger, client organizationsAPI) ([]cloud.Identity, error) {
	organizationIdToName := make(map[string]string)

	input := organizations.ListAccountsInput{}
	var accounts []cloud.Identity
	for {
		o, err := client.ListAccounts(ctx, &input)
		if err != nil {
			return nil, err
		}

		for _, account := range o.Accounts {
			if account.Status != types.AccountStatusActive || account.Id == nil {
				continue
			}

			organization, err := getOUInfoForAccount(ctx, client, organizationIdToName, account.Id)
			if err != nil {
				log.Errorf("failed to get organizational unit info for account %s: %v", *account.Id, err)
			}
			accounts = append(accounts, cloud.Identity{
				Provider:         "aws",
				Account:          *account.Id,
				AccountAlias:     pointers.Deref(account.Name),
				OrganizationId:   organization.id,
				OrganizationName: organization.name,
			})
		}

		if o.NextToken == nil {
			break
		}
		input.NextToken = o.NextToken
	}
	return accounts, nil
}

type organizationalUnitInfo struct {
	id   string
	name string
}

func getOUInfoForAccount(ctx context.Context, client organizationsAPI, cache map[string]string, accountId *string) (organizationalUnitInfo, error) {
	// We need a paginator, according to the AWS docs:
	// These operations can occasionally return an empty set of results even when there are more results available.
	paginator := organizations.NewListParentsPaginator(client, &organizations.ListParentsInput{ChildId: accountId})
	for paginator.HasMorePages() {
		o, err := paginator.NextPage(ctx)
		if err != nil {
			return organizationalUnitInfo{}, err
		}
		if len(o.Parents) == 0 {
			continue
		}

		// According to AWS, in the current release, a child can have only a single parent.
		parent := o.Parents[0]

		if parent.Type == types.ParentTypeRoot {
			return organizationalUnitInfo{
				id:   *parent.Id,
				name: "Root",
			}, nil
		}

		return describeOU(ctx, client, cache, parent.Id)
	}

	return organizationalUnitInfo{}, errors.New("empty response")
}

func describeOU(ctx context.Context, client organizationsAPI, cache map[string]string, id *string) (organizationalUnitInfo, error) {
	if id == nil {
		return organizationalUnitInfo{}, errors.New("nil id")
	}

	if savedName, ok := cache[*id]; ok {
		return organizationalUnitInfo{
			id:   *id,
			name: savedName,
		}, nil
	}

	o, err := client.DescribeOrganizationalUnit(ctx, &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: id,
	})
	if err != nil {
		return organizationalUnitInfo{id: *id}, err
	}

	name := pointers.Deref(o.OrganizationalUnit.Name)
	if cache != nil {
		cache[*id] = name
	}
	return organizationalUnitInfo{
		id:   *id,
		name: name,
	}, nil
}
