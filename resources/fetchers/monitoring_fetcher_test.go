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

package fetchers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_securityhub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

func TestMonitoringFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name              string
		monitoring        clientMocks
		securityhub       clientMocks
		wantErr           bool
		expectedResources int
	}{
		{
			name: "with resources",
			monitoring: clientMocks{
				"AggregateResources": [2]mocks{
					{mock.Anything},
					{&monitoring.Resource{
						Items: []monitoring.MonitoringItem{
							{},
							{},
						},
					}, nil},
				},
			},
			securityhub: clientMocks{
				"Describe": [2]mocks{
					{mock.Anything},
					{[]securityhub.SecurityHub{{}}, nil},
				},
			},
			expectedResources: 2,
		},
		{
			name: "with error",
			monitoring: clientMocks{
				"AggregateResources": [2]mocks{
					{mock.Anything},
					{nil, fmt.Errorf("failed to run provider")},
				},
			},
			securityhub: clientMocks{
				"Describe": [2]mocks{
					{mock.Anything},
					{[]securityhub.SecurityHub{{}}, fmt.Errorf("failed to run provider")},
				},
			},
		},
		{
			name: "with securityhub",
			monitoring: clientMocks{
				"AggregateResources": [2]mocks{
					{mock.Anything},
					{nil, fmt.Errorf("failed to run provider")},
				},
			},
			securityhub: clientMocks{
				"Describe": [2]mocks{
					{mock.Anything},
					{[]securityhub.SecurityHub{{}}, nil},
				},
			},
			expectedResources: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetching.ResourceInfo, 100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			mockClient := &monitoring.MockClient{}
			for name, call := range tt.monitoring {
				mockClient.On(name, call[0]...).Return(call[1]...)
			}

			hub := &securityhub.MockService{}
			for name, call := range tt.securityhub {
				hub.On(name, call[0]...).Return(call[1]...)
			}
			m := MonitoringFetcher{
				log:           logp.NewLogger("TestMonitoringFetcher_Fetch"),
				provider:      mockClient,
				securityhub:   hub,
				cfg:           MonitoringFetcherConfig{},
				resourceCh:    ch,
				cloudIdentity: &awslib.Identity{Account: aws.String("account")},
			}

			err := m.Fetch(ctx, fetching.CycleMetadata{})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			resources := testhelper.CollectResources(ch)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResources, len(resources))
		})
	}
}

func TestMonitoringResource_GetMetadata(t *testing.T) {
	type fields struct {
		Resource monitoring.Resource
		identity *awslib.Identity
	}
	tests := []struct {
		name    string
		fields  fields
		want    fetching.ResourceMetadata
		wantErr bool
	}{
		{
			name: "without trails",
			fields: fields{
				identity: &awslib.Identity{Account: aws.String("aws-account-id")},
				Resource: monitoring.Resource{
					Items: []monitoring.MonitoringItem{},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      "cloudtrail-aws-account-id",
				Name:    "cloudtrail-aws-account-id",
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.MultiTrailsType,
				Region:  awslib.GlobalRegion,
			},
		},
		{
			name: "with trails",
			fields: fields{
				identity: &awslib.Identity{Account: aws.String("aws-account-id")},
				Resource: monitoring.Resource{
					Items: []monitoring.MonitoringItem{
						{},
						{},
					},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      "cloudtrail-aws-account-id",
				Name:    "cloudtrail-aws-account-id",
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.MultiTrailsType,
				Region:  awslib.GlobalRegion,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := MonitoringResource{
				Resource: tt.fields.Resource,
				identity: tt.fields.identity,
			}
			got, err := r.GetMetadata()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSecurityHubResource_GetMetadata(t *testing.T) {
	accountId := "dummy-account-id"

	type fields struct {
		SecurityHub securityhub.SecurityHub
	}
	tests := []struct {
		name    string
		fields  fields
		want    fetching.ResourceMetadata
		wantErr bool
	}{
		{
			name: "enabled",
			fields: fields{
				SecurityHub: securityhub.SecurityHub{
					Enabled:   true,
					Region:    "us-east-1",
					AccountId: accountId,
					DescribeHubOutput: &aws_securityhub.DescribeHubOutput{
						HubArn: aws.String("hub:arn"),
					},
				},
			},
			want: fetching.ResourceMetadata{
				ID:      "hub:arn",
				Name:    "securityhub-us-east-1-" + accountId,
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.SecurityHubType,
				Region:  "us-east-1",
			},
		},
		{
			name: "disabled",
			fields: fields{
				SecurityHub: securityhub.SecurityHub{
					Enabled:   false,
					AccountId: accountId,
					Region:    "us-east-2",
				},
			},
			want: fetching.ResourceMetadata{
				ID:      "securityhub-us-east-2-" + accountId,
				Name:    "securityhub-us-east-2-" + accountId,
				Type:    fetching.MonitoringIdentity,
				SubType: fetching.SecurityHubType,
				Region:  "us-east-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SecurityHubResource{
				SecurityHub: tt.fields.SecurityHub,
			}
			got, err := s.GetMetadata()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
