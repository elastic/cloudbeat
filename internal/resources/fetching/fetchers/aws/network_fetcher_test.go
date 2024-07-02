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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestNetworkFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name              string
		networkProvider   func() ec2.ElasticCompute
		wantErr           bool
		expectedResources int
	}{
		{
			name: "no resources found",
			networkProvider: func() ec2.ElasticCompute {
				m := ec2.MockElasticCompute{}
				m.On("DescribeNetworkAcl", mock.Anything).Return([]awslib.AwsResource{}, nil)
				m.On("DescribeSecurityGroups", mock.Anything).Return([]awslib.AwsResource{}, nil)
				m.On("DescribeVpcs", mock.Anything).Return([]awslib.AwsResource{}, nil)
				m.On("GetEbsEncryptionByDefault", mock.Anything).Return(nil, nil)
				return &m
			},
		},
		{
			name: "with error to describe nacl",
			networkProvider: func() ec2.ElasticCompute {
				m := ec2.MockElasticCompute{}
				m.On("DescribeNetworkAcl", mock.Anything).Return(nil, errors.New("failed to get nacl"))
				m.On("DescribeSecurityGroups", mock.Anything).Return([]awslib.AwsResource{
					ec2.SecurityGroup{},
					ec2.SecurityGroup{},
				}, nil)
				m.On("DescribeVpcs", mock.Anything).Return([]awslib.AwsResource{ec2.VpcInfo{}}, nil)

				m.On("GetEbsEncryptionByDefault", mock.Anything).Return(nil, nil)
				return &m
			},
			wantErr:           false,
			expectedResources: 3,
		},
		{
			name: "with errors",
			networkProvider: func() ec2.ElasticCompute {
				m := ec2.MockElasticCompute{}
				m.On("DescribeNetworkAcl", mock.Anything).Return([]awslib.AwsResource{
					ec2.NACLInfo{},
					ec2.NACLInfo{},
				}, nil)
				m.On("DescribeSecurityGroups", mock.Anything).Return(nil, errors.New("failed to get security groups"))
				m.On("DescribeVpcs", mock.Anything).Return([]awslib.AwsResource{ec2.VpcInfo{}}, nil)
				m.On("GetEbsEncryptionByDefault", mock.Anything).Return(nil, errors.New("failed to get GetEbsEncryptionByDefault"))
				return &m
			},
			wantErr:           false,
			expectedResources: 3,
		},
		{
			name: "with error to describe VPCs",
			networkProvider: func() ec2.ElasticCompute {
				m := ec2.MockElasticCompute{}
				m.On("DescribeNetworkAcl", mock.Anything).Return([]awslib.AwsResource{
					ec2.NACLInfo{},
					ec2.NACLInfo{},
				}, nil)
				m.On("DescribeSecurityGroups", mock.Anything).Return([]awslib.AwsResource{ec2.SecurityGroup{}}, nil)
				m.On("DescribeVpcs", mock.Anything).Return(nil, errors.New("failed to get VPCs"))
				m.On("GetEbsEncryptionByDefault", mock.Anything).Return(nil, errors.New("failed to get GetEbsEncryptionByDefault"))
				return &m
			},
			wantErr:           false,
			expectedResources: 3,
		},
		{
			name: "with resources",
			networkProvider: func() ec2.ElasticCompute {
				m := ec2.MockElasticCompute{}
				m.On("DescribeNetworkAcl", mock.Anything).Return([]awslib.AwsResource{
					ec2.NACLInfo{},
					ec2.NACLInfo{},
				}, nil)
				m.On("DescribeSecurityGroups", mock.Anything).Return([]awslib.AwsResource{
					ec2.SecurityGroup{},
					ec2.SecurityGroup{},
				}, nil)
				m.On("DescribeVpcs", mock.Anything).Return([]awslib.AwsResource{
					ec2.VpcInfo{},
					ec2.VpcInfo{},
				}, nil)
				m.On("GetEbsEncryptionByDefault", mock.Anything).Return([]awslib.AwsResource{
					ec2.EBSEncryption{},
				}, nil)
				return &m
			},
			wantErr:           false,
			expectedResources: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetching.ResourceInfo, 100)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			f := NetworkFetcher{
				log:        testhelper.NewLogger(t),
				ec2Client:  tt.networkProvider(),
				resourceCh: ch,
			}

			err := f.Fetch(ctx, cycle.Metadata{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			resources := testhelper.CollectResources(ch)
			require.NoError(t, err)
			assert.Len(t, resources, tt.expectedResources)
		})
	}
}

func TestACLResource_GetMetadata(t *testing.T) {
	r := NetworkResource{
		AwsResource: ec2.NACLInfo{},
	}
	meta, err := r.GetMetadata()
	require.NoError(t, err)
	assert.Equal(t, fetching.ResourceMetadata{ID: "", Type: "cloud-compute", SubType: "aws-nacl", Name: ""}, meta)
	assert.Equal(t, ec2.NACLInfo{}, r.GetData())
	m, err := r.GetElasticCommonData()
	require.NoError(t, err)
	assert.Len(t, m, 1)
	assert.Contains(t, m, "cloud.service.name")
}
