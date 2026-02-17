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

package auth

import (
	"context"
	"fmt"

	crmv1 "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

// defaultOrganizationParentResolver resolves organization parent from config or audience.
type defaultOrganizationParentResolver struct{}

// NewDefaultOrganizationParentResolver returns an OrganizationParentResolver that uses
// the Resource Manager API to resolve org from audience when needed.
func NewDefaultOrganizationParentResolver() OrganizationParentResolver {
	return &defaultOrganizationParentResolver{}
}

func (r *defaultOrganizationParentResolver) GetOrganizationParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error) {
	if cfg.OrganizationId != "" {
		return fmt.Sprintf("organizations/%s", cfg.OrganizationId), nil
	}
	if len(clientOpts) > 0 && cfg.Audience != "" {
		orgId, err := resolveOrganizationIdFromAudience(ctx, clientOpts, cfg.Audience)
		if err != nil {
			return "", fmt.Errorf("failed to resolve organization ID: %w", err)
		}
		return fmt.Sprintf("organizations/%s", orgId), nil
	}
	return "", ErrMissingOrgId
}

const resourceIdTypeOrganization = "organization"

// resolveOrganizationIdFromAudience uses the project number in the audience to get
// the project via v3, then calls v1 getAncestry to resolve the organization in one call.
// clientOpts must authenticate with resourcemanager.projects.get.
func resolveOrganizationIdFromAudience(ctx context.Context, clientOpts []option.ClientOption, audience string) (string, error) {
	opts := append([]option.ClientOption{option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope)}, clientOpts...)
	svcV3, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create Resource Manager client: %w", err)
	}
	projectNumber, ok := projectNumberFromAudience(audience)
	if !ok {
		return "", fmt.Errorf("audience does not contain a valid project number: %w", ErrMissingOrgId)
	}
	project, err := svcV3.Projects.Get("projects/" + projectNumber).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get project (number %s): %w", projectNumber, err)
	}
	optsV1 := append([]option.ClientOption{option.WithScopes(crmv1.CloudPlatformReadOnlyScope)}, clientOpts...)
	svcV1, err := crmv1.NewService(ctx, optsV1...)
	if err != nil {
		return "", fmt.Errorf("failed to create Resource Manager v1 client: %w", err)
	}
	resp, err := svcV1.Projects.GetAncestry(project.ProjectId, &crmv1.GetAncestryRequest{}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get project ancestry: %w", err)
	}
	for _, anc := range resp.Ancestor {
		if anc.ResourceId != nil && anc.ResourceId.Type == resourceIdTypeOrganization && anc.ResourceId.Id != "" {
			return anc.ResourceId.Id, nil
		}
	}
	return "", fmt.Errorf("no organization found in resource hierarchy: %w", ErrMissingOrgId)
}
