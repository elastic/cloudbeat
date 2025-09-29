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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const provider = "aws"

type IdentityProviderGetter interface {
	GetIdentity(ctx context.Context, cfg aws.Config) (*cloud.Identity, error)
	GetCallerIdentity(ctx context.Context, cfg aws.Config) (sts.GetCallerIdentityOutput, error)
}

type IdentityProvider struct {
	Logger *clog.Logger
}

// GetIdentity returns AWS identity information
func (p IdentityProvider) GetIdentity(ctx context.Context, cfg aws.Config) (*cloud.Identity, error) {
	response, err := p.GetCallerIdentity(ctx, cfg)
	if err != nil {
		return nil, err
	}

	alias, err := p.getAccountAlias(ctx, cfg)
	if err != nil {
		p.Logger.Warnf("failed to get aliases: %v", err)
		alias = ""
	}

	// if alias is not configured, try account name.
	if alias == "" {
		name, err := p.getAccountName(ctx, cfg, response.Account)
		if err != nil {
			p.Logger.Warnf("failed to get account name: %v", err)
		}
		alias = name
	}

	return &cloud.Identity{
		Account:      *response.Account,
		AccountAlias: alias,
		Provider:     provider,
	}, nil
}

func (IdentityProvider) getAccountAlias(ctx context.Context, cfg aws.Config) (string, error) {
	aliases, err := iam.NewFromConfig(cfg).ListAccountAliases(ctx, &iam.ListAccountAliasesInput{})
	if err != nil {
		return "", err
	}

	if len(aliases.AccountAliases) > 0 {
		return aliases.AccountAliases[0], nil
	}

	return "", nil
}

func (p IdentityProvider) getAccountName(ctx context.Context, cfg aws.Config, accountID *string) (string, error) {
	// "organizations:Describe*" is part of AWS SecurityAudit managed policy, cloudbeat-asset-inventory-root and cloudbeat-root role.
	acctResp, err := organizations.NewFromConfig(cfg).DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: accountID,
	})
	if err != nil || acctResp == nil {
		return "", err
	}

	if acctResp.Account == nil {
		return "", nil
	}

	return pointers.Deref(acctResp.Account.Name), nil
}

func (p IdentityProvider) GetCallerIdentity(ctx context.Context, cfg aws.Config) (sts.GetCallerIdentityOutput, error) {
	id, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return sts.GetCallerIdentityOutput{}, fmt.Errorf("failed to get caller identity: %w", err)
	}

	if id == nil {
		return sts.GetCallerIdentityOutput{}, errors.New("empty caller identity")
	}

	return *id, nil
}
