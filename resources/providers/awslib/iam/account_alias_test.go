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
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type accountAliasMocks map[string][][]any

func Test_GetAccountAlias(t *testing.T) {
	tests := []struct {
		name              string
		accountAliasMocks accountAliasMocks
		expected          string
		wantErr           bool
	}{
		{
			name: "Should return the first account alias",
			accountAliasMocks: accountAliasMocks{
				"ListAccountAliases": {{&iamsdk.ListAccountAliasesOutput{AccountAliases: []string{"first", "second"}}, nil}},
			},
			expected: "first",
			wantErr:  false,
		}, {
			name: "Should return the empty account alias",
			accountAliasMocks: accountAliasMocks{
				"ListAccountAliases": {{&iamsdk.ListAccountAliasesOutput{AccountAliases: []string{}}, nil}},
			},
			expected: "",
			wantErr:  false,
		}, {
			name: "Should return an error when there is one",
			accountAliasMocks: accountAliasMocks{
				"ListAccountAliases": {{nil, errors.New("bla")}},
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		mockedClient := &MockClient{}
		for funcName, returnVals := range test.accountAliasMocks {
			for _, vals := range returnVals {
				mockedClient.On(funcName, mock.Anything, mock.Anything).Return(vals...).Once()
			}
		}

		p := Provider{
			client: mockedClient,
			log:    logp.NewLogger("iam-provider"),
		}

		result, err := p.GetAccountAlias(context.TODO())

		if !test.wantErr {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, test.expected, result)
			assert.Error(t, err)
		}

	}
}
