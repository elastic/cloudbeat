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

package awslib

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestIdentityProvider_GetIdentity_Partition(t *testing.T) {
	tests := []struct {
		name          string
		callerArn     string
		wantPartition string
	}{
		{
			name:          "commercial",
			callerArn:     "arn:aws:sts::111122223333:assumed-role/cloudbeat-root/session",
			wantPartition: DefaultPartition,
		},
		{
			name:          "govcloud",
			callerArn:     "arn:aws-us-gov:sts::111122223333:assumed-role/cloudbeat-asset-inventory-root/session",
			wantPartition: GovCloudPartition,
		},
		{
			name:          "empty caller arn",
			callerArn:     "",
			wantPartition: DefaultPartition,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := aws.Config{
				Region:      DefaultRegion,
				Credentials: aws.AnonymousCredentials{},
				APIOptions:  []func(*middleware.Stack) error{fakeIdentityAPI(tt.callerArn)},
			}
			p := IdentityProvider{Logger: testhelper.NewLogger(t)}

			got, err := p.GetIdentity(t.Context(), cfg)
			require.NoError(t, err)
			assert.Equal(t, &cloud.Identity{
				Provider:     "aws",
				Account:      "111122223333",
				AccountAlias: "my-alias",
				Partition:    tt.wantPartition,
			}, got)
		})
	}
}

// fakeIdentityAPI answers the STS and IAM calls of GetIdentity. It does not send requests to AWS.
func fakeIdentityAPI(callerArn string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("fakeIdentityAPI",
			func(_ context.Context, in middleware.InitializeInput, _ middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
				switch in.Parameters.(type) {
				case *sts.GetCallerIdentityInput:
					return middleware.InitializeOutput{Result: &sts.GetCallerIdentityOutput{
						Account: aws.String("111122223333"),
						Arn:     aws.String(callerArn),
					}}, middleware.Metadata{}, nil
				case *iam.ListAccountAliasesInput:
					return middleware.InitializeOutput{Result: &iam.ListAccountAliasesOutput{
						AccountAliases: []string{"my-alias"},
					}}, middleware.Metadata{}, nil
				default:
					return middleware.InitializeOutput{}, middleware.Metadata{}, fmt.Errorf("unexpected AWS call %T", in.Parameters)
				}
			}), middleware.Before)
	}
}
