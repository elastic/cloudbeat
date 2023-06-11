package factory

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"
	"github.com/elastic/elastic-agent-libs/logp"
)

type cisAwsFactory struct {
}

func (c cisAwsFactory) Create() (fetching.Fetcher, error) {
	//TODO implement me
	panic("implement me")
}

func NewCisAwsFactory(log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (FetchersMap, error) {
	m := make(FetchersMap)
	ctx := context.Background()
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identity, err := awslib.GetIdentityClient(awsConfig).GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	iamProvider := iam.NewIAMProvider(log, awsConfig, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	iamFetcher := fetchers.NewIAMFetcher(log, iamProvider, ch, identity)
	m[fetching.IAMType] = iamFetcher

	kmsProvider := kms.NewKMSProvider(log, awsConfig, &awslib.MultiRegionClientFactory[kms.Client]{})
	kmsFetcher := fetchers.NewKMSFetcher(log, kmsProvider, ch)
	m[fetching.KmsType] = kmsFetcher

	loggingProvider := logging.NewProvider(log, awsConfig, &awslib.MultiRegionClientFactory[cloudtrail.Client]{}, &awslib.MultiRegionClientFactory[s3.Client]{}, *identity.Account)
	configserviceProvider := configservice.NewProvider(log, awsConfig, &awslib.MultiRegionClientFactory[configservice.Client]{}, *identity.Account)
	loggingFetcher := fetchers.NewLoggingFetcher(log, loggingProvider, configserviceProvider, ch, identity)
	m[fetching.TrailType] = loggingFetcher

	monitoringProvider := monitoring.NewProvider(
		log,
		awsConfig,
		&awslib.MultiRegionClientFactory[cloudtrail.Client]{},
		&awslib.MultiRegionClientFactory[cloudwatch.Client]{},
		&awslib.MultiRegionClientFactory[logs.Client]{},
		&awslib.MultiRegionClientFactory[sns.Client]{},
		identity,
	)

	securityHubProvider := securityhub.NewProvider(log, awsConfig, &awslib.MultiRegionClientFactory[securityhub.Client]{}, *identity.Account)
	monitoringFetcher := fetchers.NewMonitoringFetcher(log, monitoringProvider, securityHubProvider, ch, identity)
	m[fetching.MonitoringType] = monitoringFetcher

	ec2Provider := ec2.NewEC2Provider(log, *identity.Account, awsConfig, &awslib.MultiRegionClientFactory[ec2.Client]{})
	networkFetcher := fetchers.NewNetworkFetcher(log, ec2Provider, ch, identity)
	m[fetching.EC2NetworkingType] = networkFetcher

	rdsProvider := rds.NewProvider(log, awsConfig, &awslib.MultiRegionClientFactory[rds.Client]{}, ec2Provider)
	rdsFetcher := fetchers.NewRdsFetcher(log, rdsProvider, ch)
	m[fetching.RdsType] = rdsFetcher

	s3Provider := s3.NewProvider(log, awsConfig, &awslib.MultiRegionClientFactory[s3.Client]{}, *identity.Account)
	s3Fetcher := fetchers.NewS3Fetcher(log, s3Provider, ch)
	m[fetching.S3Type] = s3Fetcher

	return m, nil
}
