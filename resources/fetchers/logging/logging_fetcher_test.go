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

package logging

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoggingFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name                  string
		loggingProvider       func() logging.Client
		configServiceProvider func() configservice.ConfigService
		expectedResources     int
	}{
		{
			name: "no resources found",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return([]awslib.AwsResource{}, nil)
				return &m
			},
			configServiceProvider: func() configservice.ConfigService {
				m := configservice.MockConfigService{}
				m.On("DescribeConfigRecorders", mock.Anything).Return([]awslib.AwsResource{}, nil)
				return &m
			},
			expectedResources: 0,
		},
		{
			name: "with error to describe trails",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return(nil, errors.New("failed to get trails"))
				return &m
			},
			configServiceProvider: func() configservice.ConfigService {
				m := configservice.MockConfigService{}
				m.On("DescribeConfigRecorders", mock.Anything).Return([]awslib.AwsResource{
					configservice.Config{},
				}, nil)
				return &m
			},
			expectedResources: 1,
		},
		{
			name: "with error to describe config recorders",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return([]awslib.AwsResource{&logging.EnrichedTrail{}}, nil)
				return &m
			},
			configServiceProvider: func() configservice.ConfigService {
				m := configservice.MockConfigService{}
				m.On("DescribeConfigRecorders", mock.Anything).Return(nil, errors.New("failed to get config recorders"))
				return &m
			},
			expectedResources: 1,
		},
		{
			name: "with resources",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return([]awslib.AwsResource{
					&logging.EnrichedTrail{},
					&logging.EnrichedTrail{},
				}, nil)
				return &m
			},
			configServiceProvider: func() configservice.ConfigService {
				m := configservice.MockConfigService{}
				m.On("DescribeConfigRecorders", mock.Anything).Return([]awslib.AwsResource{
					configservice.Config{},
				}, nil)
				return &m
			},
			expectedResources: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetching.ResourceInfo, 100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			f := New(
				WithLogger(logp.NewLogger(tt.name)),
				WithLoggingProvider(tt.loggingProvider()),
				WithConfigserviceProvider(tt.configServiceProvider()),
				WithResourceChan(ch),
				WithConfig(&config.Config{
					Fetchers: []*agentconfig.C{
						agentconfig.MustNewConfigFrom(mapstr.M{
							"name": "aws-trail",
						}),
					},
				}),
			)

			err := f.Fetch(ctx, fetching.CycleMetadata{})
			resources := testhelper.CollectResources(ch)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResources, len(resources))
		})
	}
}

func TestEnrichedTrailResource_GetMetadata(t *testing.T) {
	r := LoggingResource{
		AwsResource: logging.EnrichedTrail{
			TrailInfo: cloudtrail.TrailInfo{
				Trail: types.Trail{
					TrailARN: aws.String("test-arn"),
				},
			},
		},
	}

	meta, err := r.GetMetadata()

	assert.NoError(t, err)
	assert.Equal(t, fetching.ResourceMetadata{ID: "test-arn", Type: "cloud-audit", SubType: "aws-trail", Name: "", ECSFormat: ""}, meta)
	assert.Equal(t, logging.EnrichedTrail{TrailInfo: cloudtrail.TrailInfo{Trail: types.Trail{
		TrailARN: aws.String("test-arn"),
	}}}, r.GetData())
	assert.Equal(t, nil, r.GetElasticCommonData())
}
