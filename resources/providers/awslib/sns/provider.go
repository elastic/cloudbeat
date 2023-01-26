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
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	log     *logp.Logger
	clients map[string]Client
}

type Client interface {
	ListSubscriptionsByTopic(ctx context.Context, params *sns.ListSubscriptionsByTopicInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsByTopicOutput, error)
}

func (p *Provider) ListSubscriptionsByTopic(ctx context.Context, region *string, topic string) ([]types.Subscription, error) {
	input := sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(topic),
	}
	all := []types.Subscription{}
	client, ok := p.clients[awslib.GetRegion(region)]
	if !ok {
		return nil, awslib.ErrRegionNotFound
	}
	for {
		output, err := client.ListSubscriptionsByTopic(ctx, &input)
		if err != nil {
			return nil, err
		}
		all = append(all, output.Subscriptions...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}
	return all, nil
}
