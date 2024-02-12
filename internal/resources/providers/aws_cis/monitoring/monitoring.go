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

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	cloudwatchlogs_types "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
)

type Provider struct {
	Cloudtrail     cloudtrail.TrailService
	Cloudwatch     cloudwatch.Cloudwatch
	Cloudwatchlogs logs.CloudwatchLogs
	Sns            sns.SNS
	Log            *logp.Logger
}

type Client interface {
	// AggregateResources will gather all the resource to be used for aws cis 4.1 ... 4.15 rules
	AggregateResources(ctx context.Context) (*Resource, error)
}

type (
	Resource struct {
		Items []MonitoringItem
	}

	MonitoringItem struct {
		TrailInfo          cloudtrail.TrailInfo
		MetricFilters      []cloudwatchlogs_types.MetricFilter
		MetricTopicBinding map[string][]string
	}
)

func NewProvider(log *logp.Logger, awsConfig aws.Config, trailCrossRegionFactory awslib.CrossRegionFactory[cloudtrail.Client], cloudwatchCrossResignFactory awslib.CrossRegionFactory[cloudwatch.Client], cloudwatchlogsCrossRegionFactory awslib.CrossRegionFactory[logs.Client], snsCrossRegionFactory awslib.CrossRegionFactory[sns.Client]) *Provider {
	return &Provider{
		Cloudtrail:     cloudtrail.NewProvider(log, awsConfig, trailCrossRegionFactory),
		Cloudwatch:     cloudwatch.NewProvider(log, awsConfig, cloudwatchCrossResignFactory),
		Cloudwatchlogs: logs.NewCloudwatchLogsProvider(log, awsConfig, cloudwatchlogsCrossRegionFactory),
		Sns:            sns.NewSNSProvider(log, awsConfig, snsCrossRegionFactory),
		Log:            log,
	}
}

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
				TrailInfo:          info,
				MetricFilters:      []cloudwatchlogs_types.MetricFilter{},
				MetricTopicBinding: map[string][]string{},
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
				TrailInfo:          info,
				MetricFilters:      metrics,
				MetricTopicBinding: map[string][]string{},
			})
			continue
		}
		bindings := map[string][]string{}
		for _, name := range names {
			alarms, err := p.Cloudwatch.DescribeAlarms(ctx, info.Trail.HomeRegion, []string{name})
			if err != nil {
				p.Log.Errorf("failed to describe alarms for cloudwatch filter %v: %v", names, err)
				continue
			}
			topics := p.getSubscriptionForAlarms(ctx, info.Trail.HomeRegion, alarms)
			bindings[name] = topics
		}
		items = append(items, MonitoringItem{
			TrailInfo:          info,
			MetricFilters:      metrics,
			MetricTopicBinding: bindings,
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
