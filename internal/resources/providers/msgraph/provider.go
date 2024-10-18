package msgraph

import (
	"context"
	"fmt"

	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"

	"github.com/elastic/elastic-agent-libs/logp"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
)

type ProviderAPI interface{}

type provider struct {
	log    *logp.Logger
	client interface {
		ServicePrincipals() *serviceprincipals.ServicePrincipalsRequestBuilder
	}
}

func NewProvider(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
	// Second argument is scopes. Leave nil, then it selects default; Adjust if in trouble
	// Docs: https://learn.microsoft.com/en-us/graph/sdks/create-client?from=snippets&tabs=go
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

// TODO(kuba): replace return with []AzureAsset
// Docs:
// - https://github.com/microsoftgraph/msgraph-sdk-go
// - https://learn.microsoft.com/en-us/graph/api/serviceprincipal-list?view=graph-rest-beta&tabs=go
// - https://learn.microsoft.com/en-us/graph/sdks/paging?tabs=go
func (p *provider) ListServicePrincipals(ctx context.Context) ([]inventory.AzureAsset, error) {
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

	assets := []inventory.AzureAsset{}
	err = pageIterator.Iterate(ctx, func(pageItem *models.ServicePrincipal) bool {
		asset, err := p.transformServicePrincipal(pageItem)
		if err != nil {
			p.log.Errorf("error transforming Service Principal: %v", err)
			return true
		}
		assets = append(assets, asset)
		return true // to continue the iteration
	})
	if err != nil {
		p.log.Errorf("error iterating over Service Principals: %v", err)
	}
	return assets, nil
}

func (p *provider) transformServicePrincipal(item *models.ServicePrincipal) (inventory.AzureAsset, error) {

	return nil, nil
}
