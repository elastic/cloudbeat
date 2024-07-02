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

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type SNS interface {
	ListTopics(ctx context.Context) ([]types.Topic, error)
	ListSubscriptionsByTopic(ctx context.Context, region string, topic string) ([]types.Subscription, error)
	ListTopicsWithSubscriptions(ctx context.Context) ([]awslib.AwsResource, error)
}

func NewSNSProvider(log *logp.Logger, cfg aws.Config, factory awslib.CrossRegionFactory[Client]) *Provider {
	f := func(cfg aws.Config) Client {
		return sns.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(awslib.AllRegionSelector(), cfg, f, log)
	return &Provider{
		log:     log,
		clients: m.GetMultiRegionsClientMap(),
	}
}
