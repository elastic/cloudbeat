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

package benchmark

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/elastic/beats/v7/libbeat/management/status"
	libbeataws "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/errorhandler"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestAWS_Initialize(t *testing.T) {
	testhelper.SkipLong(t)

	tests := []struct {
		name             string
		identityProvider awslib.IdentityProviderGetter
		cfg              config.Config
		want             []string
		wantErr          string
	}{
		{
			name:    "nothing initialized",
			wantErr: "aws identity provider is uninitialized",
		},
		{
			name:             "identity provider error",
			identityProvider: mockAwsIdentityProvider(errors.New("some error")),
			wantErr:          "some error",
		},
		{
			// TODO: this doesn't finish instantly because there is code in MultiRegionClientFactory that is not initialized lazily
			name:             "no error",
			identityProvider: mockAwsIdentityProvider(nil),
			want: []string{
				fetching.IAMType,
				fetching.KmsType,
				fetching.TrailType,
				fetching.AwsMonitoringType,
				fetching.EC2NetworkingType,
				fetching.RdsType,
				fetching.S3Type,
			},
		},
		{
			name: "cloud connectors",
			cfg: config.Config{
				Benchmark: "cis_aws",
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType:     config.SingleAccount,
						Cred:            libbeataws.ConfigAWS{},
						CloudConnectors: true,
						CloudConnectorsConfig: config.CloudConnectorsConfig{
							LocalRoleARN:  "abc123",
							GlobalRoleARN: "abc456",
							ResourceID:    "abc789",
						},
					},
				},
			},
			identityProvider: func() awslib.IdentityProviderGetter {
				cfgMatcher := mock.MatchedBy(func(cfg aws.Config) bool {
					c, is := cfg.Credentials.(*aws.CredentialsCache)
					if !is {
						return false
					}
					return c.IsCredentialsProvider(&stscreds.AssumeRoleProvider{})
				})
				identityProvider := &awslib.MockIdentityProviderGetter{}
				identityProvider.EXPECT().GetIdentity(mock.Anything, cfgMatcher).Return(
					&cloud.Identity{
						Account: "test-account",
					},
					nil,
				)

				return identityProvider
			}(),
			want: []string{
				fetching.IAMType,
				fetching.KmsType,
				fetching.TrailType,
				fetching.AwsMonitoringType,
				fetching.EC2NetworkingType,
				fetching.RdsType,
				fetching.S3Type,
			},
		},
		{
			name: "no credential cache in non cloud connectors setup",
			cfg: config.Config{
				Benchmark: "cis_aws",
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{
						AccountType: config.SingleAccount,
						Cred: libbeataws.ConfigAWS{
							AccessKeyID:     "keyid",
							SecretAccessKey: "key",
						},
						CloudConnectors: false,
					},
				},
			},
			identityProvider: func() awslib.IdentityProviderGetter {
				cfgMatcher := mock.MatchedBy(func(cfg aws.Config) bool {
					_, is := cfg.Credentials.(credentials.StaticCredentialsProvider)
					return is
				})
				identityProvider := &awslib.MockIdentityProviderGetter{}
				identityProvider.EXPECT().GetIdentity(mock.Anything, cfgMatcher).Return(
					&cloud.Identity{
						Account: "test-account",
					},
					nil,
				)

				return identityProvider
			}(),
			want: []string{
				fetching.IAMType,
				fetching.KmsType,
				fetching.TrailType,
				fetching.AwsMonitoringType,
				fetching.EC2NetworkingType,
				fetching.RdsType,
				fetching.S3Type,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &AWS{
				IdentityProvider: tt.identityProvider,
				errorPublisher:   nil,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func TestAWSErrorProcessor(t *testing.T) {
	type expectedCall struct {
		status status.Status
		msg    string
	}

	tests := map[string]struct {
		inputErrors      []error
		expectedMessages []string
	}{
		"no status update": {
			inputErrors:      []error{errors.New("irrelevant")},
			expectedMessages: []string{},
		},
		"status update": {
			inputErrors: []error{&errorhandler.MissingCSPPermissionError{
				Permission: "abc",
			}},
			expectedMessages: []string{
				"missing permission on cloud provider side: abc",
			},
		},
		"status update with inner error": {
			inputErrors: []error{fmt.Errorf("error %w", &errorhandler.MissingCSPPermissionError{
				Permission: "abc",
			})},
			expectedMessages: []string{
				"missing permission on cloud provider side: abc",
			},
		},

		"multiple with appended permissions": {
			inputErrors: []error{
				&errorhandler.MissingCSPPermissionError{Permission: "abc1"},
				&errorhandler.MissingCSPPermissionError{Permission: "abc2"},
				errors.New("irrelevant"),
				&errorhandler.MissingCSPPermissionError{Permission: "abc3"},
			},
			expectedMessages: []string{
				"missing permission on cloud provider side: abc1",
				"missing permission on cloud provider side: abc1 , abc2",
				"missing permission on cloud provider side: abc1 , abc2 , abc3",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := &mockStatusReporter{}
			for _, ec := range tc.expectedMessages {
				m.On("UpdateStatus", status.Degraded, ec).Once()
			}

			defer m.AssertExpectations(t)

			s := NewAWSErrorProcessor(testhelper.NewLogger(t))

			for _, err := range tc.inputErrors {
				s.Process(m, err)
			}
		})
	}

	t.Run("clean", func(t *testing.T) {
		m := &mockStatusReporter{}
		c := make([]*mock.Call, 0)
		c = append(c, m.On("UpdateStatus", status.Degraded, "missing permission on cloud provider side: abc1").Once())
		c = append(c, m.On("UpdateStatus", status.Degraded, "missing permission on cloud provider side: abc1 , abc2").Once())
		c = append(c, m.On("UpdateStatus", status.Degraded, "missing permission on cloud provider side: abc1 , abc2 , abc3").Once())
		c = append(c, m.On("UpdateStatus", status.Degraded, "missing permission on cloud provider side: abc4").Once())
		mock.InOrder(c...)

		defer m.AssertExpectations(t)

		s := NewAWSErrorProcessor(testhelper.NewLogger(t))

		s.Process(m, &errorhandler.MissingCSPPermissionError{Permission: "abc1"})
		s.Process(m, &errorhandler.MissingCSPPermissionError{Permission: "abc2"})
		s.Process(m, &errorhandler.MissingCSPPermissionError{Permission: "abc3"})
		s.Clear()
		s.Process(m, &errorhandler.MissingCSPPermissionError{Permission: "abc4"})
	})
}

type mockStatusReporter struct {
	mock.Mock
}

func (u *mockStatusReporter) UpdateStatus(s status.Status, msg string) {
	u.Called(s, msg)
}
