package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/elastic-agent-libs/logp"
)

func Fetchers(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) []inventory.AssetFetcher {
	return []inventory.AssetFetcher{
		newEc2Fetcher(logger, identity, cfg),
	}
}
