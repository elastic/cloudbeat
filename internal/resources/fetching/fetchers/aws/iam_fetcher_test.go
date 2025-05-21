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

package fetchers

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aatypes "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type IamFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

type mocksReturnVals map[string][]any

func TestIamFetcherTestSuite(t *testing.T) {
	s := new(IamFetcherTestSuite)

	suite.Run(t, s)
}

func (s *IamFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *IamFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *IamFetcherTestSuite) TestIamFetcher_Fetch() {
	testAccount := "test-account"
	pwdPolicy := iam.PasswordPolicy{
		ReusePreventionCount: 5,
		RequireLowercase:     true,
		RequireUppercase:     true,
		RequireNumbers:       true,
		RequireSymbols:       false,
		MaxAgeDays:           90,
		MinimumLength:        8,
	}

	iamUser := iam.User{
		AccessKeys: []iam.AccessKey{{
			Active:       false,
			HasUsed:      false,
			LastAccess:   "",
			RotationDate: "",
		},
		},
		MFADevices:          nil,
		Name:                "test",
		LastAccess:          "",
		Arn:                 "testArn",
		PasswordEnabled:     true,
		PasswordLastChanged: "",
		MfaActive:           true,
	}

	iamPolicy := iam.Policy{
		Policy: types.Policy{
			Arn:             aws.String("testArn"),
			AttachmentCount: aws.Int32(1),
			IsAttachable:    true,
		},
		Document: map[string]any{
			"Statements": []map[string]any{
				{
					"Resource": "*",
					"Action":   "*",
					"Effect":   "Allow",
				},
			},
		},
	}

	certificates := iam.ServerCertificatesInfo{
		Certificates: []types.ServerCertificateMetadata{
			{
				Expiration: &time.Time{},
			},
		},
	}

	accessAnalyzers := iam.AccessAnalyzers{
		Analyzers: []iam.AccessAnalyzer{
			{
				AnalyzerSummary: aatypes.AnalyzerSummary{Arn: aws.String("some-arn")},
				Region:          "region-1",
			},
			{
				AnalyzerSummary: aatypes.AnalyzerSummary{Arn: aws.String("some-other-arn")},
				Region:          "region-2",
			},
			{
				AnalyzerSummary: aatypes.AnalyzerSummary{Arn: aws.String("some-third-arn")},
				Region:          "region-2",
			},
		},
		Regions: []string{"region-1", "region-2"},
	}

	var tests = []struct {
		name               string
		mocksReturnVals    mocksReturnVals
		account            string
		numExpectedResults int
	}{
		{
			name: "Should not get any IAM resources",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {nil, errors.New("Fail to fetch iam pwd policy")},
				"GetUsers":               {nil, errors.New("Fail to fetch iam users")},
				"GetPolicies":            {nil, errors.New("Fail to fetch iam policies")},
				"ListServerCertificates": {nil, errors.New("Fail to fetch iam certificates")},
				"GetAccessAnalyzers":     {nil, errors.New("Fail to fetch access analyzers")},
			},
			account:            testAccount,
			numExpectedResults: 0,
		},
		{
			name: "Should get all AWS resources",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {pwdPolicy, nil},
				"GetUsers":               {[]awslib.AwsResource{iamUser}, nil},
				"GetPolicies":            {[]awslib.AwsResource{iamPolicy}, nil},
				"ListServerCertificates": {&certificates, nil},
				"GetAccessAnalyzers":     {accessAnalyzers, nil},
			},
			account:            testAccount,
			numExpectedResults: 5,
		},
		{
			name: "Should only get iam pwd policy",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {pwdPolicy, nil},
				"GetUsers":               {nil, errors.New("Fail to fetch iam users")},
				"GetPolicies":            {nil, errors.New("Fail to fetch iam policies")},
				"ListServerCertificates": {nil, errors.New("Fail to fetch iam certificates")},
				"GetAccessAnalyzers":     {nil, errors.New("Fail to fetch access analyzers")},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			name: "Should only get iam users",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {nil, errors.New("Fail to fetch iam pwd policy")},
				"GetUsers":               {[]awslib.AwsResource{iamUser}, nil},
				"GetPolicies":            {nil, errors.New("Fail to fetch iam policies")},
				"ListServerCertificates": {nil, errors.New("Fail to fetch iam certificates")},
				"GetAccessAnalyzers":     {nil, errors.New("Fail to fetch access analyzers")},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			name: "Should only get iam policies",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {nil, errors.New("Fail to fetch iam pwd policy")},
				"GetUsers":               {nil, errors.New("Fail to fetch iam users")},
				"GetPolicies":            {[]awslib.AwsResource{iamPolicy}, nil},
				"ListServerCertificates": {nil, errors.New("Fail to fetch iam certificates")},
				"GetAccessAnalyzers":     {nil, errors.New("Fail to fetch access analyzers")},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
		{
			name: "Should only get iam certificates",
			mocksReturnVals: mocksReturnVals{
				"GetPasswordPolicy":      {nil, errors.New("Fail to fetch iam pwd policy")},
				"GetUsers":               {nil, errors.New("Fail to fetch iam users")},
				"GetPolicies":            {nil, errors.New("Fail to fetch iam policies")},
				"ListServerCertificates": {&certificates, nil},
				"GetAccessAnalyzers":     {nil, errors.New("Fail to fetch access analyzers")},
			},
			account:            testAccount,
			numExpectedResults: 1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			t := s.T()
			ctx := t.Context()
			iamProviderMock := &iam.MockAccessManagement{}
			for funcName, returnVals := range test.mocksReturnVals {
				iamProviderMock.On(funcName, ctx).Return(returnVals...)
			}

			iamFetcher := IAMFetcher{
				log:         testhelper.NewLogger(s.T()),
				iamProvider: iamProviderMock,
				resourceCh:  s.resourceCh,
				cloudIdentity: &cloud.Identity{
					Account: test.account,
				},
			}

			err := iamFetcher.Fetch(ctx, cycle.Metadata{})
			s.Require().NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.Len(results, test.numExpectedResults)
		})
	}
}

