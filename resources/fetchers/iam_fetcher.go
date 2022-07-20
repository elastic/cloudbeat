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

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

const (
	iamResourceType    = "identity-access-management"
	iamSubResourceType = "aws-role-policy"
)

type IAMFetcher struct {
	log         *logp.Logger
	iamProvider awslib.IAMRolePermissionGetter
	cfg         IAMFetcherConfig
	resourceCh  chan fetching.ResourceInfo
}

type IAMFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
	RoleName                      string `config:"roleName"`
}

type IAMResource struct {
	awslib.RolePolicyInfo
}

func (f IAMFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting IAMFetcher.Fetch")

	results, err := f.iamProvider.GetIAMRolePermissions(ctx, f.cfg.RoleName)
	if err != nil {
		return err
	}

	for _, rolePermission := range results {
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      IAMResource{rolePermission},
			CycleMetadata: cMetadata,
		}
	}

	return nil
}

func (f IAMFetcher) Stop() {
}

func (r IAMResource) GetData() interface{} {
	return r
}

func (r IAMResource) GetMetadata() fetching.ResourceMetadata {
	return fetching.ResourceMetadata{
		ID:      r.PolicyARN,
		Type:    iamResourceType,
		SubType: iamSubResourceType,
		Name:    *r.PolicyName,
	}
}
func (r IAMResource) GetElasticCommonData() any { return nil }
