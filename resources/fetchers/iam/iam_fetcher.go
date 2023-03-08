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

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

const Type = "aws-iam"

type (
	IAMFetcher struct {
		log           *logp.Logger
		iamProvider   iam.AccessManagement
		cfg           IAMFetcherConfig
		resourceCh    chan fetching.ResourceInfo
		cloudIdentity *awslib.Identity
	}

	IAMFetcherConfig struct {
		fetching.AwsBaseFetcherConfig `config:",inline"`
	}
)

func New(options ...Option) *IAMFetcher {
	f := &IAMFetcher{}
	for _, opt := range options {
		opt(f)
	}
	return f
}

// Fetch collects IAM resources, such as password-policy and IAM users.
// The resources are enriched by the provider and being send to evaluation.
func (f IAMFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting IAMFetcher.Fetch")
	iamResources := make([]awslib.AwsResource, 0)

	pwdPolicy, err := f.iamProvider.GetPasswordPolicy(ctx)
	if err != nil {
		f.log.Errorf("Unable to fetch PasswordPolicy, error: %v", err)
	} else {
		iamResources = append(iamResources, pwdPolicy)
	}

	users, err := f.iamProvider.GetUsers(ctx)
	if err != nil {
		f.log.Errorf("Unable to fetch IAM users, error: %v", err)
	} else {
		iamResources = append(iamResources, users...)
	}

	for _, iamResource := range iamResources {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: IAMResource{
				AwsResource: iamResource,
				identity:    f.cloudIdentity,
			},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f IAMFetcher) Stop() {}
