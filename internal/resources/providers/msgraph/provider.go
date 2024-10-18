package msgraph

import (
	"context"
	"fmt"

	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"

	"github.com/elastic/elastic-agent-libs/logp"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
)

type ProviderAPI interface {
	ListServicePrincipals(context.Context) ([]*models.ServicePrincipal, error)
}

type provider struct {
	log    *logp.Logger
	client interface {
		ServicePrincipals() *serviceprincipals.ServicePrincipalsRequestBuilder
	}
}

// Second argument is scopes. Leave nil, then it selects default; Adjust if in trouble
// Docs: https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
func NewProvider(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
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
