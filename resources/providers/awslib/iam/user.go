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
		//metadata := a.CreateMetadataFromARN(*apiDevice.SerialNumber)
		if !strings.HasPrefix(*apiDevice.SerialNumber, "arn:") {
			//metadata = a.CreateMetadataFromARN(*user.Arn)
			isVirtual = false
		}
		devices = append(devices, AuthDevice{
			MFADevice: apiDevice,
			IsVirtual: isVirtual,
		})
	}

	return devices, nil
}

func (p Provider) GetUsers(ctx context.Context) ([]awslib.AwsResource, error) {
	var users []awslib.AwsResource
	apiUsers, err := p.listUsers(ctx)
	if err != nil {
		p.log.Errorf("fail to list users, error: %v", err)
	}

	for _, apiUser := range apiUsers {
		mfaDevices, err := p.getMFADevices(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list mfa device for user: %s, error: %v", apiUser, err)
			return nil, err
		}

		keys, err := p.getUserKeys(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list access keys for user: %s, error: %v", apiUser, err)
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

func (u User) GetResourceArn() string {
	return u.Arn
}

func (u User) GetResourceName() string {
	return "iam-user"
}

func (u User) GetResourceType() string {
	return fetching.IAMUserType
}

//func (p Provider) getUserGroups(ctx context.Context, apiUser types.User) []iam.Group {
//	var groups []iam.Group
//
//	input := &iam.ListGroupsForUserInput{
//		UserName: apiUser.UserName,
//	}
//	for {
//		output, err := p.client.ListGroupsForUser(ctx, input)
//		if err != nil {
//			p.log.Errorf("Failed to locate groups attached to user '%s': %v", *apiUser.UserName, err)
//			break
//		}
//
//		for _, apiGroup := range output.Groups {
//			group, err := p.adaptGroup(apiGroup, nil)
//			if err != nil {
//				p.log.Errorf("Failed to adapt group attached to user '%s': %v", *apiUser.UserName, err)
//				continue
//			}
//			groups = append(groups, *group)
//		}
//		if !output.IsTruncated {
//			break
//		}
//		input.Marker = output.Marker
//	}
//	return groups
//}
//
//func (p Provider) adaptGroup(ctx context.Context, apiGroup types.Group, state *state.State) (*iam.Group, error) {
//	if apiGroup.Arn == nil {
//		return nil, errors.New("group arn not specified")
//	}
//	if apiGroup.GroupName == nil {
//		return nil, errors.New("group name not specified")
//	}
//
//	var policies []iam.Policy
//	{
//		input := &iam.ListAttachedGroupPoliciesInput{
//			GroupName: apiGroup.GroupName,
//		}
//		for {
//			policiesOutput, err := p.client.ListAttachedGroupPolicies(ctx, input)
//			if err != nil {
//				p.log.Errorf("Failed to locate policies attached to group '%s': %v", *apiGroup.GroupName, err)
//				break
//			}
//
//			for _, apiPolicy := range policiesOutput.AttachedPolicies {
//				policy, err := p.adaptAttachedPolicy(apiPolicy)
//				if err != nil {
//					a.Debug("Failed to adapt policy attached to group '%s': %s", *apiGroup.GroupName, err)
//					continue
//				}
//				policies = append(policies, *policy)
//			}
//
//			if !policiesOutput.IsTruncated {
//				break
//			}
//			input.Marker = policiesOutput.Marker
//		}
//	}
//
//	var users []iam.User
//	if state != nil {
//		for _, user := range state.AWS.IAM.Users {
//			for _, userGroup := range user.Groups {
//				if userGroup.Name.EqualTo(*apiGroup.GroupName) {
//					users = append(users, user)
//				}
//			}
//		}
//	}
//
//	return &iam.Group{
//		Name:     types.String(*apiGroup.GroupName, metadata),
//		Users:    users,
//		Policies: policies,
//	}, nil
//}
//
//func (p Provider) adaptPolicy(ctx context.Context, apiPolicy types.Policy) (*iam.Policy, error) {
//
//	if apiPolicy.Arn == nil {
//		return nil, errors.New("policy arn not specified")
//	}
//	if apiPolicy.PolicyName == nil {
//		return nil, errors.New("policy name not specified")
//	}
//
//	output, err := p.client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
//		PolicyArn: apiPolicy.Arn,
//		VersionId: apiPolicy.DefaultVersionId,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	document, err := iamgo.ParseString(*output.PolicyVersion.Document)
//	if err != nil {
//		return nil, err
//	}
//
//	name := defsecTypes.StringDefault("", metadata)
//	if apiPolicy.PolicyName != nil {
//		name = defsecTypes.String(*apiPolicy.PolicyName, metadata)
//	}
//
//	return &iam.Policy{
//		Metadata: metadata,
//		Name:     name,
//		Document: iam.Document{
//			Metadata: metadata,
//			Parsed:   *document,
//		},
//		Builtin: defsecTypes.Bool(strings.HasPrefix(*apiPolicy.Arn, "arn:aws:iam::aws:"), metadata),
//	}, nil
//}
//
//func (p Provider) adaptAttachedPolicy(ctx context.Context, apiPolicy types.AttachedPolicy) (*iam.Policy, error) {
//
//	if apiPolicy.PolicyArn == nil {
//		return nil, errors.New("policy arn not specified")
//	}
//	if apiPolicy.PolicyName == nil {
//		return nil, errors.New("policy name not specified")
//	}
//
//	policyOutput, err := p.client.GetPolicy(ctx, &iam.GetPolicyInput{
//		PolicyArn: apiPolicy.PolicyArn,
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	return p.adaptPolicy(*policyOutput.Policy)
//}
