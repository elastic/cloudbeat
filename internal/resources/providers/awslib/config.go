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
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	libbeataws "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"

	"github.com/elastic/cloudbeat/internal/config"
)

func InitializeAWSConfig(cfg libbeataws.ConfigAWS) (*aws.Config, error) {
	awsConfig, err := libbeataws.InitializeAWSConfig(cfg)
	if err != nil {
		return nil, err
	}

	awsConfig.Retryer = func() aws.Retryer {
		return retry.NewStandard(func(o *retry.StandardOptions) {
			o.Retryables = append(o.Retryables, retry.RetryableHTTPStatusCode{
				Codes: map[int]struct{}{
					http.StatusTooManyRequests: {},
				},
			})
		})
	}

	return &awsConfig, nil
}

func CloudConnectorsExternalID(resourceID, externalIDPart string) string {
	return fmt.Sprintf("%s-%s", resourceID, externalIDPart)
}

func InitializeAWSConfigCloudConnectors(ctx context.Context, cfg config.AwsConfig) (*aws.Config, error) {
	// 1. Load initial config
	// (TODO: check directly assuming the first role in chain and/or libbeataws.InitializeAWSConfig(cfg))
	// (TODO: consider os.Setenv("AWS_EC2_METADATA_DISABLED", "true"))
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Create an STS client using the base credentials
	firstClient := sts.NewFromConfig(awsConfig)

	const defaultDuration = 5 * time.Minute

	// Chain Part 1 - Elastic Super Role Local
	localSuperRoleProvider := stscreds.NewAssumeRoleProvider(
		firstClient,
		cfg.CloudConnectorsConfig.LocalRoleARN,
		func(aro *stscreds.AssumeRoleOptions) {
			aro.RoleSessionName = "cloudbeat-super-role-local"
			aro.Duration = defaultDuration
		},
	)
	localSuperRoleCredentialsCache := aws.NewCredentialsCache(localSuperRoleProvider)

	// Chain Part 2 - Elastic Super Role Global
	globalSuperRoleCfg := awsConfig
	globalSuperRoleCfg.Credentials = localSuperRoleCredentialsCache
	globalSuperRoleProvider := stscreds.NewAssumeRoleProvider(
		sts.NewFromConfig(globalSuperRoleCfg),
		cfg.CloudConnectorsConfig.GlobalRoleARN,
		func(aro *stscreds.AssumeRoleOptions) {
			aro.RoleSessionName = "cloudbeat-super-role-global"
			aro.Duration = defaultDuration
		},
	)
	globalSuperRoleCredentialsCache := aws.NewCredentialsCache(globalSuperRoleProvider)

	// Chain Part 3 - Elastic Super Role Local
	customerRemoteRoleCfg := awsConfig
	customerRemoteRoleCfg.Credentials = globalSuperRoleCredentialsCache
	customerRemoteRoleProvider := stscreds.NewAssumeRoleProvider(
		sts.NewFromConfig(customerRemoteRoleCfg),
		cfg.Cred.RoleArn, // Customer Remote Role passed in package policy.
		func(aro *stscreds.AssumeRoleOptions) {
			aro.RoleSessionName = "cloudbeat-remote-role"
			aro.Duration = cfg.Cred.AssumeRoleDuration
			aro.ExternalID = aws.String(CloudConnectorsExternalID(cfg.CloudConnectorsConfig.LocalRoleARN, cfg.Cred.ExternalID))
		},
	)
	customerRemoteRoleCredentialsCache := aws.NewCredentialsCache(customerRemoteRoleProvider)

	ret := awsConfig
	ret.Credentials = customerRemoteRoleCredentialsCache

	return &ret, nil
}
