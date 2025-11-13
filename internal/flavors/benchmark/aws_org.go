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

package benchmark

import (
	"context"
	"errors"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/preset"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

const (
	scanSettingTagKey   = "cloudbeat_scan_management_account"
	scanSettingTagValue = "Yes"
	scopeName           = "github.com/elastic/cloudbeat/internal/flavors/benchmark/aws_org"
)

var tracer = otel.Tracer(scopeName)

type AWSOrg struct {
	IAMProvider       iam.RoleGetter
	IdentityProvider  awslib.IdentityProviderGetter
	AccountProvider   awslib.AccountProviderAPI
	StatusHandler     statushandler.StatusHandlerAPI
	AWSCredsValidator awslib.CredentialsValidator
	RoleNamesProvider awslib.OrgIAMRoleNamesProvider
}

func (a *AWSOrg) NewBenchmark(ctx context.Context, log *clog.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, _, err := a.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
	).Build(ctx, log, cfg, resourceCh, reg, a.StatusHandler)
}

//revive:disable-next-line:function-result-limit
func (a *AWSOrg) initialize(ctx context.Context, log *clog.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := a.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	log.Infof("initializing benchmark aws org (cloud connectors %t)", cfg.CloudConfig.Aws.CloudConnectors)

	var (
		awsConfigCloudbeatRoot *awssdk.Config
		awsIdentity            *cloud.Identity
		err                    error
	)

	// cloudbeat-root role credentials.
	awsConfigCloudbeatRoot, awsIdentity, err = a.getIdentity(ctx, cfg, log)
	if err != nil && cfg.CloudConfig.Aws.Cred.DefaultRegion == "" {
		log.Warn("failed to initialize identity; retrying to check AWS Gov Cloud regions")
		cfg.CloudConfig.Aws.Cred.DefaultRegion = awslib.DefaultGovRegion
		awsConfigCloudbeatRoot, awsIdentity, err = a.getIdentity(ctx, cfg, log)
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get AWS Identity: %w", err)
	}
	log.Info("successfully retrieved AWS Identity")

	// IAMProvider which is iam.RoleGetter should be created using cloudbeat-root role credentials (requires iam:GetRole).
	a.IAMProvider = iam.NewIAMProvider(ctx, log, *awsConfigCloudbeatRoot, nil)

	cache := make(map[string]registry.FetchersMap)
	reg := registry.NewRegistry(log, registry.WithUpdater(
		func(ctx context.Context) (registry.FetchersMap, error) {
			ctx, span := tracer.Start(ctx, "benchmark.AWSOrg.initialize")
			defer span.End()
			spannedLog := log.WithSpanContext(span.SpanContext())

			accounts, err := a.getAwsAccounts(ctx, spannedLog, *awsConfigCloudbeatRoot, awsIdentity)
			if err != nil {
				return nil, observability.FailSpan(span, "failed to get AWS accounts", err)
			}

			// Filter the accounts to the ones having valid credentials on each aws account.
			// Meaning only the accounts that have the security audit role created and thus were selected by customer on cloud formation.
			filtered := lo.Filter(accounts, func(item preset.AwsAccount, _ int) bool {
				return a.AWSCredsValidator.Validate(ctx, item.Config, log)
			})

			fm := preset.NewCisAwsOrganizationFetchers(ctx, spannedLog, ch, filtered, cache, a.StatusHandler)
			m := make(registry.FetchersMap)
			for accountId, fetchersMap := range fm {
				for key, fetcher := range fetchersMap {
					m[fmt.Sprintf("%s-%s", accountId, key)] = fetcher
				}
			}

			return m, nil
		}))

	return reg, cloud.NewDataProvider(cloud.WithAccount(*awsIdentity)), nil, nil
}

