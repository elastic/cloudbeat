package azurefetcher

import (
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	azure_auth "github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
)

func New(logger *logp.Logger, provider azurelib.ProviderAPI, azureConfig *azure_auth.AzureFactoryConfig) []inventory.AssetFetcher {
	// TODO(kuba): start something that returns a list of AssetFetchers
	return nil
}
