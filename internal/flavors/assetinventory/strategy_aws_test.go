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

package assetinventory

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

func TestInitOrgFetchers_AssumeRoleARNs(t *testing.T) {
	tests := []struct {
		name         string
		partition    string
		wantRoleARNs []string
	}{
		{
			name:      "commercial",
			partition: "aws",
			wantRoleARNs: []string{
				"arn:aws:iam::111:role/cloudbeat-asset-inventory-root",
				"arn:aws:iam::111:role/cloudbeat-asset-inventory-securityaudit",
				"arn:aws:iam::222:role/cloudbeat-asset-inventory-securityaudit",
			},
		},
		{
			name:      "govcloud",
			partition: "aws-us-gov",
			wantRoleARNs: []string{
				"arn:aws-us-gov:iam::111:role/cloudbeat-asset-inventory-root",
				"arn:aws-us-gov:iam::111:role/cloudbeat-asset-inventory-securityaudit",
				"arn:aws-us-gov:iam::222:role/cloudbeat-asset-inventory-securityaudit",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotRoleARNs []string
			session := &awsSession{
				config:   fakeOrgAWSConfig(tt.partition, []string{"111", "222"}, &gotRoleARNs),
				identity: &cloud.Identity{Account: "111", Partition: tt.partition},
			}
			s := &strategy{
				logger: testhelper.NewLogger(t),
				cfg: &config.Config{CloudConfig: config.CloudConfig{Aws: config.AwsConfig{
					AccountType: config.OrganizationAccount,
				}}},
			}

			fetchers, err := s.initOrgFetchers(t.Context(), session, statushandler.NewMockStatusHandlerAPI(t))
			require.NoError(t, err)
			assert.NotEmpty(t, fetchers)
			assert.Equal(t, tt.wantRoleARNs, gotRoleARNs)
		})
	}
}

type fakeInputKey struct{}

// fakeOrgAWSConfig returns an aws.Config that answers STS AssumeRole and Organizations ListAccounts
// without calls to AWS. It records the RoleArn of each AssumeRole call. Other calls fail.
// It answers after the signing step, so the SDK gets the credentials (and assumes roles) as it does with AWS.
func fakeOrgAWSConfig(partition string, accounts []string, roleARNs *[]string) awssdk.Config {
	keepInput := middleware.InitializeMiddlewareFunc("fakeOrgAWSInput",
		func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
			return next.HandleInitialize(middleware.WithStackValue(ctx, fakeInputKey{}, in.Parameters), in)
		})
	answer := middleware.FinalizeMiddlewareFunc("fakeOrgAWSAnswer",
		func(ctx context.Context, _ middleware.FinalizeInput, _ middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
			switch input := middleware.GetStackValue(ctx, fakeInputKey{}).(type) {
			case *sts.AssumeRoleInput:
				*roleARNs = append(*roleARNs, pointers.Deref(input.RoleArn))
				return middleware.FinalizeOutput{Result: &sts.AssumeRoleOutput{
					Credentials: &ststypes.Credentials{
						AccessKeyId:     pointers.Ref("ASIATEST"),
						SecretAccessKey: pointers.Ref("secret"),
						SessionToken:    pointers.Ref("token"),
						Expiration:      pointers.Ref(time.Now().Add(time.Hour)),
					},
				}}, middleware.Metadata{}, nil
			case *organizations.ListAccountsInput:
				out := &organizations.ListAccountsOutput{}
				for _, id := range accounts {
					out.Accounts = append(out.Accounts, orgtypes.Account{
						Arn:    pointers.Ref(fmt.Sprintf("arn:%s:organizations::111:account/o-test/%s", partition, id)),
						Id:     pointers.Ref(id),
						Status: orgtypes.AccountStatusActive,
					})
				}
				return middleware.FinalizeOutput{Result: out}, middleware.Metadata{}, nil
			default:
				return middleware.FinalizeOutput{}, middleware.Metadata{}, errors.New("no fake response")
			}
		})

	return awssdk.Config{
		Region:      awslib.DefaultRegion,
		Credentials: awssdk.NewCredentialsCache(credentials.NewStaticCredentialsProvider("AKIATEST", "secret", "")),
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				if err := stack.Initialize.Add(keepInput, middleware.Before); err != nil {
					return err
				}
				return stack.Finalize.Add(answer, middleware.After)
			},
		},
	}
}
