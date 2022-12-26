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
	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
	"time"
)

var (
	virtualMfaDevices = []types.MFADevice{
		{
			SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/test-user"),
			UserName:     aws.String("test-user-1"),
			EnableDate:   aws.Time(time.Now()),
		},
	}

	mfaDevices = []types.MFADevice{
		{
			SerialNumber: aws.String("MFA-Device"),
			UserName:     aws.String("test-user-2"),
			EnableDate:   aws.Time(time.Now()),
		},
	}

	apiUsers = []types.User{
		{
			UserName:         aws.String("user1"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/user1"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Now()),
		},
		{
			UserName:         aws.String("user2"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/user2"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Time{}),
		},
	}

	CredentialsReportOutput = &iamsdk.GetCredentialReportOutput{
		Content:        []byte(credentialsReportContent),
		GeneratedTime:  &time.Time{},
		ReportFormat:   "text/csv",
		ResultMetadata: middleware.Metadata{},
	}
)

const credentialsReportContent = `"user,arn,user_creation_time,password_enabled,password_last_used,password_last_changed,password_next_rotation,mfa_active,access_key_1_active,access_key_1_last_rotated,access_key_1_last_used_date,access_key_2_active,access_key_2_last_rotated,access_key_2_last_used_date,cert_1_active,cert_2_active\n
				<root_account>,arn:aws:iam::1234567890:root,1970-01-01T00:00:00+00:00,true,2022-01-02T00:00:00+00:00,1970-01-01T00:00:00+00:00,2022-01-03T00:00:00+00:00,false,true,1970-01-01T00:00:00+00:00,2022-01-04T00:00:00+00:00,true,1970-01-01T00:00:00+00:00,2022-01-05T00:00:00+00:00,true,true\n
				user1,arn:aws:iam::1234567890:user/user1,2022-01-01T00:00:00+00:00,true,2022-01-02T00:00:00+00:00,2022-01-03T00:00:00+00:00,2022-01-04T00:00:00+00:00,true,true,2022-01-05T00:00:00+00:00,2022-01-06T00:00:00+00:00,true,2022-01-07T00:00:00+00:00,2022-01-08T00:00:00+00:00,true,true\n
				user2,arn:aws:iam::1234567890:user/user2,2022-01-09T00:00:00+00:00,false,,,2022-01-10T00:00:00+00:00,true,true,2022-01-11T00:00:00+00:00,2022-01-12T00:00:00+00:00,true,2022-01-13T00:00:00+00:00,2022-01-14T00:00:00+00:00,true,true"`
