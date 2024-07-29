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

package awsfetcher

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/lambda"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestLambdaFunction_Fetch(t *testing.T) {
	function1 := lambda.FunctionInfo{
		Function: types.FunctionConfiguration{
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
	}

	in := []awslib.AwsResource{function1}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsLambdaFunction,
			[]string{"arn:aws:lambda:us-east-1:378890115541:function:kuba-test-func"},
			"kuba-test-func",
			inventory.WithRawAsset(function1),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS Lambda",
				},
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_lambda")
	provider := newMockLambdaProvider(t)

	provider.On("ListEventSourceMappings", mock.Anything, mock.Anything).Return([]awslib.AwsResource{}, nil)
	provider.On("ListLayers", mock.Anything, mock.Anything).Return([]awslib.AwsResource{}, nil)
	provider.EXPECT().ListFunctions(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newLambdaFetcher(logger, identity, provider)

	collectResourcesAndMatch(t, fetcher, expected)
}
