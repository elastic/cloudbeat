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
	"errors"
	"testing"
	"time"

	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

func TestProvider_ListServerCertificates(t *testing.T) {
	certificates := []types.ServerCertificateMetadata{
		{
			Expiration: &time.Time{},
		},
	}

	certificatesResponse := iamsdk.ListServerCertificatesOutput{
		ServerCertificateMetadataList: certificates,
	}

	certificatesInfo := ServerCertificatesInfo{
		Certificates: certificates,
	}

	tests := []struct {
		name             string
		mockReturnValues mocksReturnVals
		want             awslib.AwsResource
		wantErr          bool
	}{
		{
			name: "Should return an error when listing server certificates fails",
			mockReturnValues: mocksReturnVals{
				"ListServerCertificates": {
					{
						nil,
						errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Should return a resource when listing server certificates succeeds",
			mockReturnValues: mocksReturnVals{
				"ListServerCertificates": {
					{
						&certificatesResponse,
						nil,
					},
				},
			},
			want:    &certificatesInfo,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createProviderFromMockValues(t, tt.mockReturnValues)

			got, err := p.ListServerCertificates(t.Context())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
