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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

var regions = []string{"us-east-1"}

func TestProvider_ListTopics(t *testing.T) {
	tests := []struct {
		name    string
		mocks   clientMocks
		topic   string
		want    []types.Topic
		wantErr bool
	}{
		{
			name: "with results",
			mocks: clientMocks{
				"ListTopics": [2]mocks{
					{mock.Anything, mock.Anything},
					{&sns.ListTopicsOutput{
						Topics: []types.Topic{
							{
								TopicArn: aws.String("topic-arn"),
							},
						},
					}, nil},
				},
			},
			want: []types.Topic{
				{
					TopicArn: aws.String("topic-arn"),
				},
			},
			topic: "topic-arn",
		},
		{
			name: "with error",
			mocks: clientMocks{
				"ListTopics": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, awslib.ErrMock},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockClient{}
			for name, call := range tt.mocks {
				c.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				clients: createMockClients(c, regions),
			}
			got, err := p.ListTopics(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_ListSubscriptionsByTopic(t *testing.T) {
	tests := []struct {
		name    string
		mocks   clientMocks
		topic   string
		want    []types.Subscription
		wantErr bool
	}{
		{
			name: "with results",
			mocks: clientMocks{
				"ListSubscriptionsByTopic": [2]mocks{
					{mock.Anything, mock.Anything},
					{&sns.ListSubscriptionsByTopicOutput{
						Subscriptions: []types.Subscription{
							{
								TopicArn: aws.String("topic-arn"),
							},
						},
					}, nil},
				},
			},
			want: []types.Subscription{
				{
					TopicArn: aws.String("topic-arn"),
				},
			},
			topic: "topic-arn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockClient{}
			for name, call := range tt.mocks {
				c.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				clients: createMockClients(c, regions),
			}
			got, err := p.ListSubscriptionsByTopic(context.Background(), regions[0], tt.topic)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_ListTopicsWithSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		mocks   clientMocks
		topic   string
		want    []awslib.AwsResource
		wantErr bool
	}{
		{
			name: "with results",
			mocks: clientMocks{
				"ListTopics": [2]mocks{
					{mock.Anything, mock.Anything},
					{&sns.ListTopicsOutput{
						Topics: []types.Topic{
							{
								TopicArn: aws.String("topic-arn"),
							},
						},
					}, nil},
				},
				"ListSubscriptionsByTopic": [2]mocks{
					{mock.Anything, mock.Anything},
					{&sns.ListSubscriptionsByTopicOutput{
						Subscriptions: []types.Subscription{
							{
								TopicArn: aws.String("topic-arn"),
							},
						},
					}, nil},
				},
			},
			want: []awslib.AwsResource{
				&TopicInfo{
					Topic: types.Topic{
						TopicArn: pointers.Ref("topic-arn"),
					},
					Subscriptions: []types.Subscription{
						{
							TopicArn: pointers.Ref("topic-arn"),
						},
					},
					region: "us-east-1",
				},
			},
		},
		{
			name: "with error in ListTopics",
			mocks: clientMocks{
				"ListTopics": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, awslib.ErrMock},
				},
				"ListSubscriptionsByTopic": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, awslib.ErrMock},
				},
			},
			wantErr: true,
		},
		{
			name: "with error in ListSubscriptionsByTopic",
			mocks: clientMocks{
				"ListTopics": [2]mocks{
					{mock.Anything, mock.Anything},
					{&sns.ListTopicsOutput{
						Topics: []types.Topic{
							{
								TopicArn: aws.String("topic-arn"),
							},
						},
					}, nil},
				},
				"ListSubscriptionsByTopic": [2]mocks{
					{mock.Anything, mock.Anything},
					{nil, awslib.ErrMock},
				},
			},
			want: []awslib.AwsResource{
				&TopicInfo{
					Topic: types.Topic{
						TopicArn: pointers.Ref("topic-arn"),
					},
					Subscriptions: []types.Subscription{},
					region:        "us-east-1",
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &MockClient{}
			for name, call := range tt.mocks {
				c.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				clients: createMockClients(c, regions),
			}
			got, err := p.ListTopicsWithSubscriptions(context.Background())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func createMockClients(c Client, regions []string) map[string]Client {
	m := make(map[string]Client)
	for _, clientRegion := range regions {
		m[clientRegion] = c
	}

	return m
}
