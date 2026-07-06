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

	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

// ProjectParentResolver resolves the GCP project parent string (e.g. "projects/project-id").
// Defined here (consumer) per Go best practice: interfaces belong in the package that uses them.
type ProjectParentResolver interface {
	GetProjectParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error)
}

// OrganizationParentResolver resolves the GCP organization parent string (e.g. "organizations/org-id").
// Defined here (consumer) per Go best practice: interfaces belong in the package that uses them.
type OrganizationParentResolver interface {
	GetOrganizationParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error)
}

// defaultParentResolver is a facade that delegates to ProjectParentResolver or
// OrganizationParentResolver based on cfg.AccountType.
type defaultParentResolver struct {
	project ProjectParentResolver
	org     OrganizationParentResolver
}

// NewDefaultParentResolver returns a ParentResolver that delegates to the default
// project and organization resolvers. The auth finder is used by the project resolver
// for the ADC fallback when project_id is not in config or audience.
func NewDefaultParentResolver(auth DefaultCredentialsFinder) ParentResolver {
	return &defaultParentResolver{
		project: NewDefaultProjectParentResolver(auth),
		org:     NewDefaultOrganizationParentResolver(),
	}
}

func (r *defaultParentResolver) GetParent(ctx context.Context, cfg config.GcpConfig, clientOpts []option.ClientOption) (string, error) {
	switch cfg.AccountType {
	case config.OrganizationAccount:
		return r.org.GetOrganizationParent(ctx, cfg, clientOpts)
	case config.SingleAccount:
		return r.project.GetProjectParent(ctx, cfg, clientOpts)
	default:
		return "", fmt.Errorf("invalid gcp account type: %s", cfg.AccountType)
	}
}
