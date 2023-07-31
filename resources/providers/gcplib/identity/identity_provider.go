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

package identity

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-libs/logp"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

const provider = "gcp"

type ProviderGetter interface {
	GetIdentity(ctx context.Context, cfg config.GcpConfig) (*cloud.Identity, error)
}

type Provider struct {
	service ResourceManager
	logger  *logp.Logger
}

// CloudResourceManagerService is a wrapper around the GCP resource manager service to make it easier to mock
type CloudResourceManagerService struct {
	service *cloudresourcemanager.Service
}

type ResourceManager interface {
	projectsGet(context.Context, string) (*cloudresourcemanager.Project, error)
}

func NewProvider(logger *logp.Logger) *Provider {
	return &Provider{logger: logger.Named("gcp.identity")}
}

// GetIdentity returns GCP identity information
func (p *Provider) GetIdentity(ctx context.Context, cfg config.GcpConfig) (*cloud.Identity, error) {
	if err := p.initialize(ctx, cfg); err != nil {
		return nil, err
	}

	proj, err := p.service.projectsGet(ctx, "projects/"+cfg.ProjectId)
	if err != nil {
		return nil, err
	}

	return &cloud.Identity{
		Provider:    provider,
		ProjectId:   proj.ProjectId,
		ProjectName: proj.DisplayName,
	}, nil
}

func (p *Provider) initialize(ctx context.Context, cfg config.GcpConfig) error {
	if p.service != nil {
		return nil
	}

	gcpClientOpt, err := auth.GetGcpClientConfig(cfg, p.logger)
	if err != nil {
		return err
	}
	gcpClientOpt = append(gcpClientOpt, option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope))
	crmService, err := cloudresourcemanager.NewService(ctx, gcpClientOpt...)
	if err != nil {
		return err
	}

	p.service = &CloudResourceManagerService{service: crmService}
	return nil
}

func (p *CloudResourceManagerService) projectsGet(ctx context.Context, id string) (*cloudresourcemanager.Project, error) {
	project, err := p.service.Projects.Get(id).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get project with id '%v': %v", id, err)
	}
	return project, nil
}