// getAwsAccounts returns all the aws accounts of the org.
// For each account it bundles together the cloud.Identity and the credentials for the cloudbeat-securityaudit role of that account.
// It requires cloudbeat-root credentials (requires organizations:ListAccounts, organizations:ListParents, organizations:DescribeOrganizationalUnit and iam:GetRole).
func (a *AWSOrg) getAwsAccounts(ctx context.Context, log *clog.Logger, cfgCloudbeatRoot awssdk.Config, rootIdentity *cloud.Identity) ([]preset.AwsAccount, error) {
	stsClient := sts.NewFromConfig(cfgCloudbeatRoot)

	// accountIdentities array contains all the Accounts and Organizational
	// Units, even if they are nested. (requires organizations:ListAccounts, organizations:ListParents, organizations:DescribeOrganizationalUnit)
	accountIdentities, err := a.AccountProvider.ListAccounts(ctx, log, cfgCloudbeatRoot)
	if err != nil {
		return nil, err
	}

	accounts := make([]preset.AwsAccount, 0, len(accountIdentities))
	for _, identity := range accountIdentities {
		// Cloudbeat fetchers will try to assume memberRole
		// ("cloudbeat-securityaudit") for all Accounts and OUs except for the
		// Management Account. However, Cloud Formation only creates the
		// memberRole in the OUs chosen by the user. If Cloudbeat tries to
		// assume a member role that doesn't exist (because the user hasn't
		// selected an Account/OU), it will fail silently and will be unable to
		// retrieve any resources from the Account/OU afterward.
		var awsConfig awssdk.Config

		if identity.Account == rootIdentity.Account {
			cfg, err := a.pickManagementAccountRole(ctx, log, stsClient, cfgCloudbeatRoot, identity)
			if err != nil {
				log.Errorf("error picking roles for account %s: %s", identity.Account, err)
				continue
			}
			awsConfig = cfg
		} else {
			// Try to assume "cloudbeat-security" and fail silently if it does
			// not exist.
			awsConfig = assumeRole(
				stsClient,
				cfgCloudbeatRoot,
				fmtIAMRole(identity.Account, a.RoleNamesProvider.MemberRoleName()),
			)
		}

		accounts = append(accounts, preset.AwsAccount{
			Identity: identity,
			Config:   awsConfig,
		})
	}
	return accounts, nil
}

// pickManagementAccountRole selects role used to fetch resources from the
// Management Account (and decides if they should be fetched at all).
func (a *AWSOrg) pickManagementAccountRole(ctx context.Context, log *clog.Logger, stsClient stscreds.AssumeRoleAPIClient, rootCfg awssdk.Config, identity cloud.Identity) (awssdk.Config, error) {
	// We will check for a tag on 'cloudbeat-root' role. If it is missing, we
	// will try to be backward compatible and use the "cloudbeat-root" role to
	// scan the Management Account. In previous CF templates, "cloudbeat-root"
	// had the built-in SecurityAudit policy attached.
	var foundTagValue string
	{
		r, err := a.IAMProvider.GetRole(ctx, a.RoleNamesProvider.RootRoleName())
		if err != nil {
			return awssdk.Config{}, fmt.Errorf("error getting root role: %w", err)
		}

		for _, tag := range r.Tags {
			if pointers.Deref(tag.Key) == scanSettingTagKey {
				foundTagValue = pointers.Deref(tag.Value)
				break
			}
		}
	}

	if foundTagValue == "" {
		// Legacy. Use 'cloudbeat-root' role for compliance reasons.
		log.Infof("%q tag not found, using '%s' role for backward compatibility", scanSettingTagKey, a.RoleNamesProvider.RootRoleName())
		return rootCfg, nil
	}

	// Log an error if 'cloudbeat-securityaudit' does not exist in the
	// Management Account. This should not happen! We log and continue
	// without exiting function, since we want to scan other selected
	// accounts, but at least the error will be visible in the logs.
	if foundTagValue == scanSettingTagValue {
		_, err := a.IAMProvider.GetRole(ctx, a.RoleNamesProvider.MemberRoleName())
		if err != nil {
			log.Errorf("Management Account should be scanned (%s: %s), but %q role is missing: %s", scanSettingTagKey, foundTagValue, a.RoleNamesProvider.MemberRoleName(), err)
		}
	}

	// If the "cloudbeat_scan_management_account" tag on the "cloudbeat-root"
	// role is set to "Yes", the user chose to scan it, and there should be a
	// "cloudbeat-securityaudit" role enabling this. If it is set to "No" we
	// will still try to use "cloudbeat-securityaudit", but it is non-existent,
	// so we will fail silently and not get any data from the Management
	// Account.
	log.Debugf("assuming '%s' role for Account %s", a.RoleNamesProvider.MemberRoleName(), identity.Account)
	config := assumeRole(
		stsClient,
		rootCfg,
		fmtIAMRole(identity.Account, a.RoleNamesProvider.MemberRoleName()),
	)
	return config, nil
}

