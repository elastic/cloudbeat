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

package ec2

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestProvider_DescribeNetworkAcl(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkAcls", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkAcls", mock.Anything, mock.Anything).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{},
						{},
					},
				}, nil)
				return m
			},
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:    logp.NewLogger(tt.name),
				client: tt.client(),
			}
			got, err := p.DescribeNetworkAcl(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResults, len(got))
		})
	}
}

func TestProvider_DescribeSecurityGroups(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{},
						{},
					},
				}, nil)
				return m
			},
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:    logp.NewLogger(tt.name),
				client: tt.client(),
			}

			got, err := p.DescribeSecurityGroups(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResults, len(got))
		})
	}
}

func TestProvider_DescribeVPCs(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{{VpcId: aws.String("vpc-123")}}}, nil)

				m.On("DescribeFlowLogs", mock.Anything, mock.Anything).Return(&ec2.DescribeFlowLogsOutput{
					FlowLogs: []types.FlowLog{{FlowLogId: aws.String("fl-123")}}}, nil)

				return m
			},
			expectedResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:    logp.NewLogger(tt.name),
				client: tt.client(),
			}

			got, err := p.DescribeVPCs(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResults, len(got))
		})
	}
}
