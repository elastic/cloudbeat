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
	"strings"

	cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	cloudwatchlogs_types "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	Cloudtrail     cloudtrail.TrailService
	Cloudwatch     cloudwatch.Cloudwatch
	Cloudwatchlogs logs.CloudwatchLogs
	Sns            sns.SNS
	Log            *logp.Logger
}

type Client interface {
	AggregateResources(ctx context.Context) (*Resource, error)
}

type (
	Resource struct {
		Items []MonitoringItem
	}

	MonitoringItem struct {
		TrailInfo     cloudtrail.TrailInfo
		MetricFilters []cloudwatchlogs_types.MetricFilter
		Topics        []string
	}
)

// AggregateResources will gather all the resource to be used for aws cis 4.1 ... 4.15 rules
func (p *Provider) AggregateResources(ctx context.Context) (*Resource, error) {
	trails, err := p.Cloudtrail.DescribeTrails(ctx)
	if err != nil {
		return nil, err
	}

	items := []MonitoringItem{}
	for _, info := range trails {
		if info.Trail.CloudWatchLogsLogGroupArn == nil {
			items = append(items, MonitoringItem{
				TrailInfo:     info,
				MetricFilters: []cloudwatchlogs_types.MetricFilter{},
				Topics:        []string{},
			})
			continue
		}
		logGroup := getLogGroupFromARN(info.Trail.CloudWatchLogsLogGroupArn)
		if logGroup == "" {
			p.Log.Warnf("cloudwatchlogs log group arn has no log group name %s", *info.Trail.CloudWatchLogsLogGroupArn)
			continue
		}
		metrics, err := p.Cloudwatchlogs.DescribeMetricFilters(ctx, info.Trail.HomeRegion, logGroup)
		if err != nil {
			p.Log.Errorf("failed to describe metric filters for cloudwatchlog log group arn %s: %v", *info.Trail.CloudWatchLogsLogGroupArn, err)
			continue
		}

		names := filterNamesFromMetrics(metrics)
		if len(names) == 0 {
			items = append(items, MonitoringItem{
				TrailInfo:     info,
				MetricFilters: metrics,
				Topics:        []string{},
			})
			continue
		}
		alarms, err := p.Cloudwatch.DescribeAlarms(ctx, info.Trail.HomeRegion, names)
		if err != nil {
			p.Log.Errorf("failed to describe alarms for cloudwatch filter %v: %v", names, err)
			continue
		}
		topics := p.getSubscriptionForAlarms(ctx, info.Trail.HomeRegion, alarms)
		items = append(items, MonitoringItem{
			TrailInfo:     info,
			MetricFilters: metrics,
			Topics:        topics,
		})

	}

	return &Resource{Items: items}, nil
}

func (p *Provider) getSubscriptionForAlarms(ctx context.Context, region *string, alarms []cloudwatch_types.MetricAlarm) []string {
	topics := []string{}
	for _, alarm := range alarms {
		for _, action := range alarm.AlarmActions {
			subscriptions, err := p.Sns.ListSubscriptionsByTopic(ctx, region, action)
			if err != nil {
				p.Log.Errorf("failed to list subscriptions for topic %s: %v", action, err)
				continue
			}
			for _, topic := range subscriptions {
				topics = append(topics, *topic.TopicArn)
			}
		}
	}
	return topics
}

func filterNamesFromMetrics(list []cloudwatchlogs_types.MetricFilter) []string {
	names := []string{}
	for _, filter := range list {
		if filter.FilterName != nil && filter.FilterPattern != nil {
			names = append(names, *filter.FilterName)
		}
	}
	return names
}

func getLogGroupFromARN(arn *string) string {
	if arn == nil {
		return ""
	}
	parts := strings.Split(*arn, ":")
	if len(parts) < 6 {
		return ""
	}
	return parts[6]
}
