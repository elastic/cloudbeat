package azurefetcher

import (
	"context"

	"github.com/elastic/cloudbeat/internal/inventory"
	azurelib "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
	"github.com/elastic/elastic-agent-libs/logp"
)

type accountFetcher struct {
	logger   *logp.Logger
	provider subscriptionProvider
}

type (
	subscriptionProviderFunc func(context.Context) ([]azurelib.AzureAsset, error)
	subscriptionProvider     interface {
		ListTenants(ctx context.Context) ([]azurelib.AzureAsset, error)
		ListSubscriptions(ctx context.Context) ([]azurelib.AzureAsset, error)
	}
)

func newAccountFetcher(logger *logp.Logger, provider subscriptionProvider) inventory.AssetFetcher {
	return &accountFetcher{
		logger:   logger,
		provider: provider,
	}
}

func (f *accountFetcher) Fetch(ctx context.Context, assetChan chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       subscriptionProviderFunc
		classification inventory.AssetClassification
	}{
		{"Tenants", f.provider.ListTenants, inventory.AssetClassificationAzureTenant},
		{"Subscriptions", f.provider.ListSubscriptions, inventory.AssetClassificationAzureSubscription},
	}
	for _, r := range resourcesToFetch {
		f.fetch(ctx, r.name, r.function, r.classification, assetChan)
	}
}

func (f *accountFetcher) fetch(ctx context.Context, resourceName string, function subscriptionProviderFunc, classification inventory.AssetClassification, assetChan chan<- inventory.AssetEvent) {
	f.logger.Infof("Fetching %s", resourceName)
	defer f.logger.Infof("Fetching %s - Finished", resourceName)

	azureAssets, err := function(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch %s: %w", resourceName, err)
		return
	}

	for _, item := range azureAssets {
		assetChan <- inventory.NewAssetEvent(
			classification,
			[]string{item.Id},
			item.DisplayName,
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: item.TenantId,
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		)
	}
}
