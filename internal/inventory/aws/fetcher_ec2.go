package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Ec2Fetcher struct {
	logger            *logp.Logger
	provider          *ec2.Provider
	ec2Classification inventory.AssetClassification
}

func newEc2Fetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) *Ec2Fetcher {
	provider := ec2.NewEC2Provider(logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	return &Ec2Fetcher{
		logger:   logger,
		provider: provider,
		ec2Classification: inventory.AssetClassification{
			Category:      inventory.CategoryInfrastructure,
			SubCategory:   inventory.SubCategoryCompute,
			Type:          inventory.TypeVirtualMachine,
			SubStype:      inventory.SubTypeEC2,
			CloudProvider: inventory.AwsCloudProvider,
		},
	}
}

func (e *Ec2Fetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.Asset) {
	instances, err := e.provider.DescribeInstances(ctx)
	if err != nil {
		e.logger.Errorf("Could not list ec2 instances (%v)", err)
		return
	}

	for _, instance := range instances {
		assetChannel <- inventory.NewAsset(
			e.ec2Classification,
			pointers.Deref(instance.InstanceId),
			instance.GetResourceName(),
		)
	}
}
