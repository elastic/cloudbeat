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

package monitoring

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudtrail_aws "github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrail_types "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	cloudwatchlogs_types "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	sns_types "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type (
	mocks       []any
	clientMocks map[string][2]mocks
)

var (
	filter                         = "metric-filter-name"
	metricFilterWithExpectedFilter = cloudwatchlogs_types.MetricFilter{
		FilterName:    aws.String(filter),
		FilterPattern: aws.String("{a = b}"),
	}
	metricFilterWithoutFilter = cloudwatchlogs_types.MetricFilter{
		FilterName: aws.String(filter),
	}
	logGroupArn = "arn:aws:logs:us-east-1:account:log-group:cloudwatchlogs-log-group-arn"
	snsTopicArn = "sns-topic-arn"

	describeCloudTrailWithResults = [2]mocks{
		{mock.Anything, mock.Anything},
		{[]cloudtrail.TrailInfo{
			{
				Trail: cloudtrail_types.Trail{
					CloudWatchLogsLogGroupArn: aws.String(logGroupArn),
					IsMultiRegionTrail:        aws.Bool(true),
				},
				Status:         expectedCommonTrailStatus,
				EventSelectors: expectedCommonTrailEventSelector,
			},
		}, nil},
	}
	describeCloudTrailWithoutResults = [2]mocks{
		{mock.Anything, mock.Anything},
		{[]cloudtrail.TrailInfo{}, nil},
	}

	metricFilterCallWithExpectedFilter = [2]mocks{
		{mock.Anything, mock.Anything, mock.Anything},
		{[]cloudwatchlogs_types.MetricFilter{metricFilterWithExpectedFilter}, nil},
	}

	metricFilterCallWithoutFilter = [2]mocks{
		{mock.Anything, mock.Anything, mock.Anything},
		{[]cloudwatchlogs_types.MetricFilter{metricFilterWithoutFilter}, nil},
	}

	describeAlarmCallWithSNSTopic = [2]mocks{
		{mock.Anything, mock.Anything, mock.Anything},
		{[]cloudwatch_types.MetricAlarm{
			{AlarmActions: []string{snsTopicArn}},
		}, nil},
	}

	listSubscriptionCallWithResult = [2]mocks{
		{mock.Anything, mock.Anything, mock.Anything},
		{[]sns_types.Subscription{{TopicArn: aws.String(snsTopicArn)}}, nil},
	}

	expectedCommonTrail = cloudtrail_types.Trail{
		CloudWatchLogsLogGroupArn: aws.String(logGroupArn),
		IsMultiRegionTrail:        aws.Bool(true),
	}

	expectedCommonTrailStatus = &cloudtrail_aws.GetTrailStatusOutput{
		IsLogging: aws.Bool(true),
	}
	expectedCommonTrailEventSelector = []cloudtrail_types.EventSelector{
		{
			IncludeManagementEvents: aws.Bool(true),
			ReadWriteType:           cloudtrail_types.ReadWriteTypeAll,
		},
	}
)

func TestProvider_AggregateResources(t *testing.T) {
	type fields struct {
		cloudtrailMocks     clientMocks
		cloudwatchMocks     clientMocks
		cloudwatchlogsMocks clientMocks
		snsMocks            clientMocks
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Resource
		wantErr bool
	}{
		{
			name: "no trails found",
			fields: fields{
				cloudtrailMocks: clientMocks{
					"DescribeTrails": describeCloudTrailWithoutResults,
				},
			},
			want: &Resource{Items: []MonitoringItem{}},
		},
		{
			name: "one trail with filter and sns setup",
			fields: fields{
				cloudtrailMocks: clientMocks{
					"DescribeTrails": describeCloudTrailWithResults,
				},
				cloudwatchlogsMocks: clientMocks{
					"DescribeMetricFilters": metricFilterCallWithExpectedFilter,
				},
				cloudwatchMocks: clientMocks{
					"DescribeAlarms": describeAlarmCallWithSNSTopic,
				},
				snsMocks: clientMocks{
					"ListSubscriptionsByTopic": listSubscriptionCallWithResult,
				},
			},
			want: &Resource{
				Items: []MonitoringItem{
					{
						TrailInfo: cloudtrail.TrailInfo{
							Trail:          expectedCommonTrail,
							Status:         expectedCommonTrailStatus,
							EventSelectors: expectedCommonTrailEventSelector,
						},
						MetricFilters: []MetricFilter{
							{
								MetricFilter:        metricFilterWithExpectedFilter,
								ParsedFilterPattern: newSimpleExpression("a", coEqual, "b")},
						},

						MetricTopicBinding: map[string][]string{
							filter: {snsTopicArn},
						},
					},
				},
			},
		},
		{
			name: "trail with no associated filter",
			fields: fields{
				cloudtrailMocks: clientMocks{
					"DescribeTrails": describeCloudTrailWithResults,
				},
				cloudwatchlogsMocks: clientMocks{
					"DescribeMetricFilters": metricFilterCallWithoutFilter,
				},
			},
			want: &Resource{
				Items: []MonitoringItem{
					{
						TrailInfo: cloudtrail.TrailInfo{
							Trail:          expectedCommonTrail,
							Status:         expectedCommonTrailStatus,
							EventSelectors: expectedCommonTrailEventSelector,
						},
						MetricTopicBinding: map[string][]string{},
						MetricFilters: []MetricFilter{
							{MetricFilter: metricFilterWithoutFilter},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := &cloudtrail.MockTrailService{}
			for name, call := range tt.fields.cloudtrailMocks {
				ct.On(name, call[0]...).Return(call[1]...)
			}
			cw := &cloudwatch.MockCloudwatch{}
			for name, call := range tt.fields.cloudwatchMocks {
				cw.On(name, call[0]...).Return(call[1]...)
			}
			cwl := &logs.MockCloudwatchLogs{}
			for name, call := range tt.fields.cloudwatchlogsMocks {
				cwl.On(name, call[0]...).Return(call[1]...)
			}
			mockSNS := &sns.MockSNS{}
			for name, call := range tt.fields.snsMocks {
				mockSNS.On(name, call[0]...).Return(call[1]...)
			}
			p := &Provider{
				Cloudtrail:     ct,
				Cloudwatch:     cw,
				Cloudwatchlogs: cwl,
				Sns:            mockSNS,
				Log:            testhelper.NewLogger(t),
			}
			got, err := p.AggregateResources(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
