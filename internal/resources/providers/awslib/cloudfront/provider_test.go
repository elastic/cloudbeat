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

package cloudfront

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func Test_DescribeDistributions(t *testing.T) {
	var distributionSummary1 = types.DistributionSummary{
		Id:  pointers.Ref("E2L2S10R2365C3"),
		ARN: pointers.Ref("arn:aws:cloudfront::account_id:distribution/E2L2S10R2365C3"),
		Aliases: &types.Aliases{
			Quantity: pointers.Ref(int32(1)),
			Items:    []string{"www.example.com"},
		},
		CacheBehaviors: &types.CacheBehaviors{
			Quantity: pointers.Ref(int32(0)),
			Items:    []types.CacheBehavior{},
		},
		Comment:     pointers.Ref("example distribution"),
		DomainName:  pointers.Ref("example.com"),
		Enabled:     pointers.Ref(true),
		HttpVersion: types.HttpVersionHttp2and3,
		Status:      pointers.Ref("Deployed"),
	}

	testCases := []struct {
		name              string
		resources         []types.DistributionSummary
		expectedResources []awslib.AwsResource
		expectedErr       bool
	}{
		{
			name:              "case 1: should fail on fetch error",
			resources:         []types.DistributionSummary{},
			expectedResources: []awslib.AwsResource{},
			expectedErr:       true,
		},
		{
			name:      "case 2: should fetch correct resource",
			resources: []types.DistributionSummary{distributionSummary1},
			expectedResources: []awslib.AwsResource{
				Distribution{DistributionSummary: distributionSummary1},
			},
			expectedErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			provider := createMockProvider(t, testCase.resources, testCase.expectedErr)
			result, err := provider.DescribeDistributions(context.Background())

			if testCase.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Len(t, result, len(testCase.expectedResources))
			for i, resource := range result {
				assert.Equal(t, testCase.expectedResources[i], resource)
			}
		})
	}
}

//nolint:revive
func createMockProvider(t *testing.T, resources []types.DistributionSummary, expectedErr bool) *Provider {
	var returnErr error
	if expectedErr {
		returnErr = fmt.Errorf("test error")
	}

	client := MockClient{}
	client.On("ListDistributions", mock.Anything, mock.Anything).Return(
		&cloudfront.ListDistributionsOutput{
			DistributionList: &types.DistributionList{
				Items:       resources,
				IsTruncated: pointers.Ref(false),
			},
		},
		returnErr,
	).Once()

	provider := &Provider{
		log:    testhelper.NewLogger(t),
		client: &client,
	}
	return provider
}
