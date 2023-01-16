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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	ec2sdk "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

const (
	euRegion = "eu-west-1"
	usRegion = "us-east-1"
	afRegion = "af-north-1"
)

var successfulOutput = &ec2sdk.DescribeRegionsOutput{
	Regions: []types.Region{
		{
			RegionName: awssdk.String(usRegion),
		},
		{
			RegionName: awssdk.String(euRegion),
		},
	},
}

func TestMultiRegionWrapper_NewMultiRegionClients(t *testing.T) {
	type args struct {
		client func() DescribeCloudRegions
		cfg    awssdk.Config
		log    *logp.Logger
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Error - return no regions",
			args: args{
				client: func() DescribeCloudRegions {
					m := &MockDescribeCloudRegions{}
					m.On("DescribeRegions", mock.Anything, mock.Anything).Return(nil, errors.New("fail to query endpoint"))
					return m
				},
				cfg: awssdk.Config{},
				log: logp.NewLogger("multi-region-test"),
			},
			want: map[string]string{DefaultRegion: DefaultRegion},
		},
		{
			name: "Should return enabled regions",
			args: args{
				client: func() DescribeCloudRegions {
					m := &MockDescribeCloudRegions{}
					m.On("DescribeRegions", mock.Anything, mock.Anything).Return(successfulOutput, nil)
					return m
				},
				cfg: awssdk.Config{},
				log: logp.NewLogger("multi-region-test"),
			},
			want: map[string]string{DefaultRegion: DefaultRegion, euRegion: euRegion},
		},
	}

	wrapper := MultiRegionWrapper[string]{}
	for _, tt := range tests {
		factory := func(cfg awssdk.Config) string {
			return cfg.Region
		}

		t.Run(tt.name, func(t *testing.T) {
			multiRegionClients := wrapper.NewMultiRegionClients(tt.args.client(), tt.args.cfg, factory, tt.args.log)
			if !reflect.DeepEqual(multiRegionClients.clients, tt.want) {
				t.Errorf("GetRegions() got = %v, want %v", multiRegionClients.clients, tt.want)
			}
		})
	}
}

func TestMultiRegionWrapper_Fetch(t *testing.T) {
	type args[T any] struct {
		fetcher func(T) ([]AwsResource, error)
	}
	type testCase[T any] struct {
		name    string
		w       MultiRegionWrapper[testClient]
		args    args[testClient]
		want    []AwsResource
		wantErr bool
	}
	tests := []testCase[testClient]{
		{
			name: "Fetch resources from multiple regions",
			w: MultiRegionWrapper[testClient]{
				clients: map[string]testClient{euRegion: &dummyTester{euRegion}, usRegion: &dummyTester{usRegion}},
			},
			args: args[testClient]{
				fetcher: func(c testClient) ([]AwsResource, error) {
					return c.DummyFunc()
				},
			},
			want:    []AwsResource{testAwsResource{resRegion: usRegion}, testAwsResource{resRegion: euRegion}},
			wantErr: false,
		},
		{
			name: "Error from a single region",
			w: MultiRegionWrapper[testClient]{
				clients: map[string]testClient{afRegion: &dummyTester{afRegion}, usRegion: &dummyTester{usRegion}},
			},
			args: args[testClient]{
				fetcher: func(c testClient) ([]AwsResource, error) {
					return c.DummyFunc()
				},
			},
			want:    []AwsResource{testAwsResource{resRegion: usRegion}},
			wantErr: true,
		},
		{
			name: "Error from all regions",
			w: MultiRegionWrapper[testClient]{
				clients: map[string]testClient{afRegion: &dummyTester{afRegion}, usRegion: &dummyTester{usRegion}},
			},
			args: args[testClient]{
				fetcher: func(c testClient) ([]AwsResource, error) {
					return nil, errors.New("service unavailable")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.w.Fetch(tt.args.fetcher)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got, tt.want) {
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

func (d dummyTester) DummyFunc() ([]AwsResource, error) {
	awsRes := []AwsResource{testAwsResource{resRegion: d.region}}
	switch d.region {
	case euRegion:
		return awsRes, nil
	case usRegion:
		return awsRes, nil
	case afRegion:
		return nil, errors.New("api error")
	}

	return nil, nil
}