func (s *IamFetcherTestSuite) TestIamResource_GetMetadata() {
	tests := []struct {
		name     string
		resource awslib.AwsResource
		expected fetching.ResourceMetadata
	}{
		{
			name: "Should return correct metadata for iam pwd policy",
			resource: iam.PasswordPolicy{
				ReusePreventionCount: 5,
				RequireLowercase:     true,
				RequireUppercase:     true,
				RequireNumbers:       true,
				RequireSymbols:       false,
				MaxAgeDays:           90,
				MinimumLength:        8,
			},
			expected: fetching.ResourceMetadata{
				Region:  "global",
				ID:      "test-account-account-password-policy",
				Type:    "identity-management",
				SubType: "aws-password-policy",
				Name:    "account-password-policy",
			},
		},
		{
			name: "Should return correct metadata for iam user",
			resource: iam.User{
				AccessKeys: []iam.AccessKey{{
					Active:       false,
					HasUsed:      false,
					LastAccess:   "",
					RotationDate: "",
				},
				},
				MFADevices:          nil,
				Name:                "test",
				LastAccess:          "",
				Arn:                 "test-user-arn",
				PasswordEnabled:     true,
				PasswordLastChanged: "",
				MfaActive:           true,
			},
			expected: fetching.ResourceMetadata{
				Region:  "global",
				ID:      "test-user-arn",
				Type:    "identity-management",
				SubType: "aws-iam-user",
				Name:    "test",
			},
		},
		{
			name: "Should return correct metadata for iam policy",
			resource: iam.Policy{
				Policy: types.Policy{
					PolicyName:      aws.String("test-policy"),
					Arn:             aws.String("test-policy-arn"),
					AttachmentCount: aws.Int32(1),
					IsAttachable:    true,
				},
				Document: map[string]any{
					"Statements": []map[string]any{
						{
							"Resource": "*",
							"Action":   "*",
							"Effect":   "Allow",
						},
					},
				},
			},
			expected: fetching.ResourceMetadata{
				Region:  "global",
				ID:      "test-policy-arn",
				Type:    "identity-management",
				SubType: "aws-policy",
				Name:    "test-policy",
			},
		},
		{
			name: "Should return correct metadata for iam certificate",
			resource: iam.ServerCertificatesInfo{
				Certificates: []types.ServerCertificateMetadata{
					{
						Expiration: &time.Time{},
					},
				},
			},
			expected: fetching.ResourceMetadata{
				Region:  "global",
				ID:      "test-account-account-iam-server-certificates",
				Type:    "identity-management",
				SubType: "aws-iam-server-certificate",
				Name:    "account-iam-server-certificates",
			},
		},
		{
			name: "Should return correct metadata for access analyzer",
			resource: iam.AccessAnalyzers{
				Analyzers: []iam.AccessAnalyzer{
					{

						AnalyzerSummary: aatypes.AnalyzerSummary{Arn: aws.String("some-arn")},
						Region:          "region-1",
					},
				},
			},
			expected: fetching.ResourceMetadata{
				Region:  "global",
				ID:      "test-account-account-access-analyzers",
				Type:    "identity-management",
				SubType: "aws-access-analyzers",
				Name:    "account-access-analyzers",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			iamResource := IAMResource{
				AwsResource: test.resource,
				identity: &cloud.Identity{
					Account: "test-account",
				},
			}

			meta, err := iamResource.GetMetadata()
			s.Require().NoError(err)
			s.Equal(test.expected, meta)

			m, err := iamResource.GetElasticCommonData()
			s.Require().NoError(err)
			switch iamResource.GetResourceType() {
			case fetching.IAMUserType, fetching.PolicyType:
				s.Len(m, 3)
			default:
				s.Len(m, 1)
			}
			s.Contains(m, "cloud.service.name")
		})
	}
}
