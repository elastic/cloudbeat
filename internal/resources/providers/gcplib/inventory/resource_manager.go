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

package inventory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
)

type ResourceManagerWrapper struct {
	config                     auth.GcpFactoryConfig
	accountNamesCache          sync.Map                                        // cache for CloudAccountMetadata
	getProjectDisplayName      func(ctx context.Context, parent string) string // returns project display name or an empty string
	getOrganizationDisplayName func(ctx context.Context, parent string) string // returns org display name or an empty string
}

func NewResourceManagerWrapper(ctx context.Context, log *clog.Logger, gcpConfig auth.GcpFactoryConfig) (*ResourceManagerWrapper, error) {
	var gcpClientOpt []option.ClientOption
	gcpClientOpt = append(append(gcpClientOpt, option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope)), gcpConfig.ClientOpts...)
	crmService, err := cloudresourcemanager.NewService(ctx, gcpClientOpt...)
	if err != nil {
		return nil, err
	}

	return &ResourceManagerWrapper{
		config:            gcpConfig,
		accountNamesCache: sync.Map{},

		// fetches GCP Project and Organization display names; errors are ignored as they're non-critical
		getProjectDisplayName: func(ctx context.Context, parent string) string {
			prj, err := crmService.Projects.Get(parent).Context(ctx).Do()
			if err != nil {
				log.Errorf("error fetching GCP Project: %s, error: %s", parent, err)
				return ""
			}
			return prj.DisplayName
		},
		getOrganizationDisplayName: func(ctx context.Context, parent string) string {
			org, err := crmService.Organizations.Get(parent).Context(ctx).Do()
			if err != nil {
				log.Errorf("error fetching GCP Org: %s, error: %s", parent, err)
				return ""
			}
			return org.DisplayName
		},
	}, nil
}

func (c *ResourceManagerWrapper) GetCloudMetadata(ctx context.Context, asset *assetpb.Asset) *fetching.CloudAccountMetadata {
	orgId := getOrganizationId(asset.Ancestors)
	projectId := getProjectId(asset.Ancestors)
	key := fmt.Sprintf("%s/%s", projectId, orgId)
	cloudAccount, ok := c.accountNamesCache.Load(key)
	if ok {
		cloudAccountMetadata, valid := cloudAccount.(*fetching.CloudAccountMetadata)
		if valid {
			return cloudAccountMetadata
		}
	}
	cloudAccountMetadata := c.getMetadata(ctx, orgId, projectId)
	c.accountNamesCache.Store(key, cloudAccountMetadata)
	return cloudAccountMetadata
}

func (c *ResourceManagerWrapper) getMetadata(ctx context.Context, orgId string, projectId string) *fetching.CloudAccountMetadata {
	var orgName string
	var projectName string
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if isOrganization(c.config.Parent) {
			orgName = c.getOrganizationDisplayName(ctx, fmt.Sprintf("organizations/%s", orgId))
		}
	}()
	go func() {
		defer wg.Done()
		// some assets are not associated with a project
		if projectId != "" {
			projectName = c.getProjectDisplayName(ctx, fmt.Sprintf("projects/%s", projectId))
		}
	}()
	wg.Wait()

	return &fetching.CloudAccountMetadata{
		AccountId:        projectId,
		AccountName:      projectName,
		OrganisationId:   orgId,
		OrganizationName: orgName,
	}
}

func (c *ResourceManagerWrapper) Clear() {
	c.accountNamesCache.Clear()
}

func getOrganizationId(ancestors []string) string {
	last := ancestors[len(ancestors)-1]
	parts := strings.Split(last, "/") // organizations/1234567890

	if parts[0] == "organizations" {
		return parts[1]
	}

	return ""
}

func getProjectId(ancestors []string) string {
	parts := strings.Split(ancestors[0], "/") // projects/1234567890

	if parts[0] == "projects" {
		return parts[1]
	}

	return ""
}
