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

package iam

import (
	"errors"
	"slices"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestProvider_GetAccessAnalyzers(t *testing.T) {
	tests := []struct {
		name    string
		clients map[string]AccessAnalyzerClient
		want    awslib.AwsResource
		wantErr string
	}{
		{
			name:    "Clients not initialized",
			clients: nil,
			wantErr: "multi region clients have not been initialized",
		},
		{
			name: "Error in client",
			clients: map[string]AccessAnalyzerClient{
				"region-1": mockAccessAnalyzerWithError(),
			},
			wantErr: "some error",
		},
		{
			name: "Ok",
			clients: map[string]AccessAnalyzerClient{
				"region-1": mockAccessAnalyzerWithArns("some-arn", "some-other-arn"),
				"region-2": mockAccessAnalyzerWithArns("some-third-arn"),
			},
			want: AccessAnalyzers{
				Analyzers: []AccessAnalyzer{
					{
						AnalyzerSummary: types.AnalyzerSummary{Arn: aws.String("some-arn")},
						Region:          "region-1",
					},
					{
						AnalyzerSummary: types.AnalyzerSummary{Arn: aws.String("some-other-arn")},
						Region:          "region-1",
					},
					{
						AnalyzerSummary: types.AnalyzerSummary{Arn: aws.String("some-third-arn")},
						Region:          "region-2",
					},
				},
				Regions: []string{"region-1", "region-2"},
			},
		},
		{
			name: "With empty regions",
			clients: map[string]AccessAnalyzerClient{
				"region-1": mockAccessAnalyzerWithArns(),
				"region-2": mockAccessAnalyzerWithArns("some-arn"),
				"region-3": mockAccessAnalyzerWithArns(),
			},
			want: AccessAnalyzers{
				Analyzers: []AccessAnalyzer{
					{
						AnalyzerSummary: types.AnalyzerSummary{Arn: aws.String("some-arn")},
						Region:          "region-2",
					},
				},
				Regions: []string{"region-1", "region-2", "region-3"},
			},
		},
		{
			name: "Ok and error => error",
			clients: map[string]AccessAnalyzerClient{
				"region-1": mockAccessAnalyzerWithArns("whatever"),
				"region-2": mockAccessAnalyzerWithError(),
			},
			wantErr: "some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:                   testhelper.NewLogger(t),
				client:                nil,
				accessAnalyzerClients: tt.clients,
			}

			allAnalyzers, err := p.GetAccessAnalyzers(t.Context())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			allAnalyzersTyped, ok := allAnalyzers.(AccessAnalyzers)
			require.True(t, ok)
			sort.Slice(allAnalyzersTyped.Analyzers, func(i, j int) bool {
				a := allAnalyzersTyped.Analyzers[i]
				b := allAnalyzersTyped.Analyzers[j]
				if a.Region != b.Region {
					return a.Region < b.Region
				}
				return *a.Arn < *b.Arn
			})
			slices.Sort(allAnalyzersTyped.Regions)

			assert.Equal(t, tt.want, allAnalyzers)
		})
	}
}

func mockAccessAnalyzerWithError() *MockAccessAnalyzerClient {
	client := &MockAccessAnalyzerClient{}
	client.On("ListAnalyzers", mock.Anything, mock.Anything).Return(nil, errors.New("some error")).Once()
	return client
}

func mockAccessAnalyzerWithArns(arns ...string) *MockAccessAnalyzerClient {
	var output []types.AnalyzerSummary
	for _, arn := range arns {
		output = append(output, types.AnalyzerSummary{Arn: aws.String(arn)})
	}

	client := &MockAccessAnalyzerClient{}
	client.On("ListAnalyzers", mock.Anything, mock.Anything).Return(&accessanalyzer.ListAnalyzersOutput{
		Analyzers: output,
	}, nil).Once()
	return client
}
