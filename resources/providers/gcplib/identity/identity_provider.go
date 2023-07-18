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

package gcplib

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/elastic-agent-libs/logp"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"

	gcpdataprovider "github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	gcplib "github.com/elastic/cloudbeat/resources/providers/gcplib/auth"

	"github.com/elastic/cloudbeat/config"
)

const provider = "gcp"

type IdentityProviderGetter interface {
	GetIdentity(ctx context.Context, cfg config.GcpConfig) (*gcpdataprovider.Identity, error)
}

type organizationInfo struct {
	id   string
	name string
}

type IdentityProvider struct {
	service ResourceManager
	logger  *logp.Logger
}

// CloudResourceManagerService is a wrapper around the GCP resource manager service to make it easier to mock
type CloudResourceManagerService struct {
	service *cloudresourcemanager.Service
}

type ResourceManager interface {
	projectsGet(context.Context, string) (*cloudresourcemanager.Project, error)
	foldersGet(context.Context, string) (*cloudresourcemanager.Folder, error)
	organizationsGet(context.Context, string) (*cloudresourcemanager.Organization, error)
	organizationsSearch(context.Context) (*cloudresourcemanager.Organization, error)
}

func NewIdentityProvider(ctx context.Context, cfg *config.Config, logger *logp.Logger) *IdentityProvider {
	gcpClientOpt, err := gcplib.GetGcpClientConfig(cfg, logger)
	if err != nil {
		logger.Errorf("failed to get GCP client config: %v", err)
		return nil
	}
	gcpClientOpt = append(gcpClientOpt, option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope))
	crmService, err := cloudresourcemanager.NewService(ctx, gcpClientOpt...)
	if err != nil {
		logger.Errorf("failed to create GCP resource manager service: %v", err)
		return nil
	}

	return &IdentityProvider{
		service: &CloudResourceManagerService{service: crmService},
		logger:  logger,
	}
}

// GetIdentity returns GCP identity information
func (p *IdentityProvider) GetIdentity(ctx context.Context, cfg config.GcpConfig) (*gcpdataprovider.Identity, error) {
	proj, err := p.service.projectsGet(ctx, "projects/"+cfg.ProjectId)
	if err != nil {
		return nil, err
	}

	identity := &gcpdataprovider.Identity{
		Provider:    provider,
		ProjectId:   proj.ProjectId,
		ProjectName: proj.DisplayName,
	}

	// Check if the project has a parent.
	var orgInfo *organizationInfo
	if proj.Parent != "" {
		// Start recursive traversal to handle nested folders or organization.
		orgInfo, err = p.traverseResourceHierarchy(ctx, proj.Parent)
		if err != nil {
			// In case of an error, we try to search for the organization the user has access to.
			// It's used as a fallback, as user might have access to multiple organizations.
			org, err := p.service.organizationsSearch(ctx)
			if err != nil {
				p.logger.Errorf("failed to search for organization: %v", err)
				return identity, nil
			}

			if org != nil {
				identity.Account = getResourceIDFromName(org.Name)
				identity.AccountAlias = org.DisplayName
			}

			return identity, nil
		}

		if orgInfo != nil {
			identity.Account = orgInfo.id
			identity.AccountAlias = orgInfo.name
		}
	}

	return identity, nil
}

// traverseResourceHierarchy recursively traverses the resource hierarchy.
func (p *IdentityProvider) traverseResourceHierarchy(ctx context.Context, parent string) (*organizationInfo, error) {
	if parent == "" {
		fmt.Println("The project is not associated with any organization or folder.")
		return nil, nil
	}

	if isOrganization(parent) {
		organizationID := getResourceIDFromName(parent)
		orgInfo := &organizationInfo{
			id: organizationID,
		}
		organization, err := p.service.organizationsGet(ctx, organizationID)
		p.logger.Errorf("failed to get organization details: %v", err)
		if err != nil {
			return orgInfo, nil
		}

		orgInfo.name = organization.DisplayName
		return orgInfo, nil
	}

	// If the resource is a folder, fetch folder details and continue the recursion.
	folder, err := p.service.foldersGet(ctx, parent)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder details: %v", err)
	}

	// Continue the recursion with the parent folder ID.
	return p.traverseResourceHierarchy(ctx, folder.Parent)
}

func (p *CloudResourceManagerService) projectsGet(ctx context.Context, id string) (*cloudresourcemanager.Project, error) {
	project, err := p.service.Projects.Get(id).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get project with id '%v': %v", id, err)
	}
	return project, nil
}

func (p *CloudResourceManagerService) organizationsGet(ctx context.Context, name string) (*cloudresourcemanager.Organization, error) {
	org, err := p.service.Organizations.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get oragnization with name '%v': %v", name, err)
	}
	return org, nil
}

func (p *CloudResourceManagerService) organizationsSearch(ctx context.Context) (*cloudresourcemanager.Organization, error) {
	res, err := p.service.Organizations.Search().Context(ctx).PageSize(1).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get oragnization: %v", err)
	}

	if len(res.Organizations) == 0 {
		return nil, nil
	}

	return res.Organizations[0], nil
}

func (p *CloudResourceManagerService) foldersGet(ctx context.Context, name string) (*cloudresourcemanager.Folder, error) {
	folder, err := p.service.Folders.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get oragnization with name '%v': %v", name, err)
	}
	return folder, nil
}

// isOrganization checks if the resource name corresponds to an organization.
func isOrganization(resource string) bool {
	return strings.HasPrefix(resource, "organizations/")
}

// getResourceIDFromName extracts the resource ID from the resource name.
func getResourceIDFromName(resource string) string {
	parts := strings.Split(resource, "/")
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
