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

package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

var onlyDefaultRegion = []string{awslib.DefaultRegion}

func TestProvider_ListFunctions_and_ListAliases(t *testing.T) {
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
				m.On("ListFunctions", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with ListAliases error",
			client: func() Client {
				m := &MockClient{}
				m.On("ListAliases", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				m.On("ListFunctions", mock.Anything, mock.Anything).Return(&lambda.ListFunctionsOutput{}, nil)
				return m
			},
			wantErr:         false,
			regions:         onlyDefaultRegion,
			expectedResults: 0,
		},
		{
			name: "with resources and NO aliases",
			client: func() Client {
				m := &MockClient{}
				m.On("ListAliases", mock.Anything, mock.Anything).Return(&lambda.ListAliasesOutput{}, nil)
				m.On("ListFunctions", mock.Anything, mock.Anything).
					Return(&lambda.ListFunctionsOutput{
						Functions: []types.FunctionConfiguration{
							{
								FunctionName: pointers.Ref("kuba-test-func"),
								FunctionArn:  pointers.Ref("arn:aws:lambda:us-east-1:378890115541:function:kuba-test-func"),
								Runtime:      types.RuntimePython310,
								Role:         pointers.Ref("arn:aws:iam::378890115541:role/service-role/kuba-test-func-role-67nk11yy"),
								Handler:      pointers.Ref("lambda_function.lambda_handler"),
								CodeSize:     int64(440),
								Description:  pointers.Ref("A starter AWS Lambda function."),
								Timeout:      pointers.Ref(int32(3)),
								MemorySize:   pointers.Ref(int32(128)),
								LastModified: pointers.Ref("2024-06-13T11:31:20.250+0000"),
								CodeSha256:   pointers.Ref("JvD8E0a5DGJkAGOHZOinNAMnz8rwSCBvYz4EYaOA0k4="),
								Version:      pointers.Ref("$LATEST"),
							},
						},
					}, nil)
				return m
			},
			regions:         onlyDefaultRegion,
			expectedResults: 1,
		},
		{
			name: "with resources and aliases",
			client: func() Client {
				m := &MockClient{}
				m.On("ListAliases", mock.Anything, mock.Anything).Return(&lambda.ListAliasesOutput{
					Aliases: []types.AliasConfiguration{
						{
							AliasArn:        pointers.Ref("arn:aws:...:alias"),
							Description:     pointers.Ref("this is a description"),
							FunctionVersion: pointers.Ref("$LATEST"),
							Name:            pointers.Ref("Alias name"),
						},
					},
				}, nil)
				m.On("ListFunctions", mock.Anything, mock.Anything).
					Return(&lambda.ListFunctionsOutput{
						Functions: []types.FunctionConfiguration{
							{
								FunctionName: pointers.Ref("kuba-test-func"),
								FunctionArn:  pointers.Ref("arn:aws:lambda:us-east-1:378890115541:function:kuba-test-func"),
								Runtime:      types.RuntimePython310,
								Role:         pointers.Ref("arn:aws:iam::378890115541:role/service-role/kuba-test-func-role-67nk11yy"),
								Handler:      pointers.Ref("lambda_function.lambda_handler"),
								CodeSize:     int64(440),
								Description:  pointers.Ref("A starter AWS Lambda function."),
								Timeout:      pointers.Ref(int32(3)),
								MemorySize:   pointers.Ref(int32(128)),
								LastModified: pointers.Ref("2024-06-13T11:31:20.250+0000"),
								CodeSha256:   pointers.Ref("JvD8E0a5DGJkAGOHZOinNAMnz8rwSCBvYz4EYaOA0k4="),
								Version:      pointers.Ref("$LATEST"),
							},
						},
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
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.ListFunctions(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_ListLayers(t *testing.T) {
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
				m.On("ListLayers", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("ListLayers", mock.Anything, mock.Anything).
					Return(&lambda.ListLayersOutput{
						Layers: []types.LayersListItem{
							{
								LatestMatchingVersion: &types.LayerVersionsListItem{
									LayerVersionArn: pointers.Ref("arn:aws:lambda:us-east-1:378890115541:layer:kuba-jq:1"),
									Version:         int64(1),
									CreatedDate:     pointers.Ref("2024-06-13T11:34:25.613+0000"),
									CompatibleArchitectures: []types.Architecture{
										types.ArchitectureArm64,
									},
								},
								LayerArn:  pointers.Ref("arn:aws:lambda:us-east-1:378890115541:layer:kuba-jq"),
								LayerName: pointers.Ref("kuba-jq"),
							},
						},
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
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.ListLayers(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}

func TestProvider_ListEventSourceMappings(t *testing.T) {
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
				m.On("ListEventSourceMappings", mock.Anything, mock.Anything).Return(nil, errors.New("failed"))
				return m
			},
			wantErr: true,
			regions: onlyDefaultRegion,
		},
		{
			name: "with resources",
			client: func() Client {
				m := &MockClient{}
				m.On("ListEventSourceMappings", mock.Anything, mock.Anything).
					Return(&lambda.ListEventSourceMappingsOutput{
						EventSourceMappings: []types.EventSourceMappingConfiguration{
							{
								UUID:           pointers.Ref("a-b-c-d"),
								FunctionArn:    pointers.Ref("arn:aws:lambda:us-east-1:378890115541:function:kuba-test-func"),
								EventSourceArn: pointers.Ref("arn:aws:lambda:us-east-1:378890115541:function:kuba-test-func"),
								State:          pointers.Ref("Enabling"),
							},
						},
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
				log:     testhelper.NewLogger(t),
				clients: clients,
			}
			got, err := p.ListEventSourceMappings(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.expectedResults)
		})
	}
}
