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
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"time"
)

var (
	keyMetadata = []types.AccessKeyMetadata{
		{
			AccessKeyId: aws.String("test_access_key"),
			CreateDate:  nil,
			Status:      types.StatusTypeActive,
			UserName:    aws.String("username"),
		},
	}

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
			UserName:         aws.String("test-user-1"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/test-user-1"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Now()),
		},
		{
			UserName:         aws.String("test-user-2"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/test-user-2"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Time{}),
		},
	}
)
