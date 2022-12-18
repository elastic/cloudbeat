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
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"strings"
	"time"
)

type User struct {
	Name        string
	AccessKeys  []AccessKey
	MFADevices  []AuthDevice
	LastAccess  time.Time
	Arn         string
	HasLoggedIn bool
}

type AuthDevice struct {
	types.MFADevice
	IsVirtual bool
}

type AccessKey struct {
	AccessKeyId  string
	Active       bool
	CreationDate time.Time
	LastAccess   time.Time
	HasUsed      bool
}

func (p Provider) GetUsers(ctx context.Context) ([]awslib.AwsResource, error) {
	var users []awslib.AwsResource
	apiUsers, err := p.listUsers(ctx)
	if err != nil {
		p.log.Errorf("fail to list users, error: %v", err)
		return nil, err
	}

	for _, apiUser := range apiUsers {
		mfaDevices, err := p.getMFADevices(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list mfa device for user: %v, error: %v", apiUser, err)
			return nil, err
		}

		keys, err := p.getUserKeys(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list access keys for user: %v, error: %v", apiUser, err)
			return nil, err
		}

		var lastAccess time.Time
		if apiUser.PasswordLastUsed != nil {
			lastAccess = *apiUser.PasswordLastUsed
		}

		var username string
		if apiUser.UserName != nil {
			username = *apiUser.UserName
		}

		var arn string
		if apiUser.Arn != nil {
			arn = *apiUser.Arn
		}

		users = append(users, User{
			Name:        username,
			Arn:         arn,
			AccessKeys:  keys,
			MFADevices:  mfaDevices,
			LastAccess:  lastAccess,
			HasLoggedIn: !lastAccess.IsZero(),
		})
	}

	return users, nil
}

func (u User) GetResourceArn() string {
	return u.Arn
}

func (u User) GetResourceName() string {
	return "iam-user"
}

func (u User) GetResourceType() string {
	return fetching.IAMUserType
}

func (p Provider) listUsers(ctx context.Context) ([]types.User, error) {
	p.log.Debug("IAMProvider.getUsers")
	var nativeUsers []types.User

	input := &iamsdk.ListUsersInput{}
	for {
		users, err := p.client.ListUsers(ctx, input)
		if err != nil {
			return nil, err
		}
		nativeUsers = append(nativeUsers, users.Users...)
		if !users.IsTruncated {
			break
		}
		input.Marker = users.Marker
	}

	p.log.Debugf("IAMProvider.getUsers return %d users", len(nativeUsers))
	return nativeUsers, nil
}

func (p Provider) getMFADevices(ctx context.Context, user types.User) ([]AuthDevice, error) {
	input := &iamsdk.ListMFADevicesInput{
		Marker:   nil,
		UserName: user.UserName,
	}

	var apiDevices []types.MFADevice
	for {
		output, err := p.client.ListMFADevices(ctx, input)
		if err != nil {
			return nil, err
		}
		apiDevices = append(apiDevices, output.MFADevices...)
		if !output.IsTruncated {
			break
		}
		input.Marker = output.Marker
	}

	var devices []AuthDevice
	for _, apiDevice := range apiDevices {
		isVirtual := true
		if !strings.HasPrefix(*apiDevice.SerialNumber, "arn:") {
			isVirtual = false
		}
		devices = append(devices, AuthDevice{
			MFADevice: apiDevice,
			IsVirtual: isVirtual,
		})
	}

	return devices, nil
}

func (p Provider) getUserKeys(ctx context.Context, apiUser types.User) ([]AccessKey, error) {

	var keys []AccessKey
	input := iamsdk.ListAccessKeysInput{
		UserName: apiUser.UserName,
	}
	for {
		output, err := p.client.ListAccessKeys(ctx, &input)
		if err != nil {
			return nil, err
		}

		for _, apiAccessKey := range output.AccessKeyMetadata {
			output, err := p.client.GetAccessKeyLastUsed(ctx, &iamsdk.GetAccessKeyLastUsedInput{
				AccessKeyId: apiAccessKey.AccessKeyId,
			})

			var lastUsed time.Time
			if err == nil {
				if output.AccessKeyLastUsed != nil && output.AccessKeyLastUsed.LastUsedDate != nil {
					lastUsed = *output.AccessKeyLastUsed.LastUsedDate
				}
			}

			var accessKeyId string
			if apiAccessKey.AccessKeyId != nil {
				accessKeyId = *apiAccessKey.AccessKeyId
			}

			creationDate := time.Now()
			if apiAccessKey.CreateDate != nil {
				creationDate = *apiAccessKey.CreateDate
			}

			keys = append(keys, AccessKey{
				AccessKeyId:  accessKeyId,
				Active:       apiAccessKey.Status == types.StatusTypeActive,
				CreationDate: creationDate,
				LastAccess:   lastUsed,
				HasUsed:      !lastUsed.IsZero(),
			})
		}

		if !output.IsTruncated {
			break
		}

		input.Marker = output.Marker
	}

	return keys, nil
}
