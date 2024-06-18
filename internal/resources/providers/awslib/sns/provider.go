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
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	log     *logp.Logger
	clients map[string]Client
}

type Client interface {
	sns.ListTopicsAPIClient
	sns.ListSubscriptionsByTopicAPIClient
}

func (p *Provider) ListTopics(ctx context.Context) ([]types.Topic, error) {
	topics, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, _ string, c Client) ([]types.Topic, error) {
		var all []types.Topic
		input := &sns.ListTopicsInput{}

		for {
			output, err := c.ListTopics(ctx, input)
			if err != nil {
				p.log.Errorf("Could not list SNS Topics. Error: %s", err)
			}
			all = append(all, output.Topics...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
		return all, nil
	})
	return lo.Flatten(topics), err
}

func (p *Provider) ListSubscriptionsByTopic(ctx context.Context, region string, topic string) ([]types.Subscription, error) {
	input := sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(topic),
	}
	var all []types.Subscription
	client, err := awslib.GetClient(pointers.Ref(region), p.clients)
	if err != nil {
		return nil, err
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

func (p *Provider) ListTopicsWithSubscriptions(ctx context.Context) ([]awslib.AwsResource, error) {
	topicInfos, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var all []awslib.AwsResource
		input := &sns.ListTopicsInput{}

		for {
			output, err := c.ListTopics(ctx, input)
			if err != nil {
				p.log.Errorf("Could not list SNS Topics. Error: %s", err)
			}

			for _, topic := range output.Topics {
				topicInfo := &TopicInfo{
					Topic:  topic,
					region: region,
				}
				subscriptions, err := p.ListSubscriptionsByTopic(ctx, region, topicInfo.GetResourceArn())
				if err != nil {
					p.log.Errorf("Could not list SNS Subscriptions for Topic %q. Error: %s", topicInfo.GetResourceArn(), err)
				} else {
					topicInfo.Subscriptions = subscriptions
				}
				all = append(all, topicInfo)
			}
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
		return all, nil
	})
	return lo.Flatten(topicInfos), err
}
