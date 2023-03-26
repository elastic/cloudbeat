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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type (
	Client interface {
		DescribeHub(ctx context.Context, params *securityhub.DescribeHubInput, optFns ...func(*securityhub.Options)) (*securityhub.DescribeHubOutput, error)
	}

	Provider struct {
		log       *logp.Logger
		clients   map[string]Client
		region    string
		accountId string
	}
)

func NewProvider(cfg aws.Config, log *logp.Logger, factory awslib.CrossRegionFactory[Client], accountId string) *Provider {
	f := func(cfg aws.Config) Client {
		return securityhub.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(awslib.AllRegionSelector(), cfg, f, log)
	return &Provider{
		log:       log,
		region:    cfg.Region,
		accountId: accountId,
		clients:   m.GetMultiRegionsClientMap(),
	}
}

func (p *Provider) Describe(ctx context.Context) ([]SecurityHub, error) {
	return awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) (SecurityHub, error) {
		out, err := c.DescribeHub(ctx, &securityhub.DescribeHubInput{})
		if err != nil {
			res := SecurityHub{
				Enabled:           false,
				Region:            region,
				AccountId:         p.accountId,
				DescribeHubOutput: out,
			}
			if strings.Contains(err.Error(), "is not subscribed to AWS Security Hub") {
				return res, nil
			}
			return res, err
		}
		return SecurityHub{
			Enabled:           true,
			DescribeHubOutput: out,
			AccountId:         p.accountId,
			Region:            region,
		}, nil
	})
}
