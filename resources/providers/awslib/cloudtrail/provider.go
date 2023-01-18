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

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	log          *logp.Logger
	client       Client
	awsAccountID string
	awsRegion    string
}

type Client interface {
	DescribeTrails(ctx context.Context, params *cloudtrail.DescribeTrailsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error)
	GetTrailStatus(ctx context.Context, params *cloudtrail.GetTrailStatusInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetTrailStatusOutput, error)
	GetEventSelectors(ctx context.Context, params *cloudtrail.GetEventSelectorsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetEventSelectorsOutput, error)
}

func (p *Provider) DescribeCloudTrails(ctx context.Context) ([]awslib.AwsResource, error) {
	input := cloudtrail.DescribeTrailsInput{}
	output, err := p.client.DescribeTrails(ctx, &input)
	if err != nil {
		return nil, err
	}
	result := []awslib.AwsResource{}
	for _, trail := range output.TrailList {
		if trail.Name == nil {
			continue
		}
		input := cloudtrail.GetTrailStatusInput{
			Name: trail.Name,
		}
		status, err := p.client.GetTrailStatus(ctx, &input)
		if err != nil {
			p.log.Errorf("fail to get trail status %s %v", *trail.TrailARN, err.Error())
		}

		selector, err := p.client.GetEventSelectors(ctx, &cloudtrail.GetEventSelectorsInput{
			TrailName: trail.Name,
		})
		if err != nil {
			p.log.Errorf("fail to get trail event selector %s %v", *trail.TrailARN, err.Error())
		}

		result = append(result, TrailInfo{
			trail:         trail,
			status:        status,
			eventSelector: selector,
		})
	}
	return result, nil
}
