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

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type IAMFetcher struct {
	log           *clog.Logger
	iamProvider   iam.AccessManagement
	resourceCh    chan fetching.ResourceInfo
	cloudIdentity *cloud.Identity
	statusHandler statushandler.StatusHandlerAPI
}

type IAMFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type IAMResource struct {
	awslib.AwsResource
	identity *cloud.Identity
}

func NewIAMFetcher(log *clog.Logger, provider iam.AccessManagement, ch chan fetching.ResourceInfo, identity *cloud.Identity, statusHandler statushandler.StatusHandlerAPI) *IAMFetcher {
	return &IAMFetcher{
		log:           log,
		iamProvider:   provider,
		resourceCh:    ch,
		cloudIdentity: identity,
		statusHandler: statusHandler,
	}
}

// Fetch collects IAM resources, such as password-policy and IAM users.
// The resources are enriched by the provider and being send to evaluation.
func (f IAMFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Debug("Starting IAMFetcher.Fetch")
	iamResources := make([]awslib.AwsResource, 0)

	pwdPolicy, err := f.iamProvider.GetPasswordPolicy(ctx)
	if err != nil {
		f.log.Errorf(ctx, "Unable to fetch PasswordPolicy, error: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	} else {
		iamResources = append(iamResources, pwdPolicy)
	}

	users, err := f.iamProvider.GetUsers(ctx)
	if err != nil {
		f.log.Errorf(ctx, "Unable to fetch IAM users, error: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	} else {
		iamResources = append(iamResources, users...)
	}

	policies, err := f.iamProvider.GetPolicies(ctx)
	if err != nil {
		f.log.Errorf(ctx, "Unable to fetch IAM policies, error: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	} else {
		iamResources = append(iamResources, policies...)
	}

	serverCertificates, err := f.iamProvider.ListServerCertificates(ctx)
	if err != nil {
		f.log.Errorf(ctx, "Unable to fetch IAM server certificates, error: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	} else {
		iamResources = append(iamResources, serverCertificates)
	}

	accessAnalyzers, err := f.iamProvider.GetAccessAnalyzers(ctx)
	if err != nil {
		f.log.Errorf(ctx, "Unable to fetch access access analyzers, error: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
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

func (r IAMResource) GetIds() []string {
	return []string{r.GetResourceArn(), r.identity.Account}
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
	m := map[string]any{
		"cloud.service.name": "IAM",
	}

	switch r.AwsResource.GetResourceType() {
	case fetching.IAMUserType:
		{
			m["user.id"] = r.GetResourceArn()
			m["user.name"] = r.GetResourceName()
		}
	case fetching.PolicyType:
		{
			m["user.effective.id"] = r.GetResourceArn()
			m["user.effective.name"] = r.GetResourceName()
		}
	default:
	}

	return m, nil
}
