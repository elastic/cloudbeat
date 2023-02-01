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

package dataprovider

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	cloudAccountIdField   = "cloud.account.id"
	cloudAccountNameField = "cloud.account.name"
	cloudProviderField    = "cloud.provider"
	cloudProviderValue    = "aws"
)

type commonAwsData struct {
	accountId   string
	accountName string
}

type awsDataProvider struct {
	log              *logp.Logger
	identityProvider awslib.IdentityProviderGetter
	iamProvider      iam.AccessManagement
}

func NewAwsDataProvider(log *logp.Logger, cfg *config.Config) (EnvironmentCommonDataProvider, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identityClient := awslib.GetIdentityClient(awsConfig)
	iamProvider := iam.NewIAMProvider(log, awsConfig)

	return &awsDataProvider{log, identityClient, iamProvider}, nil
}

func (a awsDataProvider) FetchData(ctx context.Context) (CommonData, error) {
	identity, err := a.identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	alias, err := a.iamProvider.GetAccountAlias(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS account alias: %w", err)
	}

	return &commonAwsData{
		accountId:   *identity.Account,
		accountName: alias,
	}, nil
}

func (c commonAwsData) GetResourceId(metadata fetching.ResourceMetadata) string {
	return metadata.ID
}

func (c commonAwsData) GetVersionInfo() version.CloudbeatVersionInfo {
	return version.CloudbeatVersionInfo{
		Version: version.CloudbeatVersion(),
		Policy:  version.PolicyVersion(),
	}
}

func (c commonAwsData) EnrichEvent(event beat.Event) error {
	_, err := event.Fields.Put(cloudAccountIdField, c.accountId)
	if err != nil {
		return err
	}

	_, err = event.Fields.Put(cloudAccountNameField, c.accountName)
	if err != nil {
		return err
	}

	_, err = event.Fields.Put(cloudProviderField, cloudProviderValue)
	if err != nil {
		return err
	}

	return nil
}
