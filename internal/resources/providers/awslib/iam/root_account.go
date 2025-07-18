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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elastic/elastic-agent-libs/logp"
)

func (p Provider) getRootAccountUser(rootAccount *CredentialReport) *types.User {
	if rootAccount == nil {
		p.log.Error("no root account entry was provided")
		return nil
	}

	rootDate, err := time.Parse(time.RFC3339, rootAccount.UserCreation)
	if err != nil {
		p.log.With("aws.iam.user.name", rootAccount.User, logp.Error(err)).Errorf("fail to parse root account user creation, error: %v", err)
		return nil
	}

	pwdLastUsed := time.Time{}
	// "no_information" if never used, "N/A" if user has no password
	// Docs: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_getting-report.html
	if rootAccount.PasswordLastUsed != "no_information" && rootAccount.PasswordLastUsed != "N/A" {
		pwdLastUsed, err = time.Parse(time.RFC3339, rootAccount.PasswordLastUsed)
		if err != nil {
			p.log.With("aws.iam.user.name", rootAccount.User, logp.Error(err)).Errorf("fail to parse root account password last used, error: %v", err)
			return nil
		}
	}

	return &types.User{
		UserName:         &rootAccount.User,
		Arn:              &rootAccount.Arn,
		CreateDate:       &rootDate,
		PasswordLastUsed: &pwdLastUsed,
		UserId:           aws.String("0"),
	}
}

func (p Provider) listRootMFADevice(ctx context.Context, userAccount *CredentialReport) ([]AuthDevice, error) {
	if !userAccount.MfaActive {
		p.log.Debug("mfa is not enabled for the root account")
		return nil, nil
	}

	input := &iamsdk.ListVirtualMFADevicesInput{
		// We only want MFA devices associated with a user.
		AssignmentStatus: types.AssignmentStatusTypeAssigned,
	}

	// fetch all virtual mfa devices and find if one is assigned to the root account user.
	var virtualDevices []types.VirtualMFADevice
	for {
		output, err := p.client.ListVirtualMFADevices(ctx, input)
		if err != nil {
			return nil, err
		}
		virtualDevices = append(virtualDevices, output.VirtualMFADevices...)
		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	var devices []AuthDevice
	var rootMFADevice AuthDevice
	for _, device := range virtualDevices {
		if strings.HasSuffix(*device.SerialNumber, "root-account-mfa-device") {
			rootMFADevice = AuthDevice{
				IsVirtual: true,
				MFADevice: types.MFADevice{
					EnableDate:   device.EnableDate,
					SerialNumber: device.SerialNumber,
					UserName:     device.User.UserName,
				},
			}
			return append(devices, rootMFADevice), nil
		}
	}

	// represent a hardware mfa device assigned to the root account user
	rootMFADevice = AuthDevice{
		IsVirtual: false,
		MFADevice: types.MFADevice{},
	}

	return append(devices, rootMFADevice), nil
}

func isRootUser(username string) bool {
	return username == rootAccount
}
