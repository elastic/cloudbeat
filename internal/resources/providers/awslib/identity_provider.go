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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

const provider = "aws"

type IdentityProviderGetter interface {
	GetIdentity(ctx context.Context, cfg aws.Config) (*cloud.Identity, error)
}

type IdentityProvider struct {
	Logger *clog.Logger
}

// GetIdentity returns AWS identity information
func (p IdentityProvider) GetIdentity(ctx context.Context, cfg aws.Config) (*cloud.Identity, error) {
	response, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	alias, err := p.getAccountAlias(ctx, cfg)
	if err != nil {
		p.Logger.Warnf("failed to get aliases: %v", err)
		alias = ""
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
