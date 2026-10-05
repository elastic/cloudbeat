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

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

// Client abstracts the subset of the EKS SDK client used by the provider.
type Client interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
}

type ClustersDescriber interface {
	DescribeClusters(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log     *clog.Logger
	clients map[string]Client
}

func NewProvider(ctx context.Context, log *clog.Logger, cfg aws.Config, factory awslib.CrossRegionFactory[Client]) *Provider {
	f := func(cfg aws.Config) Client {
		return eks.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ctx, awslib.AllRegionSelector(), cfg, f, log)
	return &Provider{
		log:     log,
		clients: m.GetMultiRegionsClientMap(),
	}
}
