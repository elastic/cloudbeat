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

	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

// defaultProjectParentResolver resolves project parent from config, audience, or ADC.
type defaultProjectParentResolver struct {
	auth DefaultCredentialsFinder
}

// NewDefaultProjectParentResolver returns a ProjectParentResolver that uses the given
// credentials finder for the ADC fallback when project_id is not in config or audience.
func NewDefaultProjectParentResolver(auth DefaultCredentialsFinder) ProjectParentResolver {
	return &defaultProjectParentResolver{auth: auth}
}

func (r *defaultProjectParentResolver) GetProjectParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error) {
	if cfg.ProjectId != "" {
		return fmt.Sprintf("projects/%s", cfg.ProjectId), nil
	}
	if len(clientOpts) > 0 && cfg.Audience != "" {
		projectId, err := resolveProjectIdFromAudience(ctx, clientOpts, cfg.Audience)
		if err != nil {
			return "", fmt.Errorf("failed to get project ID: %w", err)
		}
		return fmt.Sprintf("projects/%s", projectId), nil
	}
	cred, err := r.auth.FindDefaultCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get project ID: %w", err)
	}
	if cred.ProjectID == "" {
		return "", ErrProjectNotFound
	}
	return fmt.Sprintf("projects/%s", cred.ProjectID), nil
}

// resolveProjectIdFromAudience uses the Cloud Resource Manager v3 API to resolve
// the project ID (user-assigned string) from the project number in the audience.
// clientOpts must authenticate as an identity with resourcemanager.projects.get.
func resolveProjectIdFromAudience(ctx context.Context, clientOpts []option.ClientOption, audience string) (string, error) {
	projectNumber, ok := projectNumberFromAudience(audience)
	if !ok {
		return "", fmt.Errorf("audience does not contain a valid project number (expected format: //iam.googleapis.com/projects/PROJECT_NUMBER/...): %w", ErrProjectNotFound)
	}
	opts := append([]option.ClientOption{option.WithScopes(cloudresourcemanager.CloudPlatformReadOnlyScope)}, clientOpts...)
	svc, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create Resource Manager client: %w", err)
	}
	project, err := svc.Projects.Get("projects/" + projectNumber).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get project (number %s): %w", projectNumber, err)
	}
	if project.ProjectId == "" {
		return "", fmt.Errorf("project response missing project_id: %w", ErrProjectNotFound)
	}
	return project.ProjectId, nil
}
