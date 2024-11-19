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

package preset

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"
	"go.uber.org/zap"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
)

type AwsAccount struct {
	cloud.Identity
	aws.Config
}

type wrapResource struct {
	wrapped  fetching.Resource
	identity cloud.Identity
}

func (w *wrapResource) GetMetadata() (fetching.ResourceMetadata, error) {
	mdata, err := w.wrapped.GetMetadata()
	if err != nil {
		return mdata, err
	}
	mdata.AccountName = w.identity.AccountAlias
	mdata.AccountId = w.identity.Account
	mdata.OrganisationId = w.identity.OrganizationId
	mdata.OrganizationName = w.identity.OrganizationName
	return mdata, nil
}

func (w *wrapResource) GetIds() []string {
	return w.wrapped.GetIds()
}

func (w *wrapResource) GetData() any { return w.wrapped.GetData() }
func (w *wrapResource) GetElasticCommonData() (map[string]any, error) {
	return w.wrapped.GetElasticCommonData()
}

func NewCisAwsOrganizationFetchers(ctx context.Context, log *logp.Logger, rootCh chan fetching.ResourceInfo, accounts []AwsAccount, cache map[string]registry.FetchersMap) map[string]registry.FetchersMap {
	return newCisAwsOrganizationFetchers(ctx, log, rootCh, accounts, cache, NewCisAwsFetchers)
}

// awsFactory is the same function type as NewCisAwsFetchers, and it's used to mock the function in tests
type awsFactory func(context.Context, *logp.Logger, aws.Config, chan fetching.ResourceInfo, *cloud.Identity) registry.FetchersMap

func newCisAwsOrganizationFetchers(
	ctx context.Context,
	log *logp.Logger,
	rootCh chan fetching.ResourceInfo,
	accounts []AwsAccount,
	cache map[string]registry.FetchersMap,
	factory awsFactory,
) map[string]registry.FetchersMap {
	seen := make(map[string]bool)
	for key := range cache {
		seen[key] = false
	}

	m := make(map[string]registry.FetchersMap)
	for _, account := range accounts {
		if existing := cache[account.Account]; existing != nil {
			m[account.Account] = existing
			seen[account.Account] = true
			continue
		}

		ch := make(chan fetching.ResourceInfo)
		go func(identity cloud.Identity) {
			for {
				select {
				case <-ctx.Done():
					return
				case resourceInfo, ok := <-ch:
					if !ok {
						return
					}

					wrappedResourceInfo := fetching.ResourceInfo{
						Resource: &wrapResource{
							wrapped:  resourceInfo.Resource,
							identity: identity,
						},
						CycleMetadata: resourceInfo.CycleMetadata,
					}

					select {
					case <-ctx.Done():
						return
					case rootCh <- wrappedResourceInfo:
					}
				}
			}
		}(account.Identity)

		f := factory(
			ctx,
			log.Named("aws").WithOptions(zap.Fields(zap.String("cloud.account.id", account.Identity.Account))),
			account.Config,
			ch,
			&account.Identity,
		)
		m[account.Account] = f
		if cache != nil {
			cache[account.Account] = f
		}
	}

	for key, v := range seen {
		if !v {
			delete(cache, key)
		}
	}

	return m
}
