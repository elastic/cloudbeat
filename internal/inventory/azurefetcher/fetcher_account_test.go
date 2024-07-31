package azurefetcher

import (
	"testing"

	"github.com/elastic/cloudbeat/internal/inventory"
	azurelib_inventory "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
)

func TestAccountFetcher_Fetch_Tenants(t *testing.T) {
	azureAssets := []azurelib_inventory.AzureAsset{
		{
			Id:          "/tenants/<tenant UUID>",
			Name:        "<tenant UUID>",
			DisplayName: "Mario",
			TenantId:    "<tenant UUID>",
		},
	}
	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureTenant,
			[]string{"/tenants/<tenant UUID>"},
			"Mario",
			inventory.WithRawAsset(azureAssets[0]),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: "<tenant UUID>",
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		),
	}

	// setup
	logger := logp.NewLogger("azurefetcher_test")
	provider := newMockSubscriptionProvider(t)
	provider.EXPECT().ListTenants(mock.Anything).Return(azureAssets, nil)
	provider.EXPECT().ListSubscriptions(mock.Anything).Return(nil, nil)
	fetcher := newAccountFetcher(logger, provider)
	// test & compare
	collectResourcesAndMatch(t, fetcher, expected)
}
