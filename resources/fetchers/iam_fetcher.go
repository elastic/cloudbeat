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
	"github.com/elastic/cloudbeat/resources/aws_providers"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type IAMFetcher struct {
	iamProvider *aws_providers.IAMProvider
	cfg         IAMFetcherConfig
}

type IAMFetcherConfig struct {
	fetching.BaseFetcherConfig
	RoleName string `config:"roleName"`
}

type IAMResource struct {
	Data interface{}
}

func (f IAMFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	results := make([]fetching.Resource, 0)

	result, err := f.iamProvider.GetIAMRolePermissions(ctx, f.cfg.RoleName)
	results = append(results, IAMResource{result})

	return results, err
}

func (f IAMFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r IAMResource) GetID() (string, error) {
	return "", nil
}

func (r IAMResource) GetData() interface{} {
	return r.Data
}
