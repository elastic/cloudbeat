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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gocarina/gocsv"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const (
	rootAccount = "<root_account>"
	maxRetries  = 5
	interval    = 3 * time.Second
)

func (p Provider) GetUsers(ctx context.Context) ([]awslib.AwsResource, error) {
	apiUsers, err := p.listUsers(ctx)
	if err != nil {
		return nil, err
	}

	credentialReport, err := p.getCredentialReport(ctx)
	if err != nil {
		return nil, err
	}

	rootUser := p.getRootAccountUser(credentialReport[rootAccount])
	if rootUser != nil {
		apiUsers = append(apiUsers, *rootUser)
	}

	users := make([]awslib.AwsResource, 0, len(apiUsers))
	var userAccount *CredentialReport
	for _, apiUser := range apiUsers {
		var username string
		if apiUser.UserName != nil {
			username = *apiUser.UserName
		}

		var arn string
		if apiUser.Arn != nil {
			arn = *apiUser.Arn
		}

		keys := p.getUserKeys(*apiUser.UserName, credentialReport)

		if userAccount = credentialReport[aws.ToString(apiUser.UserName)]; userAccount == nil {
			continue
		}

		mfaDevices, err := p.getMFADevices(ctx, apiUser, userAccount)
		if err != nil {
			p.log.With("aws.iam.user.name", username, logp.Error(err)).Error("fail to list mfa device for user")
		}

		pwdEnabled, err := isPasswordEnabled(userAccount)
		if err != nil {
			p.log.With("aws.iam.user.name", username, logp.Error(err)).Error("fail to parse PasswordEnabled for user")
			pwdEnabled = false
		}

		inlinePolicies, err := p.listInlinePolicies(ctx, apiUser.UserName)
		if err != nil && !isRootUser(username) {
			p.log.With("aws.iam.user.name", username, logp.Error(err)).Error("fail to list inline policies for user")
		}

		attachedPolicies, err := p.listAttachedPolicies(ctx, apiUser.UserName)
		if err != nil && !isRootUser(username) {
			p.log.With("aws.iam.user.name", username, logp.Error(err)).Error("fail to list attached policies for user")
		}

		users = append(users, User{
			AccessKeys:          keys,
			MFADevices:          mfaDevices,
			InlinePolicies:      inlinePolicies,
			AttachedPolicies:    attachedPolicies,
			Name:                username,
			LastAccess:          userAccount.PasswordLastUsed,
			Arn:                 arn,
			PasswordLastChanged: userAccount.PasswordLastChanged,
			PasswordEnabled:     pwdEnabled,
			MfaActive:           userAccount.MfaActive,
			UserId:              pointers.Deref(apiUser.UserId),
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

func (u User) GetRegion() string {
	return awslib.GlobalRegion
}

func (p Provider) listUsers(ctx context.Context) ([]types.User, error) {
	p.log.Debug("IAMProvider.getUsers")
	var nativeUsers []types.User

	input := &iamsdk.ListUsersInput{}
	for {
		users, err := p.client.ListUsers(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
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

func (p Provider) getMFADevices(ctx context.Context, user types.User, userAccount *CredentialReport) ([]AuthDevice, error) {
	// For the root user, it's not possible to list all the devices, so instead we check all the virtual devices
	// to confirm if one is assigned the root user. If this is not the case, we can infer a hardware device is configured
	// (since we know MFA is active for the root user but cannot find a virtual device).
	if *user.UserName == rootAccount {
		return p.listRootMFADevice(ctx, userAccount)
	}

	return p.listMFADevices(ctx, user)
}

func (p Provider) listMFADevices(ctx context.Context, user types.User) ([]AuthDevice, error) {
	input := &iamsdk.ListMFADevicesInput{
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

	devices := make([]AuthDevice, 0, len(apiDevices))
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

func (p Provider) getUserKeys(username string, report map[string]*CredentialReport) []AccessKey {
	p.log.Debugf("aggregate access keys data for user: %s", username)
	entry := report[username]

	if entry == nil {
		p.log.Debugf("no entry for user: %s in credentials report", username)
		return nil
	}

	return []AccessKey{
		{
			Active:       entry.AccessKey1Active,
			LastAccess:   entry.AccessKey1LastUsed,
			HasUsed:      entry.AccessKey1LastUsed != "N/A",
			RotationDate: entry.AccessKey1LastRotated,
		}, {
			Active:       entry.AccessKey2Active,
			LastAccess:   entry.AccessKey2LastUsed,
			HasUsed:      entry.AccessKey2LastUsed != "N/A",
			RotationDate: entry.AccessKey2LastRotated,
		},
	}
}

//revive:disable-next-line:cognitive-complexity,cyclomatic
func (p Provider) getCredentialReport(ctx context.Context) (map[string]*CredentialReport, error) {
	report, err := p.client.GetCredentialReport(ctx, &iamsdk.GetCredentialReportInput{})
	if err != nil {
		var awsFailErr *types.ServiceFailureException
		if errors.As(err, &awsFailErr) {
			return nil, fmt.Errorf("could not gather aws iam credential report: %w", err)
		}

		// if we have an error, and it is not a server err we generate a report
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ReportNotPresent" || apiErr.ErrorCode() == "ReportExpired" {
				// generate a new report
				_, err = p.client.GenerateCredentialReport(ctx, &iamsdk.GenerateCredentialReportInput{})
				if err != nil {
					return nil, fmt.Errorf("failed to generate credential report: %w", err)
				}
			}
		}

		// loop until max retries or till the report is ready
		countRetries := 0
		report, err = p.client.GetCredentialReport(ctx, &iamsdk.GetCredentialReportInput{})
		if errors.As(err, &apiErr) {
			for apiErr.ErrorCode() == "NoSuchEntity" || apiErr.ErrorCode() == "ReportInProgress" {
				if countRetries >= maxRetries {
					return nil, fmt.Errorf("reached max retries: %w", err)
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
		if err != nil {
			return nil, fmt.Errorf("could not gather aws iam credential report: %w", err)
		}
		return nil, nil
	}

	parsedReport, err := parseCredentialsReport(report)
	if err != nil {
		return nil, fmt.Errorf("fail to parse credentials report: %w", err)
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

func isPasswordEnabled(userAccount *CredentialReport) (bool, error) {
	if userAccount.PasswordEnabled == "not_supported" {
		return false, nil
	}

	return strconv.ParseBool(userAccount.PasswordEnabled)
}
