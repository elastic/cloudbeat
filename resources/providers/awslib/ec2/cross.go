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

package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type crossRegionProvider struct {
	CrossRegionFetcher awslib.CrossRegionFetcher[ElasticCompute]
}

func (c *crossRegionProvider) DescribeNetworkAcl(ctx context.Context) ([]awslib.AwsResource, error) {
	return c.CrossRegionFetcher.Fetch(func(ec ElasticCompute) ([]awslib.AwsResource, error) {
		return ec.DescribeNetworkAcl(ctx)
	})
}

func (c *crossRegionProvider) DescribeSecurityGroups(ctx context.Context) ([]awslib.AwsResource, error) {
	return c.CrossRegionFetcher.Fetch(func(ec ElasticCompute) ([]awslib.AwsResource, error) {
		return ec.DescribeSecurityGroups(ctx)
	})
}

func NewCrossRegionProvider(log *logp.Logger, awsAccountID string, cfg aws.Config, factory awslib.CrossRegionFactory[ElasticCompute]) ElasticCompute {
	f := func(cfg aws.Config) ElasticCompute {
		return NewEC2Provider(log, awsAccountID, cfg)
	}

	m := factory.NewMultiRegionClients(ec2.NewFromConfig(cfg), cfg, f, log)
	return &crossRegionProvider{
		CrossRegionFetcher: m,
	}
}
