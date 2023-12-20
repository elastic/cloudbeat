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

package fetchers

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
)

type IAMFetcher struct {
	log           *logp.Logger
	iamProvider   iam.AccessManagement
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *cloud.Identity
}

type IAMFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type IAMResource struct {
	awslib.AwsResource
	identity *cloud.Identity
}

func NewIAMFetcher(log *logp.Logger, provider iam.AccessManagement, ch chan fetching.ResourceInfo, identity *cloud.Identity) *IAMFetcher {
	return &IAMFetcher{
		log:           log,
		iamProvider:   provider,
		resourceCh:    ch,
		cloudIdentity: identity,
	}
}

// Fetch collects IAM resources, such as password-policy and IAM users.
// The resources are enriched by the provider and being send to evaluation.
func (f IAMFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
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

	policies, err := f.iamProvider.GetPolicies(ctx)
	if err != nil {
		f.log.Errorf("Unable to fetch IAM policies, error: %v", err)
	} else {
		iamResources = append(iamResources, policies...)
	}

	serverCertificates, err := f.iamProvider.ListServerCertificates(ctx)
	if err != nil {
		f.log.Errorf("Unable to fetch IAM server certificates, error: %v", err)
	} else {
		iamResources = append(iamResources, serverCertificates)
	}

	accessAnalyzers, err := f.iamProvider.GetAccessAnalyzers(ctx)
	if err != nil {
		f.log.Errorf("Unable to fetch access access analyzers, error: %v", err)
	} else {
		iamResources = append(iamResources, accessAnalyzers)
	}

	for _, iamResource := range iamResources {
		f.resourceCh <- fetching.ResourceInfo{
			Resource: IAMResource{
				AwsResource: iamResource,
				identity:    f.cloudIdentity,
			},
			CycleMetadata: cycleMetadata,
		}
	}

	return nil
}

func (f IAMFetcher) Stop() {}

func (r IAMResource) GetData() any {
	return r.AwsResource
}

func (r IAMResource) GetMetadata() (fetching.ResourceMetadata, error) {
	identifier := r.GetResourceArn()
	if identifier == "" {
		identifier = fmt.Sprintf("%s-%s", r.identity.Account, r.GetResourceName())
	}

	return fetching.ResourceMetadata{
		ID:      identifier,
		Type:    fetching.CloudIdentity,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
		Region:  r.GetRegion(),
	}, nil
}

func (r IAMResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud.service.name": "IAM",
	}, nil
}
