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
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var onlyDefaultRegion = []string{awslib.DefaultRegion}

func TestProvider_DescribeNetworkAcl(t *testing.T) {
	tests := []struct {
		name            string
		client          func() Client
		expectedResults int
		wantErr         bool
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeNetworkAcls", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
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
			regions:         onlyDefaultRegion,
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     logp.NewLogger(tt.name),
				clients: clients,
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
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeSecurityGroups", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			regions: onlyDefaultRegion,
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
			regions:         onlyDefaultRegion,
			expectedResults: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     logp.NewLogger(tt.name),
				clients: clients,
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
		regions         []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			regions: onlyDefaultRegion,
			wantErr: true,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("DescribeVpcs", mock.Anything, mock.Anything).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{{VpcId: aws.String("vpc-123")}},
				}, nil)

				m.On("DescribeFlowLogs", mock.Anything, mock.Anything).Return(&ec2.DescribeFlowLogsOutput{
					FlowLogs: []types.FlowLog{{FlowLogId: aws.String("fl-123")}},
				}, nil)

				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:     logp.NewLogger(tt.name),
				clients: clients,
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

func TestProvider_GetEbsEncryptionByDefault(t *testing.T) {
	tests := []struct {
		name    string
		client  func() Client
		want    []awslib.AwsResource
		wantErr bool
		regions []string
	}{
		{
			name: "with error",
			client: func() Client {
				m := &MockClient{}
				m.On("GetEbsEncryptionByDefault", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to get ebs"))
				return m
			},
			regions: onlyDefaultRegion,
			wantErr: true,
		},
		{
			name: "with resource",
			client: func() Client {
				m := &MockClient{}
				m.On("GetEbsEncryptionByDefault", mock.Anything, mock.Anything).Return(&ec2.GetEbsEncryptionByDefaultOutput{
					EbsEncryptionByDefault: aws.Bool(true),
				}, nil)
				return m
			},
			want: []awslib.AwsResource{
				&EBSEncryption{
					Enabled:    true,
					region:     awslib.DefaultRegion,
					awsAccount: "aws-account",
				},
			},
			regions: onlyDefaultRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clients := map[string]Client{}
			for _, r := range tt.regions {
				clients[r] = tt.client()
			}
			p := &Provider{
				log:          logp.NewLogger(tt.name),
				clients:      clients,
				awsAccountID: "aws-account",
			}
			got, err := p.GetEbsEncryptionByDefault(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_DescribeInstances(t *testing.T) {
	type fields struct {
		log          *logp.Logger
		clients      map[string]Client
		awsAccountID string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Instance
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:          tt.fields.log,
				clients:      tt.fields.clients,
				awsAccountID: tt.fields.awsAccountID,
			}
			got, err := p.DescribeInstances(tt.args.ctx, "us-east-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.DescribeInstances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.DescribeInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_CreateSnapshots(t *testing.T) {
	type fields struct {
		log          *logp.Logger
		clients      map[string]Client
		awsAccountID string
	}
	type args struct {
		ctx context.Context
		ins types.Instance
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.SnapshotInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:          tt.fields.log,
				clients:      tt.fields.clients,
				awsAccountID: tt.fields.awsAccountID,
			}
			got, err := p.CreateSnapshots(tt.args.ctx, tt.args.ins, "us-east-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.CreateSnapshots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.CreateSnapshots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_DescribeSnapshots(t *testing.T) {
	type fields struct {
		log          *logp.Logger
		clients      map[string]Client
		awsAccountID string
	}
	type args struct {
		ctx  context.Context
		snap types.SnapshotInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Snapshot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:          tt.fields.log,
				clients:      tt.fields.clients,
				awsAccountID: tt.fields.awsAccountID,
			}
			got, err := p.DescribeSnapshots(tt.args.ctx, tt.args.snap, "us-east-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.DescribeSnapshots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.DescribeSnapshots() = %v, want %v", got, tt.want)
			}
		})
	}
}
