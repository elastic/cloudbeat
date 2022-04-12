package fetchers

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"regexp"
)

const (
	ECRType = "aws-ecr"
)

func init() {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig := awsConfigProvider.GetConfig()
	ecr := awslib.NewEcrProvider(awsConfig.Config)
	identityProvider := awslib.NewAWSIdentityProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	manager.Factories.ListFetcherFactory(ECRType, &ECRFactory{
		kubernetesClientGetter: kubeGetter,
		identityProviderGetter: identityProvider,
		ecrRepoDescriber:       ecr,
		awsConfig:              awsConfig,
	})
}

type ECRFactory struct {
	awsConfig              awslib.Config
	kubernetesClientGetter providers.KubernetesClientGetter
	identityProviderGetter awslib.IdentityProviderGetter
	ecrRepoDescriber       awslib.EcrRepositoryDescriber
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
