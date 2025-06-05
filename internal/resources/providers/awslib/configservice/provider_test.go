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

package configservice

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	configSDK "github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

const awsAccountId = "dummy-account-id"

func TestProvider_DescribeConfigRecorders(t *testing.T) {
	tests := []struct {
		name            string
		mockClient      func() Client
		regions         []string
		wantErr         bool
		expectedResults int
	}{
		{
			name: "Should return a config without recorders",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{}, nil)
				return &m
			},
			regions:         []string{"us-east-1"},
			wantErr:         false,
			expectedResults: 1,
		},
		{
			name: "Should not return a config due to error",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(nil, errors.New("API_ERROR"))
				return &m
			},
			regions:         []string{"us-east-1"},
			wantErr:         true,
			expectedResults: 0,
		},
		{
			name: "Should return config resources",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test2")}}}, nil).Once()

				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test2")}}}, nil).Once()

				return &m
			},
			regions:         []string{"us-east-1", "us-east-2"},
			wantErr:         false,
			expectedResults: 2,
		},
		{
			name: "Should return config resources from a single region",
			mockClient: func() Client {
				m := MockClient{}
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorders", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecordersOutput{
					ConfigurationRecorders: []types.ConfigurationRecorder{{Name: aws.String("test2")}}}, nil).Once()

				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(&configSDK.DescribeConfigurationRecorderStatusOutput{
					ConfigurationRecordersStatus: []types.ConfigurationRecorderStatus{{Name: aws.String("test1")}}}, nil).Once()
				m.On("DescribeConfigurationRecorderStatus", mock.Anything, mock.Anything).Return(nil, errors.New("API_ERROR")).Once()

				return &m
			},
			regions:         []string{"us-east-1", "us-east-2"},
			wantErr:         true,
			expectedResults: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				log:          testhelper.NewLogger(t),
				awsAccountId: awsAccountId,
				clients:      testhelper.CreateMockClients[Client](tt.mockClient(), tt.regions),
			}

			got, err := p.DescribeConfigRecorders(t.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeConfigRecorders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Len(t, got, tt.expectedResults)
		})
	}
}
