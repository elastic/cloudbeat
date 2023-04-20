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
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sort"
	"testing"
)

func TestProvider_GetAccessAnalyzers(t *testing.T) {
	tests := []struct {
		name    string
		clients map[string]AccessAnalyzer
		want    []AnalyzersForRegion
		wantErr string
	}{
		{
			name:    "Clients not initialized",
			clients: nil,
			want:    nil,
			wantErr: "multi region clients have not been initialized",
		},
		{
			name: "Error in client",
			clients: map[string]AccessAnalyzer{
				"region-1": mockAccessAnalyzerWithError(),
			},
			wantErr: "some error",
		},
		{
			name: "Ok",
			clients: map[string]AccessAnalyzer{
				"region-1": mockAccessAnalyzerWithArns("some-arn", "zzz-last-arn"),
				"region-2": mockAccessAnalyzerWithArns("some-other-arn"),
			},
			want: []AnalyzersForRegion{
				{
					Analyzers: []types.AnalyzerSummary{
						{
							Arn: aws.String("some-arn"),
						},
						{
							Arn: aws.String("zzz-last-arn"),
						},
					},
					Region: "region-1",
				},
				{
					Analyzers: []types.AnalyzerSummary{
						{
							Arn: aws.String("some-other-arn"),
						},
					},
					Region: "region-2",
				},
			},
		},
		{
			name: "Ok and error => error",
			clients: map[string]AccessAnalyzer{
				"region-1": mockAccessAnalyzerWithArns("whatever"),
				"region-2": mockAccessAnalyzerWithError(),
			},
			want:    nil,
			wantErr: "some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:                   logp.NewLogger("iam-provider"),
				client:                nil,
				accessAnalyzerClients: tt.clients,
			}

			allAnalyzers, err := p.GetAccessAnalyzers(context.Background())
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			// Return order of AnalyzersForRegion is undefined, so sort results first
			sort.Slice(allAnalyzers, func(i, j int) bool {
				return allAnalyzers[i].Region < allAnalyzers[j].Region
			})
			for _, analyzersForRegion := range allAnalyzers {
				sort.Slice(analyzersForRegion.Analyzers, func(i, j int) bool {
					return *analyzersForRegion.Analyzers[i].Arn < *analyzersForRegion.Analyzers[j].Arn
				})
			}

			assert.Equal(t, tt.want, allAnalyzers)
		})
	}
}

func mockAccessAnalyzerWithError() *MockAccessAnalyzer {
	client := &MockAccessAnalyzer{}
	client.On("ListAnalyzers", mock.Anything, mock.Anything).Return(nil, errors.New("some error")).Once()
	return client
}

func mockAccessAnalyzerWithArns(arns ...string) *MockAccessAnalyzer {
	var output []types.AnalyzerSummary
	for _, arn := range arns {
		output = append(output, types.AnalyzerSummary{Arn: aws.String(arn)})
	}

	client := &MockAccessAnalyzer{}
	client.On("ListAnalyzers", mock.Anything, mock.Anything).Return(&accessanalyzer.ListAnalyzersOutput{
		Analyzers: output,
	}, nil).Once()
	return client
}
