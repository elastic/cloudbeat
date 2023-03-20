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

//go:build component
// +build component

package flavors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/assert"
)

type APIResponse struct {
	// Full URL of the request, including query string and params
	URL string `json:"url"`
	// HTTP method (GET, POST, etc)
	Method string `json:"method"`
	// The response as json
	Response interface{} `json:"response"`
	// Error message instead of normal response
	Error *APIResponseError `json:"error"`

	target interface{}
}

type APIResponseError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func MustLoad(location string) []byte {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	b, err := os.ReadFile(path.Join(pwd, "test_data", location))
	if err != nil {
		panic(err)
	}
	return b
}

func mockAWSCalls(requests map[string]interface{}, t *testing.T) func(*middleware.Stack) error {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	calls := map[string]APIResponse{}
	for request, target := range requests {
		b, err := os.ReadFile(path.Join(pwd, "test_data", request))
		assert.NoError(t, err)
		res := APIResponse{}
		assert.NoError(t, json.Unmarshal(b, &res))
		res.target = target
		calls[fmt.Sprintf("%s_%s", res.Method, res.URL)] = res
	}
	return func(stack *middleware.Stack) error {
		return stack.Deserialize.Add(
			middleware.DeserializeMiddlewareFunc(
				"Fake",
				func(ctx context.Context, di middleware.DeserializeInput, dh middleware.DeserializeHandler) (middleware.DeserializeOutput, middleware.Metadata, error) {
					req, ok := di.Request.(*smithyhttp.Request)
					if !ok {
						return dh.HandleDeserialize(ctx, di)
					}
					r, ok := calls[fmt.Sprintf("%s_%s", req.Method, req.URL.String())]
					if !ok {
						t.Fatalf("%s call to %s expected but not registred", req.Method, req.URL.String())
						t.FailNow()
					}
					if r.Error != nil {
						if r.Error.Type == "APIError" {
							return middleware.DeserializeOutput{}, middleware.Metadata{}, &smithy.GenericAPIError{
								Code:    r.Error.Code,
								Message: r.Error.Message,
							}
						}
					}
					b, err := json.Marshal(r.Response)
					assert.NoError(t, err)
					response := r.target
					assert.NoError(t, json.Unmarshal(b, response))
					return middleware.DeserializeOutput{
						Result: response,
					}, middleware.Metadata{}, nil
				}),
			middleware.Before,
		)
	}
}
