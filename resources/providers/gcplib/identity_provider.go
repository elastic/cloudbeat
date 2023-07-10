package gcplib

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/config"
)

type IdentityProviderGetter interface {
	GetIdentity(context.Context, []option.ClientOption, config.GcpConfig) (*Identity, error)
}

type Identity struct {
	OrgId       string
	OrgName     string
	ProjectId   string
	ProjectName string
	Provider    string
}

type IdentityProvider struct{}

// GetIdentity returns GCP identity information
func (p IdentityProvider) GetIdentity(ctx context.Context, gcpClientOpt []option.ClientOption, cfg config.GcpConfig) (*Identity, error) {
	gcpClientOpt = append(gcpClientOpt, option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope))
	crmService, err := cloudresourcemanager.NewService(ctx, gcpClientOpt...)
	if err != nil {
		return nil, err
	}

	proj, err := crmService.Projects.Get(cfg.ProjectId).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var org *cloudresourcemanager.Organization
	if proj.Parent != nil {
		if proj.Parent.Type == "organization" {
			org, err = crmService.Organizations.Get("organizations/" + proj.Parent.Id).Do()
			if err != nil {
				return nil, fmt.Errorf("failed to get GCP project organization: %v", err)
			}
		}
	}

	return &Identity{
		OrgId:       proj.Parent.Id,
		OrgName:     org.DisplayName,
		ProjectId:   proj.ProjectId,
		ProjectName: proj.Name,
		Provider:    "gcp",
	}, nil
}