// getIdentity should assume the cloudbeat-root role and then perform the GetIdentity
// and return the aws config (having credentials from cloudbeat-root role) and cloud.Identity data.
func (a *AWSOrg) getIdentity(ctx context.Context, cfg *config.Config, log *clog.Logger) (*awssdk.Config, *cloud.Identity, error) {
	awsConfig, err := a.getInitialAWSConfig(ctx, cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	// Ensure cloudbeat-root role credentials.
	// Depending on case we might have or not have the GetRole policy, or we might already own the cloudbeat-root role,
	// case A [EC2 Instance]: in this case we have already assumed cloudbeat-root role automatically because of InstanceProfile.
	// case B [Direct Credentials]: in this case we have not assumed the cloudbeat-root role but we have the same policies with cloudbeat-root role, added to user.
	// case C [Cloud Connectors]: in this case we have not assumed the cloudbeat-root role nor have the same policies with cloudbeat-root.
	// So we will try to infer cloudbeat-root role ARN by using the same account id with our current identity.
	identity, err := a.IdentityProvider.GetCallerIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AWS credentials, failed to call GetCallerIdentity: %w", err)
	}

	var cfgCloudbeatRoot awssdk.Config

	if strings.Contains(pointers.Deref(identity.Arn), a.RoleNamesProvider.RootRoleName()) {
		// case A [EC2 Instance] already cloudbeat-root, no need to re-assume.
		cfgCloudbeatRoot = *awsConfig
	} else {
		cfgCloudbeatRoot = assumeRole(
			sts.NewFromConfig(*awsConfig),
			*awsConfig,
			fmtIAMRole(pointers.Deref(identity.Account), a.RoleNamesProvider.RootRoleName()),
		)
	}

	// the next operation requires cloudbeat-root role (requires iam:ListAccountAliases policy).
	awsIdentity, err := a.IdentityProvider.GetIdentity(ctx, cfgCloudbeatRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	return &cfgCloudbeatRoot, awsIdentity, nil
}

// getInitialAWSConfig return the initial aws.Config based on the received configuration.
func (a *AWSOrg) getInitialAWSConfig(ctx context.Context, cfg *config.Config, log *clog.Logger) (*awssdk.Config, error) {
	if cfg.CloudConfig.Aws.CloudConnectors {
		// [Cloud Connectors] On cloud connectors this ends up assuming the customer remote role (using role chaining)
		return awslib.InitializeAWSConfigCloudConnectors(ctx, cfg.CloudConfig.Aws)
	}

	// [EC2 Instance] On EC2 created with our cloud formation, the identity is inferred by the EC2 instance InstanceProfile which has the cloudbeat-root role.
	// [Direct Credentials] On Direct credentials this identity is the user created by the cloud formation.
	// [Custom Setup] On custom setup like manual authentication for organization-level onboarding.
	return awslib.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred, log.Logger)
}

func (a *AWSOrg) checkDependencies() error {
	if a.IAMProvider == nil {
		return errors.New("aws iam provider is uninitialized")
	}
	if a.IdentityProvider == nil {
		return errors.New("aws identity provider is uninitialized")
	}
	if a.AccountProvider == nil {
		return errors.New("aws account provider is uninitialized")
	}
	return nil
}

func assumeRole(client stscreds.AssumeRoleAPIClient, cfg awssdk.Config, arn string) awssdk.Config {
	cfg.Credentials = awssdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(client, arn))
	return cfg
}

func fmtIAMRole(account string, role string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", account, role)
}
