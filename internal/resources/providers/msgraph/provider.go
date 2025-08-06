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

package msgraph

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/directoryroles"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
)

type ProviderAPI interface {
	ListServicePrincipals(context.Context) ([]*models.ServicePrincipal, error)
	ListDirectoryRoles(context.Context) ([]*models.DirectoryRole, error)
	ListGroups(context.Context) ([]*models.Group, error)
	ListUsers(context.Context) ([]*models.User, error)
}

type provider struct {
	log    *clog.Logger
	client interface {
		ServicePrincipals() *serviceprincipals.ServicePrincipalsRequestBuilder
		DirectoryRoles() *directoryroles.DirectoryRolesRequestBuilder
		Groups() *groups.GroupsRequestBuilder
		Users() *users.UsersRequestBuilder
	}
}

// Second argument is scopes. Leave nil, then it selects default; Adjust if in trouble
// Docs: https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
func NewProvider(log *clog.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
	// Requires 'Directory.Read.All' API permission
	c, err := graph.NewGraphServiceClientWithCredentials(azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating MS Graph client: %w", err)
	}

	p := &provider{
		log:    log.Named("msgraph"),
		client: c,
	}

	return p, nil
}

// Docs:
// - https://github.com/microsoftgraph/msgraph-sdk-go
// - https://learn.microsoft.com/en-us/graph/api/serviceprincipal-list?view=graph-rest-beta&tabs=go
// - https://learn.microsoft.com/en-us/graph/sdks/paging?tabs=go
func (p *provider) ListServicePrincipals(ctx context.Context) ([]*models.ServicePrincipal, error) {
	requestConfig := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{}

	response, err := p.client.ServicePrincipals().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("error listing Azure Service Principals: %w", err)
	}

	pageIterator, err := graphcore.NewPageIterator[*models.ServicePrincipal](
		response,
		p.client.ServicePrincipals().RequestAdapter,
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("error paging Azure Service Principals: %w", err)
	}

	items := []*models.ServicePrincipal{}
	err = pageIterator.Iterate(ctx, func(pageItem *models.ServicePrincipal) bool {
		items = append(items, pageItem)
		return true // to continue the iteration
	})
	if err != nil {
		p.log.Errorf(ctx, "error iterating over Service Principals: %v", err)
	}
	return items, nil
}

func (p *provider) ListDirectoryRoles(ctx context.Context) ([]*models.DirectoryRole, error) {
	requestConfig := &directoryroles.DirectoryRolesRequestBuilderGetRequestConfiguration{}

	response, err := p.client.DirectoryRoles().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("error listing Azure Directory Roles: %w", err)
	}

	pageIterator, err := graphcore.NewPageIterator[*models.DirectoryRole](
		response,
		p.client.DirectoryRoles().RequestAdapter,
		models.CreateDirectoryRoleCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("error paging Azure Directory Roles: %w", err)
	}

	items := []*models.DirectoryRole{}
	err = pageIterator.Iterate(ctx, func(pageItem *models.DirectoryRole) bool {
		items = append(items, pageItem)
		return true // to continue the iteration
	})
	if err != nil {
		p.log.Errorf(ctx, "error iterating over Directory Roles: %v", err)
	}
	return items, nil
}

func (p *provider) ListGroups(ctx context.Context) ([]*models.Group, error) {
	requestConfig := &groups.GroupsRequestBuilderGetRequestConfiguration{}

	response, err := p.client.Groups().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("error listing Azure Groups: %w", err)
	}

	pageIterator, err := graphcore.NewPageIterator[*models.Group](
		response,
		p.client.Groups().RequestAdapter,
		models.CreateGroupCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("error paging Azure Groups: %w", err)
	}

	items := []*models.Group{}
	err = pageIterator.Iterate(ctx, func(pageItem *models.Group) bool {
		items = append(items, pageItem)
		return true // to continue the iteration
	})
	if err != nil {
		p.log.Errorf(ctx, "error iterating over Groups: %v", err)
	}
	return items, nil
}

func (p *provider) ListUsers(ctx context.Context) ([]*models.User, error) {
	requestConfig := &users.UsersRequestBuilderGetRequestConfiguration{}

	response, err := p.client.Users().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("error listing Azure Users: %w", err)
	}

	pageIterator, err := graphcore.NewPageIterator[*models.User](
		response,
		p.client.Users().RequestAdapter,
		models.CreateUserCollectionResponseFromDiscriminatorValue,
	)
	if err != nil {
		return nil, fmt.Errorf("error paging Azure Users: %w", err)
	}

	items := []*models.User{}
	err = pageIterator.Iterate(ctx, func(pageItem *models.User) bool {
		items = append(items, pageItem)
		return true // to continue the iteration
	})
	if err != nil {
		p.log.Errorf(ctx, "error iterating over Users: %v", err)
	}
	return items, nil
}
