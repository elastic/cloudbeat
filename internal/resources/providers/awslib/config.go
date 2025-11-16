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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	libbeataws "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
)

func RetryableCodesOption(o *retry.StandardOptions) {
	o.Retryables = append(o.Retryables, retry.RetryableHTTPStatusCode{
		Codes: map[int]struct{}{
			http.StatusTooManyRequests: {},
		},
	})
}

func awsConfigRetrier() aws.Retryer {
	return retry.NewStandard(RetryableCodesOption)
}

func InitializeAWSConfig(cfg libbeataws.ConfigAWS, logger *clog.Logger) (*aws.Config, error) {
	awsConfig, err := libbeataws.InitializeAWSConfig(cfg, logger.Logger)
	if err != nil {
		return nil, err
	}

	awsConfig.Retryer = awsConfigRetrier

	observability.AppendAWSMiddlewares(&awsConfig)

	return &awsConfig, nil
}

func CloudConnectorsExternalID(resourceID, externalIDPart string) string {
	return fmt.Sprintf("%s-%s", resourceID, externalIDPart)
}

func InitializeAWSConfigCloudConnectors(ctx context.Context, cfg config.AwsConfig) (*aws.Config, error) {
	const defaultDuration = 20 * time.Minute

	// 1. Load initial config - Chain Step 1 - Elastic Super Role Local implicitly assumed through IRSA.
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	observability.AppendAWSMiddlewares(&awsConfig)

	chain := []AWSRoleChainingStep{
		// Chain Step 2 - Elastic Super Role Global
		{
			RoleARN: cfg.CloudConnectorsConfig.GlobalRoleARN,
			Options: func(aro *stscreds.AssumeRoleOptions) {
				aro.RoleSessionName = "cloudbeat-super-role-global"
				aro.Duration = defaultDuration
			},
		},
		// Chain Step 3 - Elastic Super Role Local
		{
			RoleARN: cfg.Cred.RoleArn,
			Options: func(aro *stscreds.AssumeRoleOptions) {
				aro.RoleSessionName = "cloudbeat-remote-role"
				aro.Duration = cfg.Cred.AssumeRoleDuration
				aro.ExternalID = aws.String(CloudConnectorsExternalID(cfg.CloudConnectorsConfig.ResourceID, cfg.Cred.ExternalID))
			},
		},
	}

	retConf := AWSConfigRoleChaining(awsConfig, chain)
	retConf.Retryer = awsConfigRetrier

	return retConf, nil
}

// AWSConfigRoleChaining initializes an assume role provider and an credential cache for each step on the chain, using the previous one as client.
func AWSConfigRoleChaining(initialConfig aws.Config, chain []AWSRoleChainingStep) *aws.Config {
	var client *sts.Client
	var assumeRoleProvider *stscreds.AssumeRoleProvider
	var credentialsCache *aws.CredentialsCache
	cnf := initialConfig

	for _, c := range chain {
		client = sts.NewFromConfig(cnf) // create client using the credentials from previous or initial step.

		// create a assume role provider for the current chain part role.
		assumeRoleProvider = stscreds.NewAssumeRoleProvider(
			client,
			c.RoleARN,
			c.Options,
		)
		credentialsCache = aws.NewCredentialsCache(assumeRoleProvider)

		cnf.Credentials = credentialsCache
	}

	return &cnf
}

type AWSRoleChainingStep struct {
	RoleARN string
	Options func(aro *stscreds.AssumeRoleOptions)
}

func CredentialsValid(ctx context.Context, cnf aws.Config, log *clog.Logger) bool {
	_, err := cnf.Credentials.Retrieve(ctx)

	if err == nil {
		return true
	}

	if !strings.Contains(err.Error(), "not authorized to perform: sts:AssumeRole") {
		log.Errorf("Expected a 403 authorization error, but got: %v", err)
	}

	return false
}

type CredentialsValidator interface {
	Validate(ctx context.Context, cnf aws.Config, log *clog.Logger) bool
}

type CredentialsValidatorFunc func(ctx context.Context, cnf aws.Config, log *clog.Logger) bool

func (c CredentialsValidatorFunc) Validate(ctx context.Context, cnf aws.Config, log *clog.Logger) bool {
	return c(ctx, cnf, log)
}

var _ CredentialsValidator = (CredentialsValidatorFunc)(nil)

var CredentialsValidatorNOOP CredentialsValidatorFunc = func(_ context.Context, _ aws.Config, _ *clog.Logger) bool { return true }
