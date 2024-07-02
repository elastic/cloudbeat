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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Provider struct {
	log     *logp.Logger
	clients map[string]Client
}

type Client interface {
	DescribeTrails(ctx context.Context, params *cloudtrail.DescribeTrailsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error)
	GetTrailStatus(ctx context.Context, params *cloudtrail.GetTrailStatusInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetTrailStatusOutput, error)
	GetEventSelectors(ctx context.Context, params *cloudtrail.GetEventSelectorsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetEventSelectorsOutput, error)
}

func (p Provider) DescribeTrails(ctx context.Context) ([]TrailInfo, error) {
	input := cloudtrail.DescribeTrailsInput{}
	defaultClient, err := awslib.GetDefaultClient(p.clients)
	if err != nil {
		return nil, fmt.Errorf("could not select default region client: %w", err)
	}
	output, err := defaultClient.DescribeTrails(ctx, &input)
	if err != nil {
		return nil, err
	}

	result := make([]TrailInfo, 0, len(output.TrailList))
	for _, trail := range output.TrailList {
		if trail.Name == nil {
			continue
		}
		status, err := p.getTrailStatus(ctx, trail)
		if err != nil {
			p.log.Errorf("failed to get trail status %s %v", *trail.TrailARN, err.Error())
		}

		selectors, err := p.getEventSelectors(ctx, trail)
		if err != nil {
			p.log.Errorf("failed to get trail event selector %s %v", *trail.TrailARN, err.Error())
		}

		result = append(result, TrailInfo{
			Trail:          trail,
			Status:         status,
			EventSelectors: selectors,
		})
	}
	return result, nil
}

func (p Provider) getTrailStatus(ctx context.Context, trail types.Trail) (*cloudtrail.GetTrailStatusOutput, error) {
	client, err := awslib.GetClient(trail.HomeRegion, p.clients)
	if err != nil {
		return nil, err
	}

	return client.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{Name: trail.TrailARN})
}

func (p Provider) getEventSelectors(ctx context.Context, trail types.Trail) ([]types.EventSelector, error) {
	client, err := awslib.GetClient(trail.HomeRegion, p.clients)
	if err != nil {
		return nil, err
	}

	output, err := client.GetEventSelectors(ctx, &cloudtrail.GetEventSelectorsInput{TrailName: trail.TrailARN})
	if err != nil {
		return []types.EventSelector{}, err
	}

	return output.EventSelectors, nil
}
