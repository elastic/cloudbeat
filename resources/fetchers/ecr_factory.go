package fetchers

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/aws"
	"regexp"
)

const (
	ECRType = "aws-ecr"
)

func init() {
	awsConfigProvider := aws.ConfigProvider{}
	awsConfig := awsConfigProvider.GetConfig()
	ecr := aws.NewEcrProvider(awsConfig.Config)
	identityProvider := aws.NewAWSIdentityProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	manager.Factories.ListFetcherFactory(ECRType, &ECRFactory{
		kubernetesClientGetter: kubeGetter,
		identityProviderGetter: identityProvider,
		ecrRepoDescriber:       ecr,
		awsConfig:              awsConfig,
	})
}

type ECRFactory struct {
	awsConfig              aws.Config
	kubernetesClientGetter providers.KubernetesClientGetter
	identityProviderGetter aws.IdentityProviderGetter
	ecrRepoDescriber       aws.EcrRepositoryDescriber
}

func (f *ECRFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := ECRFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *ECRFactory) CreateFrom(cfg ECRFetcherConfig) (fetching.Fetcher, error) {
	ctx := context.Background()
	identity, err := f.identityProviderGetter.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, f.awsConfig.Config.Region)
	kubeClient, err := f.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	fe := &ECRFetcher{
		cfg:         cfg,
		ecrProvider: f.ecrRepoDescriber,
		kubeClient:  kubeClient,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}
	return fe, nil
}
