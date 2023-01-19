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

package aws_cis

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
	Cloudtrail     cloudtrail.CloudTrail
	Cloudwatch     cloudwatch.Cloudwatch
	Cloudwatchlogs logs.CloudwatchLogs
	Sns            sns.SNS
	Log            *logp.Logger
}

type Client interface {
	Rule41(ctx context.Context) (Rule41Output, error)
}

type (
	Rule41Output struct {
		Items []Rule41Item
	}
	Rule41Item struct {
		TrailInfo                 cloudtrail.TrailInfo
		Topics                    []string
		AuthorizationFilterExists bool
	}
)

// Rule41 (4.1)
func (p *Provider) Rule41(ctx context.Context) (Rule41Output, error) {
	trails, err := p.Cloudtrail.DescribeCloudTrails(ctx)
	if err != nil {
		return Rule41Output{}, err
	}

	items := []Rule41Item{}
	for _, trail := range trails {
		if trail.Trail.CloudWatchLogsLogGroupArn == nil {
			continue
		}
		filter := filterNameFromARN(trail.Trail.CloudWatchLogsLogGroupArn)
		if filter == "" {
			p.Log.Warnf("cloudwatchlogs log group arn has no log group name %s", *trail.Trail.CloudWatchLogsLogGroupArn)
			continue
		}
		metrics, err := p.Cloudwatchlogs.DescribeMetricFilters(ctx, filter)
		if err != nil {
			p.Log.Errorf("failed to describe metric filters for cloudwatchlog log group arn %s: %v", *trail.Trail.CloudWatchLogsLogGroupArn, err)
			continue
		}
		filters := metricsMatchToPattern(metrics, []string{UnauthorizedAPICallsPattern})
		if len(filters) == 0 {
			items = append(items, Rule41Item{
				TrailInfo:                 trail,
				AuthorizationFilterExists: false,
			})
			continue
		}

		alarms, err := p.Cloudwatch.DescribeAlarms(ctx, filterNamesFromMetrics(filters))
		if err != nil {
			p.Log.Errorf("failed to describe alarms for cloudwatch filter %v: %v", filters, err)
			continue
		}
		topics := p.getSubscriptionForAlarms(ctx, alarms)
		items = append(items, Rule41Item{
			TrailInfo:                 trail,
			Topics:                    topics,
			AuthorizationFilterExists: true,
		})
	}

	return Rule41Output{Items: items}, nil
}

func (p *Provider) getSubscriptionForAlarms(ctx context.Context, alarms []cloudwatch_types.MetricAlarm) []string {
	topics := []string{}
	for _, alarm := range alarms {
		for _, action := range alarm.AlarmActions {
			subscriptions, err := p.Sns.ListSubscriptionsByTopic(ctx, action)
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

func metricsMatchToPattern(list []cloudwatchlogs_types.MetricFilter, patterns []string) []cloudwatchlogs_types.MetricFilter {
	filters := []cloudwatchlogs_types.MetricFilter{}
	for _, metric := range list {
		if metric.FilterPattern == nil {
			continue
		}
		for _, p := range patterns {
			if *metric.FilterPattern == p {
				filters = append(filters, metric)
				break
			}
		}
	}
	return filters
}

func filterNamesFromMetrics(list []cloudwatchlogs_types.MetricFilter) []string {
	names := []string{}
	for _, filter := range list {
		if filter.FilterName != nil {
			names = append(names, *filter.FilterName)
		}
	}
	return names
}

func filterNameFromARN(arn *string) string {
	if arn == nil {
		return ""
	}
	parts := strings.Split(*arn, ":")
	if len(parts) < 6 {
		return ""
	}
	return parts[6]
}
