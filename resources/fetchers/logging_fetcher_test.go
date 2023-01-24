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
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoggingFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name              string
		loggingProvider   func() logging.Client
		wantErr           bool
		expectedResources int
	}{
		{
			name: "no resources found",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return([]awslib.AwsResource{}, nil)
				return &m
			},
			wantErr:           false,
			expectedResources: 0,
		},
		{
			name: "with error to describe trails",
			loggingProvider: func() logging.Client {
				m := logging.MockClient{}
				m.On("DescribeTrails", mock.Anything).Return(nil, errors.New("failed to get trails"))
				return &m
			},
			wantErr:           true,
			expectedResources: 0,
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
			wantErr:           false,
			expectedResources: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetching.ResourceInfo, 100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			f := LoggingFetcher{
				log:        logp.NewLogger(tt.name),
				provider:   tt.loggingProvider(),
				cfg:        fetching.AwsBaseFetcherConfig{},
				resourceCh: ch,
			}

			err := f.Fetch(ctx, fetching.CycleMetadata{})
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

func TestEnrichedTrailResource_GetMetadata(t *testing.T) {
	r := LoggingResource{
		AwsResource: logging.EnrichedTrail{
			TrailInfo: cloudtrail.TrailInfo{
				TrailARN: "test-arn",
			},
		},
	}

	meta, err := r.GetMetadata()

	assert.NoError(t, err)
	assert.Equal(t, fetching.ResourceMetadata{ID: "test-arn", Type: "cloud-audit", SubType: "aws-trail", Name: "", ECSFormat: ""}, meta)
	assert.Equal(t, logging.EnrichedTrail{TrailInfo: cloudtrail.TrailInfo{TrailARN: "test-arn"}}, r.GetData())
	assert.Equal(t, nil, r.GetElasticCommonData())
}
