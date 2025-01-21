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
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"

	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type ProviderAPI interface {
	ListServicePrincipals(context.Context) ([]*models.ServicePrincipal, error)
}

type provider struct {
	log    *clog.Logger
	client interface {
		ServicePrincipals() *serviceprincipals.ServicePrincipalsRequestBuilder
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
		p.log.Errorf("error iterating over Service Principals: %v", err)
	}
	return items, nil
}
