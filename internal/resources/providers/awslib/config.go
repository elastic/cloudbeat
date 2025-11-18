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
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	libbeataws "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/logp"

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

func InitializeAWSConfig(cfg libbeataws.ConfigAWS, logger *logp.Logger) (*aws.Config, error) {
	awsConfig, err := libbeataws.InitializeAWSConfig(cfg, logger)
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

// InitializeAWSConfigCloudConnectors initializes AWS config for Cloud Connectors deployment.
// It automatically selects between OIDC-based authentication (if JWT token is available)
// or IRSA-based authentication, both using multi-role assumption chains.
func InitializeAWSConfigCloudConnectors(ctx context.Context, cfg config.AwsConfig) (*aws.Config, error) {
	irsaFilePath := os.Getenv(config.CloudConnectorsAWSTokenEnvVar)
	if irsaFilePath != "" {
		return NewAWSConfigIRSAChain(ctx, cfg)
	}

	oidcFilePath := os.Getenv(config.CloudConnectorsJWTPathEnvVar)
	if oidcFilePath != "" {
		return NewAWSConfigOIDCChain(ctx, oidcFilePath, cfg)
	}

	return nil, errors.New("unable to initialize AWS config for Cloud Connectors: no authentication method available")
}

// NewAWSConfigIRSAChain creates an AWS config using IRSA (IAM Roles for Service Accounts) with role chaining.
// This function performs a two-step role assumption chain:
//  1. Uses IRSA to implicitly authenticate as the local role (via LoadDefaultConfig)
//  2. Uses the local role credentials to assume the global role, then the remote/target role with an external ID
func NewAWSConfigIRSAChain(ctx context.Context, cfg config.AwsConfig) (*aws.Config, error) {
	const defaultDuration = 20 * time.Minute

	// 1. Load initial config - Chain Step 1 - Elastic Super Role Local implicitly assumed through IRSA.
	awsConfig, err := LoadDefaultConfigWithRegion(ctx, cfg)
	if err != nil {
		return nil, err
	}

	observability.AppendAWSMiddlewares(awsConfig)

	chain := []AWSRoleChainingStep{
		// Chain Step 2 - Elastic Super Role Global
		&AssumeRoleStep{
			RoleARN: cfg.CloudConnectorsConfig.GlobalRoleARN,
			Options: func(aro *stscreds.AssumeRoleOptions) {
				aro.RoleSessionName = "cloudbeat-super-role-global"
				aro.Duration = defaultDuration
			},
		},
		// Chain Step 3 - Remote Role
		&AssumeRoleStep{
			RoleARN: cfg.Cred.RoleArn,
			Options: func(aro *stscreds.AssumeRoleOptions) {
				aro.RoleSessionName = "cloudbeat-remote-role"
				aro.Duration = cfg.Cred.AssumeRoleDuration
				aro.ExternalID = aws.String(CloudConnectorsExternalID(cfg.CloudConnectorsConfig.ResourceID, cfg.Cred.ExternalID))
			},
		},
	}

	retConf := AWSConfigRoleChaining(*awsConfig, chain)
	retConf.Retryer = awsConfigRetrier

	return retConf, nil
}

// NewAWSConfigOIDCChain creates an AWS config using OIDC/Web Identity token-based authentication with role chaining.
// This function performs a two-step role assumption chain:
//  1. Uses AssumeRoleWithWebIdentity to authenticate as the global role using a JWT token from the specified file
//  2. Uses the global role credentials to assume the remote/target role with an external ID
func NewAWSConfigOIDCChain(ctx context.Context, jwtFilePath string, cfg config.AwsConfig) (*aws.Config, error) {
	const defaultDuration = 20 * time.Minute

	// Load base AWS config
	awsConfig, err := LoadDefaultConfigWithRegion(ctx, cfg)
	if err != nil {
		return nil, err
	}

	observability.AppendAWSMiddlewares(awsConfig)

	chain := []AWSRoleChainingStep{
		// Chain Step 1 - Elastic Super Role Global via Web Identity
		&WebIdentityRoleStep{
			RoleARN:              cfg.CloudConnectorsConfig.GlobalRoleARN,
			WebIdentityTokenFile: jwtFilePath,
			Options: func(o *stscreds.WebIdentityRoleOptions) {
				o.RoleSessionName = "cloudbeat-super-role-global"
				o.Duration = defaultDuration
			},
		},
		// Chain Step 2 - Remote Role
		&AssumeRoleStep{
			RoleARN: cfg.Cred.RoleArn,
			Options: func(aro *stscreds.AssumeRoleOptions) {
				aro.RoleSessionName = "cloudbeat-remote-role"
				aro.Duration = cfg.Cred.AssumeRoleDuration
				aro.ExternalID = aws.String(CloudConnectorsExternalID(cfg.CloudConnectorsConfig.ResourceID, cfg.Cred.ExternalID))
			},
		},
	}

	retConf := AWSConfigRoleChaining(*awsConfig, chain)
	retConf.Retryer = awsConfigRetrier

	return retConf, nil
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

func LoadDefaultConfigWithRegion(ctx context.Context, beatsConfig config.AwsConfig) (*aws.Config, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load default AWS config: %w", err)
	}

	if awsConfig.Region == "" {
		if beatsConfig.Cred.DefaultRegion != "" {
			awsConfig.Region = beatsConfig.Cred.DefaultRegion
		} else {
			awsConfig.Region = "us-east-1"
		}
	}

	return &awsConfig, nil
}
