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

package awslib

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

const (
	euRegion = "eu-west-1"
	usRegion = "us-east-1"
	afRegion = "af-north-1"
)

var successfulOutput = []string{
	usRegion,
	euRegion,
}

func TestMultiRegionWrapper_NewMultiRegionClients(t *testing.T) {
	type args struct {
		selector func() RegionsSelector
		cfg      aws.Config
		log      *clog.Logger
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Error - return no regions",
			args: args{
				selector: func() RegionsSelector {
					m := &MockRegionsSelector{}
					m.On("Regions", mock.Anything, mock.Anything).Return(nil, errors.New("fail to query endpoint"))
					return m
				},
				cfg: aws.Config{
					Region: afRegion,
				},
				log: testhelper.NewLogger(t),
			},
			want: map[string]string{afRegion: afRegion},
		},
		{
			name: "Should return enabled regions",
			args: args{
				selector: func() RegionsSelector {
					m := &MockRegionsSelector{}
					m.On("Regions", mock.Anything, mock.Anything).Return(successfulOutput, nil)
					return m
				},
				cfg: aws.Config{},
				log: testhelper.NewLogger(t),
			},
			want: map[string]string{DefaultRegion: DefaultRegion, euRegion: euRegion},
		},
	}

	wrapper := MultiRegionClientFactory[string]{}
	for _, tt := range tests {
		factory := func(cfg aws.Config) string {
			return cfg.Region
		}

		t.Run(tt.name, func(t *testing.T) {
			multiRegionClients := wrapper.NewMultiRegionClients(t.Context(), tt.args.selector(), tt.args.cfg, factory, tt.args.log)
			clients := multiRegionClients.GetMultiRegionsClientMap()
			if !reflect.DeepEqual(clients, tt.want) {
				t.Errorf("GetRegions() got = %v, want %v", clients, tt.want)
			}
		})
	}
}

func TestMultiRegionFetch(t *testing.T) {
	type testCase[T any] struct {
		name    string
		clients map[string]testClient
		fetcher func(context.Context, string, T) ([]AwsResource, error)
		want    []AwsResource
		wantErr bool
	}
	tests := []testCase[testClient]{
		{
			name:    "Fetch resources from multiple regions",
			clients: map[string]testClient{euRegion: &dummyTester{euRegion}, usRegion: &dummyTester{usRegion}},
			fetcher: func(_ context.Context, _ string, c testClient) ([]AwsResource, error) {
				return c.DummyFunc()
			},
			want:    []AwsResource{testAwsResource{resRegion: usRegion}, testAwsResource{resRegion: euRegion}},
			wantErr: false,
		},
		{
			name:    "Error from a single region",
			clients: map[string]testClient{afRegion: &dummyTester{afRegion}, usRegion: &dummyTester{usRegion}},
			fetcher: func(_ context.Context, _ string, c testClient) ([]AwsResource, error) {
				return c.DummyFunc()
			},
			want:    []AwsResource{testAwsResource{resRegion: usRegion}},
			wantErr: true,
		},
		{
			name:    "Error from all regions",
			clients: map[string]testClient{afRegion: &dummyTester{afRegion}, usRegion: &dummyTester{usRegion}},
			fetcher: func(_ context.Context, _ string, _ testClient) ([]AwsResource, error) {
				return nil, errors.New("service unavailable")
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MultiRegionFetch(t.Context(), tt.clients, tt.fetcher)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if !assert.ElementsMatch(t, lo.Flatten(got), tt.want) {
				t.Errorf("Fetch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type dummyTester struct {
	region string
}
type testClient interface {
	DummyFunc() ([]AwsResource, error)
}

type testAwsResource struct {
	resRegion string
}

func (t testAwsResource) GetResourceArn() string {
	return t.resRegion
}

func (t testAwsResource) GetResourceName() string { return "" }

func (t testAwsResource) GetResourceType() string { return "" }

func (t testAwsResource) GetRegion() string { return "" }

func (d dummyTester) DummyFunc() ([]AwsResource, error) {
	awsRes := []AwsResource{testAwsResource{resRegion: d.region}}
	switch d.region {
	case euRegion:
		return awsRes, nil
	case usRegion:
		return awsRes, nil
	case afRegion:
		return nil, errors.New("api error")
	default:
		return nil, nil
	}
}

func Test_shouldDrop(t *testing.T) {
	var s1 *string
	assert.True(t, shouldDrop(s1))
	assert.True(t, shouldDrop(func() *string { return nil }()))
	assert.True(t, shouldDrop(nil))

	assert.False(t, shouldDrop([]string{}))
	var s2 []string
	assert.False(t, shouldDrop(s2))
	assert.False(t, shouldDrop(""))
	assert.False(t, shouldDrop(aws.String("")))

	type test struct{}
	assert.False(t, shouldDrop(&test{}))
	assert.False(t, shouldDrop(test{}))
	//
}
