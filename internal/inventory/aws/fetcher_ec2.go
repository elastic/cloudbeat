package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Ec2Fetcher struct {
	logger            *logp.Logger
	provider          *ec2.Provider
	ec2Classification inventory.AssetClassification
}

func newEc2Fetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) inventory.AssetFetcher {
	provider := ec2.NewEC2Provider(logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	return &Ec2Fetcher{
		logger:   logger,
		provider: provider,
		ec2Classification: inventory.AssetClassification{
			Category:    inventory.CategoryInfrastructure,
			SubCategory: inventory.SubCategoryCompute,
			Type:        inventory.TypeVirtualMachine,
			SubStype:    inventory.SubTypeEC2,
		},
	}
}

func (e *Ec2Fetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	instances, err := e.provider.DescribeInstances(ctx)
	if err != nil {
		e.logger.Errorf("Could not list ec2 instances (%v)", err)
		return
	}

	for _, instance := range instances {
		if instance == nil {
			continue
		}

		iamFetcher := inventory.EmptyEnricher()
		if instance.IamInstanceProfile != nil {
			iamFetcher = inventory.WithIAM(inventory.AssetIAM{
				Id:  instance.IamInstanceProfile.Id,
				Arn: instance.IamInstanceProfile.Arn,
			})
		}

		tags := make(map[string]string, len(instance.Tags))
		for _, t := range instance.Tags {
			if t.Key == nil {
				continue
			}

			tags[*t.Key] = pointers.Deref(t.Value)
		}

		assetChannel <- inventory.NewAssetEvent(
			e.ec2Classification,
			instance.GetResourceArn(),
			instance.GetResourceName(),

			inventory.WithRawAsset(instance),
			inventory.WithTags(tags),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   instance.Region,
			}),
			inventory.WithHost(inventory.AssetHost{
				Architecture:    string(instance.Architecture),
				ImageId:         instance.ImageId,
				InstanceType:    string(instance.InstanceType),
				Platform:        string(instance.Platform),
				PlatformDetails: instance.PlatformDetails,
			}),
			iamFetcher,
			inventory.WithNetwork(inventory.AssetNetwork{
				NetworkId:        instance.VpcId,
				SubnetId:         instance.SubnetId,
				Ipv6Address:      instance.Ipv6Address,
				PublicIpAddress:  instance.PublicIpAddress,
				PrivateIpAddress: instance.PrivateIpAddress,
				PublicDnsName:    instance.PublicDnsName,
				PrivateDnsName:   instance.PrivateDnsName,
			}),
		)
	}
}
