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
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	smithy "github.com/aws/smithy-go"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const dateLayout = "2006-01-02T15:04:05+00:00"

func (p Provider) GetUsers(ctx context.Context) ([]awslib.AwsResource, error) {
	var users []awslib.AwsResource
	apiUsers, err := p.listUsers(ctx)
	if err != nil {
		p.log.Errorf("fail to list users, error: %v", err)
		return nil, err
	}

	credentialReport, err := p.getCredentialReport(ctx)
	if err != nil {
		return nil, err
	}

	rootUser, err := p.createRootAccountUser(credentialReport["<root_account>"])
	if err != nil {
		return nil, errors.Wrap(err, "fail to construct a root account user")
	}
	apiUsers = append(apiUsers, *rootUser)

	for _, apiUser := range apiUsers {
		mfaDevices, err := p.getMFADevices(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list mfa device for user: %v, error: %v", apiUser, err)
		}

		keys, err := p.getUserKeys(ctx, apiUser)
		if err != nil {
			p.log.Errorf("fail to list access keys for user: %v, error: %v", apiUser, err)
		}

		var username string
		if apiUser.UserName != nil {
			username = *apiUser.UserName
		}

		var arn string
		if apiUser.Arn != nil {
			arn = *apiUser.Arn
		}

		userAccount := credentialReport[aws.ToString(apiUser.UserName)]
		users = append(users, User{
			Name:                username,
			Arn:                 arn,
			AccessKeys:          keys,
			MFADevices:          mfaDevices,
			LastAccess:          userAccount.PasswordLastUsed,
			PasswordEnabled:     userAccount.PasswordEnabled,
			PasswordLastChanged: userAccount.PasswordLastChanged,
			MfaActive:           userAccount.MfaActive,
		})
	}

	return users, nil
}

func (u User) GetResourceArn() string {
	return u.Arn
}

func (u User) GetResourceName() string {
	return u.Name
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
				CreationDate: creationDate.String(),
				LastAccess:   lastUsed.String(),
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

func (p Provider) getCredentialReport(ctx context.Context) (map[string]*CredentialReport, error) {
	var (
		countRetries = 0
		maxRetries   = 5
		interval     = 3 * time.Second
	)

	var ae smithy.APIError
	report, err := p.client.GetCredentialReport(ctx, &iamsdk.GetCredentialReportInput{})
	if err != nil {
		var awsFailErr *types.ServiceFailureException
		if errors.As(err, &awsFailErr) {
			return nil, errors.Wrap(err, "could not gather aws iam credential report")
		}

		// if we have an error, and it is not a server err we generate a report
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "ReportNotPresent" || ae.ErrorCode() == "ReportExpired" {
				// generate a new report
				_, err := p.client.GenerateCredentialReport(ctx, &iamsdk.GenerateCredentialReportInput{})
				if err != nil {
					return nil, err
				}
			}
		}

		// loop until max retires or till the report is ready
		report, err = p.client.GetCredentialReport(ctx, &iamsdk.GetCredentialReportInput{})
		if errors.As(err, &ae) {
			for ae.ErrorCode() == "NoSuchEntity" || ae.ErrorCode() == "ReportInProgress" {
				if countRetries >= maxRetries {
					return nil, errors.Wrap(err, "reached to max retries")
				}

				report, err = p.client.GetCredentialReport(ctx, &iamsdk.GetCredentialReportInput{})
				if err == nil {
					break
				}

				countRetries++
				time.Sleep(interval)
			}
		}
	}

	if report == nil {
		return nil, errors.Wrap(err, "could not gather aws iam credential report")
	}

	parsedReport, err := parseCredentialsReport(report)
	if err != nil {
		return nil, errors.Wrap(err, "fail to parse credentials report")
	}

	return parsedReport, nil
}

func parseCredentialsReport(report *iamsdk.GetCredentialReportOutput) (map[string]*CredentialReport, error) {
	var credentialReportCSV []*CredentialReport
	if err := gocsv.Unmarshal(bytes.NewReader(report.Content), &credentialReportCSV); err != nil {
		return nil, err
	}

	credentialReport := make(map[string]*CredentialReport)
	for i := range credentialReportCSV {
		credentialReport[credentialReportCSV[i].User] = credentialReportCSV[i]
	}

	return credentialReport, nil
}

func (p Provider) createRootAccountUser(rootAccount *CredentialReport) (*types.User, error) {
	rootDate, err := time.Parse(dateLayout, rootAccount.UserCreation)
	if err != nil {
		return nil, err
	}

	pwdLastUsed, err := time.Parse(dateLayout, rootAccount.PasswordLastUsed)
	if err != nil {
		return nil, err
	}

	return &types.User{
		UserName:         aws.String("root_account"),
		Arn:              &rootAccount.Arn,
		CreateDate:       &rootDate,
		PasswordLastUsed: &pwdLastUsed,
		UserId:           aws.String("0"),
	}, nil
}
