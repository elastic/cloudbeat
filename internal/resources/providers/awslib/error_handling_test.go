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
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/internal/statushandler"
)

func TestIsPermissionErrorAndReportMissingPermissions(t *testing.T) {
	const expectedMessage = "missing permission on cloud provider side: arn:aws:iam::aws:policy/SecurityAudit"
	tests := map[string]struct {
		input                     error
		expectedIsPermissionError bool
		initMock                  func(*statushandler.MockStatusHandlerAPI)
	}{
		"simpler error": {
			input:                     errors.New("error"),
			expectedIsPermissionError: false,
		},
		"AccessDenied": {
			input:                     &smithy.GenericAPIError{Code: "AccessDenied"},
			expectedIsPermissionError: true,
			initMock: func(msha *statushandler.MockStatusHandlerAPI) {
				msha.EXPECT().Degraded(expectedMessage).Once()
			},
		},
		"AccessDeniedException": {
			input:                     &smithy.GenericAPIError{Code: "AccessDeniedException"},
			expectedIsPermissionError: true,
			initMock: func(msha *statushandler.MockStatusHandlerAPI) {
				msha.EXPECT().Degraded(expectedMessage).Once()
			},
		},
		"UnauthorizedOperation": {
			input:                     &smithy.GenericAPIError{Code: "UnauthorizedOperation"},
			expectedIsPermissionError: true,
			initMock: func(msha *statushandler.MockStatusHandlerAPI) {
				msha.EXPECT().Degraded(expectedMessage).Once()
			},
		},
		"MissingAuthenticationToken": {
			input:                     &smithy.GenericAPIError{Code: "MissingAuthenticationToken"},
			expectedIsPermissionError: false,
			initMock:                  func(_ *statushandler.MockStatusHandlerAPI) {},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mh := statushandler.NewMockStatusHandlerAPI(t)
			if tc.initMock != nil {
				tc.initMock(mh)
			}
			got := isPermissionError(tc.input)
			assert.Equal(t, tc.expectedIsPermissionError, got)
			ReportMissingPermission(mh, tc.input)
		})
	}
}
