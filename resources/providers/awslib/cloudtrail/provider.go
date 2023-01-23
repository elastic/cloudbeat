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

package cloudtrail

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	log    *logp.Logger
	client TrailClient
}

type TrailClient interface {
	ListTrails(ctx context.Context, params *cloudtrail.ListTrailsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.ListTrailsOutput, error)
	GetTrail(ctx context.Context, params *cloudtrail.GetTrailInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetTrailOutput, error)
	DescribeTrails(ctx context.Context, params *cloudtrail.DescribeTrailsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error)
	GetTrailStatus(ctx context.Context, params *cloudtrail.GetTrailStatusInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetTrailStatusOutput, error)
	GetEventSelectors(ctx context.Context, params *cloudtrail.GetEventSelectorsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetEventSelectorsOutput, error)
}

func (p *Provider) ListTrails(ctx context.Context) ([]awslib.AwsResource, error) {
	output, err := p.client.ListTrails(ctx, nil)
	if err != nil {
		return nil, err
	}

	var result []awslib.AwsResource
	for _, trail := range output.Trails {

		if trail.Name == nil {
			continue
		}

		response, err := p.client.GetTrail(ctx, &cloudtrail.GetTrailInput{Name: trail.Name})
		if err != nil || response.Trail == nil {
			p.log.Errorf("fail to get trail info %s %s", *trail.TrailARN, err.Error())
			continue
		}
		tailInfo := response.Trail

		status, err := p.client.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{Name: trail.Name})
		if err != nil {
			p.log.Errorf("fail to get trail status %s %s", *trail.TrailARN, err.Error())
			status = &cloudtrail.GetTrailStatusOutput{}
		}

		selectors, err := p.getEventSelectors(ctx, tailInfo)
		if err != nil {
			p.log.Errorf("fail to get trail event selector %s %s", *trail.TrailARN, err.Error())
		}

		result = append(result, TrailInfo{
			TrailARN:                  getValue(trail.TrailARN),
			Name:                      getValue(trail.Name),
			EnableLogFileValidation:   getValue(tailInfo.LogFileValidationEnabled),
			IsMultiRegion:             getValue(tailInfo.IsMultiRegionTrail),
			KMSKeyID:                  getValue(tailInfo.KmsKeyId),
			CloudWatchLogsLogGroupArn: getValue(tailInfo.CloudWatchLogsLogGroupArn),
			IsLogging:                 getValue(status.IsLogging),
			SnsTopicARN:               getValue(tailInfo.SnsTopicARN),
			BucketName:                getValue(tailInfo.S3BucketName),
			EventSelectors:            selectors,
		})
	}
	return result, nil
}

func (p *Provider) getEventSelectors(ctx context.Context, trail *types.Trail) ([]EventSelector, error) {
	var eventSelectors []EventSelector
	if trail.HasCustomEventSelectors != nil && *trail.HasCustomEventSelectors {
		output, err := p.client.GetEventSelectors(ctx, &cloudtrail.GetEventSelectorsInput{TrailName: trail.Name})
		if err != nil {
			return []EventSelector{}, err
		}

		for _, eventSelector := range output.EventSelectors {
			var resources []DataResource
			for _, dataResource := range eventSelector.DataResources {
				var values []string
				for _, value := range dataResource.Values {
					values = append(values, value)
				}

				resources = append(resources, DataResource{
					Type:   getValue(dataResource.Type),
					Values: values,
				})
			}

			eventSelectors = append(eventSelectors, EventSelector{
				DataResources: resources,
				ReadWriteType: eventSelector.ReadWriteType,
			})
		}
	}

	return eventSelectors, nil
}

func getValue[T any](ptr *T) T {
	var initVal T
	if ptr != nil {
		return *ptr
	}
	return initVal
}
