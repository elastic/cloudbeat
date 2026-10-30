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
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

var expectedAWSSubtypes = []string{
	"123-" + fetching.IAMType,
	"123-" + fetching.KmsType,
	"123-" + fetching.TrailType,
	"123-" + fetching.AwsMonitoringType,
	"123-" + fetching.EC2NetworkingType,
	"123-" + fetching.RdsType,
	"123-" + fetching.S3Type,
	"456-" + fetching.IAMType,
	"456-" + fetching.KmsType,
	"456-" + fetching.TrailType,
	"456-" + fetching.AwsMonitoringType,
	"456-" + fetching.EC2NetworkingType,
	"456-" + fetching.RdsType,
	"456-" + fetching.S3Type,
}

func TestAWSOrg_Initialize(t *testing.T) {
	testhelper.SkipLong(t)

	tests := []struct {
		name             string
		iamProvider      iam.RoleGetter
		identityProvider awslib.IdentityProviderGetter
		accountProvider  awslib.AccountProviderAPI
		cfg              config.Config
		want             []string
		wantErr          string
	}{
		{
			name:    "nothing initialized",
			wantErr: "aws iam provider is uninitialized",
		},
		{
			name:             "account provider uninitialized",
			iamProvider:      getMockIAMRoleGetter(nil),
			identityProvider: mockAwsIdentityProviderWithCallerIdentityCall(nil),
			accountProvider:  nil,
			wantErr:          "account provider is uninitialized",
		},
		{
			name:             "identity provider error",
			iamProvider:      getMockIAMRoleGetter(nil),
			identityProvider: mockAwsIdentityProviderWithCallerIdentityCall(errors.New("some error")),
			accountProvider:  mockAccountProvider(errors.New("not this error")),
			wantErr:          "some error",
		},
		{
			name:             "account provider error",
			iamProvider:      getMockIAMRoleGetter(nil),
			identityProvider: mockAwsIdentityProviderWithCallerIdentityCall(nil),
			accountProvider:  mockAccountProvider(errors.New("some error")),
			want:             []string{},
		},
		{
			name:             "no error",
			iamProvider:      getMockIAMRoleGetter(nil),
			identityProvider: mockAwsIdentityProviderWithCallerIdentityCall(nil),
			accountProvider:  mockAccountProvider(nil),
			want:             expectedAWSSubtypes,
		},
		{
			name:             "no error, already cloudbeat-root",
			iamProvider:      getMockIAMRoleGetter(nil),
			identityProvider: mockIdentityProviderAlreadyCloudbeatRoot(),
			accountProvider:  mockAccountProvider(nil),
			want:             expectedAWSSubtypes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &AWSOrg{
				IAMProvider:       tt.iamProvider,
				IdentityProvider:  tt.identityProvider,
				AccountProvider:   tt.accountProvider,
				StatusHandler:     statushandler.NewMockStatusHandlerAPI(t),
				AWSCredsValidator: awslib.CredentialsValidatorNOOP,
				RoleNamesProvider: awslib.BenchmarkOrgIAMRoleNamesProvider{},
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func Test_getAwsAccounts(t *testing.T) {
	tests := []struct {
		name            string
		accountProvider awslib.AccountProviderAPI
		rootIdentity    cloud.Identity
		want            []cloud.Identity
		wantErr         string
	}{
		{
			name:            "error",
			accountProvider: mockAccountProvider(errors.New("some error")),
			rootIdentity:    cloud.Identity{Account: "123"},
			wantErr:         "some error",
		},
		{
			name: "success",
			accountProvider: mockAccountProviderWithIdentities([]cloud.Identity{
				{
					Account:      "123",
					AccountAlias: "alias",
				},
				{
					Account:      "456",
					AccountAlias: "alias2",
				},
			}),
			rootIdentity: cloud.Identity{
				Account: "123",
			},
			want: []cloud.Identity{
				{
					Account:      "123",
					AccountAlias: "alias",
				},
				{
					Account:      "456",
					AccountAlias: "alias2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AWSOrg{
				IAMProvider:       getMockIAMRoleGetter([]iam.Role{*makeRole("cloudbeat-root")}),
				IdentityProvider:  nil,
				AccountProvider:   tt.accountProvider,
				StatusHandler:     statushandler.NewMockStatusHandlerAPI(t),
				AWSCredsValidator: awslib.CredentialsValidatorNOOP,
				RoleNamesProvider: awslib.BenchmarkOrgIAMRoleNamesProvider{},
			}
			log := testhelper.NewLogger(t)
			got, err := a.getAwsAccounts(t.Context(), log, aws.Config{}, &tt.rootIdentity)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.want))

			for i, account := range got {
				assert.Equal(t, tt.want[i], account.Identity)
				// if the account is other than the management, a credential cache has been created (assuming the member role)
				if account.Account != tt.rootIdentity.Account {
					assert.IsType(t, &aws.CredentialsCache{}, account.Credentials)
				}
			}
		})
	}
}

func Test_getAwsAccounts_Partition(t *testing.T) {
	tests := []struct {
		name         string
		partition    string
		wantRoleARNs []string
	}{
		{
			name:      "commercial",
			partition: awslib.DefaultPartition,
			wantRoleARNs: []string{
				"arn:aws:iam::123:role/cloudbeat-securityaudit",
				"arn:aws:iam::456:role/cloudbeat-securityaudit",
			},
		},
		{
			name:      "unknown partition uses commercial",
			partition: "",
			wantRoleARNs: []string{
				"arn:aws:iam::123:role/cloudbeat-securityaudit",
				"arn:aws:iam::456:role/cloudbeat-securityaudit",
			},
		},
		{
			name:      "govcloud",
			partition: awslib.GovCloudPartition,
			wantRoleARNs: []string{
				"arn:aws-us-gov:iam::123:role/cloudbeat-securityaudit",
				"arn:aws-us-gov:iam::456:role/cloudbeat-securityaudit",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AWSOrg{
				// Cloudbeat scans the management account with the member role. Thus, both accounts make an AssumeRole call.
				IAMProvider: getMockIAMRoleGetter([]iam.Role{
					*makeRole("cloudbeat-root", scanSettingTagKey, scanSettingTagValue),
					*makeRole("cloudbeat-securityaudit"),
				}),
				AccountProvider: mockAccountProviderWithIdentities([]cloud.Identity{
					{Account: "123", AccountAlias: "management", Partition: tt.partition},
					{Account: "456", AccountAlias: "member", Partition: tt.partition},
				}),
				IdentityProvider:  nil,
				StatusHandler:     statushandler.NewMockStatusHandlerAPI(t),
				AWSCredsValidator: awslib.CredentialsValidatorNOOP,
				RoleNamesProvider: awslib.BenchmarkOrgIAMRoleNamesProvider{},
			}
			rootIdentity := cloud.Identity{Account: "123"}
			recorder := &assumeRoleRecorder{mu: sync.Mutex{}, arns: nil}

			got, err := a.getAwsAccounts(t.Context(), testhelper.NewLogger(t), recorder.config(), &rootIdentity)
			require.NoError(t, err)
			require.Len(t, got, 2)

			for _, account := range got {
				assert.Equal(t, tt.partition, account.Identity.Partition)
				// The credentials provider makes the AssumeRole call only when you get the credentials.
				_, err := account.Credentials.Retrieve(t.Context())
				require.NoError(t, err)
			}
			assert.ElementsMatch(t, tt.wantRoleARNs, recorder.roleARNs())
		})
	}
}

func Test_cloudbeatRootRoleARN(t *testing.T) {
	tests := []struct {
		name      string
		callerArn *string
		want      string
	}{
		{
			name:      "commercial",
			callerArn: pointers.Ref("arn:aws:sts::111122223333:assumed-role/some-role/session"),
			want:      "arn:aws:iam::111122223333:role/cloudbeat-root",
		},
		{
			name:      "govcloud",
			callerArn: pointers.Ref("arn:aws-us-gov:sts::111122223333:assumed-role/some-role/session"),
			want:      "arn:aws-us-gov:iam::111122223333:role/cloudbeat-root",
		},
		{
			name:      "govcloud iam user",
			callerArn: pointers.Ref("arn:aws-us-gov:iam::111122223333:user/cloudbeat"),
			want:      "arn:aws-us-gov:iam::111122223333:role/cloudbeat-root",
		},
		{
			name:      "no caller arn uses commercial",
			callerArn: nil,
			want:      "arn:aws:iam::111122223333:role/cloudbeat-root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AWSOrg{
				IAMProvider:       nil,
				IdentityProvider:  nil,
				AccountProvider:   nil,
				StatusHandler:     nil,
				AWSCredsValidator: nil,
				RoleNamesProvider: awslib.BenchmarkOrgIAMRoleNamesProvider{},
			}
			got := a.cloudbeatRootRoleARN(sts.GetCallerIdentityOutput{
				Account: pointers.Ref("111122223333"),
				Arn:     tt.callerArn,
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// assumeRoleRecorder records the RoleArn of each STS AssumeRole call and returns test credentials.
// It does not send requests to AWS.
type assumeRoleRecorder struct {
	mu   sync.Mutex
	arns []string
}

func (r *assumeRoleRecorder) roleARNs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.arns...)
}

func (r *assumeRoleRecorder) config() aws.Config {
	record := middleware.InitializeMiddlewareFunc("assumeRoleRecorder",
		func(_ context.Context, in middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
			input, ok := in.Parameters.(*sts.AssumeRoleInput)
			if !ok {
				return middleware.InitializeOutput{}, middleware.Metadata{}, fmt.Errorf("unexpected AWS call %T", in.Parameters)
			}

			r.mu.Lock()
			r.arns = append(r.arns, pointers.Deref(input.RoleArn))
			r.mu.Unlock()

			return middleware.InitializeOutput{Result: &sts.AssumeRoleOutput{
				AssumedRoleUser: &ststypes.AssumedRoleUser{
					Arn:           pointers.Ref("arn:aws:sts::123:assumed-role/test/session"),
					AssumedRoleId: pointers.Ref("AROATEST:session"),
				},
				Credentials: &ststypes.Credentials{
					AccessKeyId:     pointers.Ref("ASIATEST"),
					SecretAccessKey: pointers.Ref("secret"),
					SessionToken:    pointers.Ref("token"),
					Expiration:      pointers.Ref(time.Now().Add(time.Hour)),
				},
			}}, middleware.Metadata{}, nil
		})

	return aws.Config{
		Region:      awslib.DefaultRegion,
		Credentials: aws.AnonymousCredentials{},
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(record, middleware.Before)
			},
		},
	}
}

type mockStsClient struct{}

func (c *mockStsClient) AssumeRole(_ context.Context, _ *sts.AssumeRoleInput, _ ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return &sts.AssumeRoleOutput{}, nil
}

func Test_pickManagementAccountRole(t *testing.T) {
	tests := []struct {
		name                 string
		roles                []iam.Role
		expectedLog          string
		expectedErrorMessage string
	}{
		{
			name:        "success: cloudbeat-root is not tagged (backward compatibility)",
			roles:       []iam.Role{*makeRole("cloudbeat-root")},
			expectedLog: "using 'cloudbeat-root' role for backward compatibility",
		},
		{
			name: "success: cloudbeat_scan_management_account: Yes",
			roles: []iam.Role{
				*makeRole("cloudbeat-root", scanSettingTagKey, "Yes"),
				*makeRole("cloudbeat-securityaudit"),
			},
			expectedLog: "assuming 'cloudbeat-securityaudit' role",
		},
		{
			name: "success: cloudbeat_scan_management_account: No",
			roles: []iam.Role{
				*makeRole("cloudbeat-root", scanSettingTagKey, "No"),
			},
			expectedLog: "assuming 'cloudbeat-securityaudit' role",
		},
		{
			name:                 "fail: cloudbeat-root does not exist",
			roles:                []iam.Role{},
			expectedErrorMessage: "role \"cloudbeat-root\" does not exist",
		},
		{
			name: "warn: cloudbeat_scan_management_account: Yes, but cloudbeat-securityaudit does not exist",
			roles: []iam.Role{
				*makeRole("cloudbeat-root", scanSettingTagKey, "Yes"),
			},
			expectedLog: fmt.Sprintf(
				"should be scanned (%s: %s), but %q role is missing",
				scanSettingTagKey, scanSettingTagValue, awslib.BenchmarkOrgIAMRoleNamesProvider{}.MemberRoleName(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AWSOrg{
				IAMProvider:       getMockIAMRoleGetter(tt.roles),
				IdentityProvider:  mockAwsIdentityProvider(nil),
				AccountProvider:   mockAccountProvider(nil),
				StatusHandler:     statushandler.NewMockStatusHandlerAPI(t),
				AWSCredsValidator: awslib.CredentialsValidatorNOOP,
				RoleNamesProvider: awslib.BenchmarkOrgIAMRoleNamesProvider{},
			}

			// set up log capture
			logCaptureBuf := &bytes.Buffer{}
			replacement := zap.WrapCore(func(zapcore.Core) zapcore.Core {
				return zapcore.NewCore(
					zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
					zapcore.AddSync(logCaptureBuf),
					zapcore.DebugLevel,
				)
			})
			log := testhelper.NewLogger(t).WithOptions(replacement)

			stsClient := &mockStsClient{}
			rootCfg := assumeRole(stsClient, aws.Config{}, "cloudbeat-root")
			identity := cloud.Identity{
				Account:      "123",
				AccountAlias: "some-name",
			}

			_, err := a.pickManagementAccountRole(t.Context(), log, stsClient, rootCfg, identity)
			if tt.expectedLog != "" {
				require.NotEmpty(t, logCaptureBuf, "expected logs, but captured none")
				require.Contains(t, logCaptureBuf.String(), tt.expectedLog, "expected message not found")
			}
			if tt.expectedErrorMessage != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedErrorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mockAccountProvider(err error) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	on := provider.EXPECT().ListAccounts(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return([]cloud.Identity{
			{
				Account:      "123",
				AccountAlias: "some-name",
			},
			{
				Account:      "456",
				AccountAlias: "some-other-name",
			},
		}, nil)
	} else {
		on.Return(nil, err)
	}
	return &provider
}

func mockAccountProviderWithIdentities(identities []cloud.Identity) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	provider.EXPECT().ListAccounts(mock.Anything, mock.Anything, mock.Anything).Return(identities, nil)
	return &provider
}

type mockIAMProvider struct {
	iam.MockRoleGetter
	roles []iam.Role
}

func getMockIAMRoleGetter(roles []iam.Role) iam.RoleGetter {
	result := &mockIAMProvider{
		MockRoleGetter: iam.MockRoleGetter{},
		roles:          roles,
	}
	on := result.MockRoleGetter.EXPECT().GetRole(mock.Anything, mock.AnythingOfType("string"))
	on.RunAndReturn(
		func(_ context.Context, roleName string) (*iam.Role, error) {
			for _, role := range result.roles {
				if *role.RoleName == roleName {
					return &role, nil
				}
			}
			return nil, fmt.Errorf("role %q does not exist", roleName)
		},
	)
	return result
}

func makeRole(name string, tagKeyValues ...string) *iam.Role {
	arn := "arn:aws:iam::123456789012" + name
	tags := []types.Tag{}
	for i := 0; i < len(tagKeyValues); i += 2 {
		t := types.Tag{
			Key:   &tagKeyValues[i],
			Value: &tagKeyValues[i+1],
		}
		tags = append(tags, t)
	}
	return &iam.Role{
		Role: types.Role{
			Arn:      &arn,
			RoleName: &name,
			Tags:     tags,
		},
	}
}

func mockAwsIdentityProviderWithCallerIdentityCall(err error) *awslib.MockIdentityProviderGetter {
	const account = "test-account"

	identityProvider := &awslib.MockIdentityProviderGetter{}

	callerIdentity := sts.GetCallerIdentityOutput{
		Account: pointers.Ref(account),
		Arn:     pointers.Ref(""),
	}

	onCallerIdentity := identityProvider.EXPECT().
		GetCallerIdentity(mock.Anything, mock.Anything).
		Return(callerIdentity, nil)

	onCallerIdentity.Once()
	if err != nil {
		onCallerIdentity.Times(2)
	}

	on := identityProvider.EXPECT().GetIdentity(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&cloud.Identity{
				Account: account,
			},
			nil,
		)
	} else {
		on.Return(nil, err)
	}

	return identityProvider
}

func mockIdentityProviderAlreadyCloudbeatRoot() awslib.IdentityProviderGetter {
	const account = "test-account"

	identityProvider := &awslib.MockIdentityProviderGetter{}

	callerIdentity := sts.GetCallerIdentityOutput{
		Account: pointers.Ref(account),
		Arn:     pointers.Ref("arn:aws:sts::test-account:assumed-role/cloudbeat-root/session-name"),
	}

	identityProvider.EXPECT().
		GetCallerIdentity(mock.Anything, mock.Anything).
		Return(callerIdentity, nil).
		Once()

	identityProvider.EXPECT().
		GetIdentity(mock.Anything, mock.MatchedBy(func(cnf aws.Config) bool {
			_ = cnf.Credentials
			cache, is := cnf.Credentials.(*aws.CredentialsCache)
			if !is {
				return false
			}

			sl := cache.ProviderSources()
			if len(sl) == 0 {
				return false
			}

			return sl[0] != aws.CredentialSourceSTSAssumeRole // no assume was run.
		})).
		Return(&cloud.Identity{Account: account}, nil)

	return identityProvider
}
