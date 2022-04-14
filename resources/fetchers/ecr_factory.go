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
	manager.Factories.ListFetcherFactory(ECRType, &ECRFactory{extraElements: getEcrExtraElements})
}

type ECRFactory struct {
	extraElements func() (ecrExtraElements, error)
}

type ecrExtraElements struct {
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
	elements, err := f.extraElements()
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg, elements)
}

func getEcrExtraElements() (ecrExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return ecrExtraElements{}, err
	}
	ecr := awslib.NewEcrProvider(awsConfig.Config)
	identityProvider := awslib.NewAWSIdentityProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	extraElements := ecrExtraElements{
		awsConfig:              awsConfig,
		kubernetesClientGetter: kubeGetter,
		identityProviderGetter: identityProvider,
		ecrRepoDescriber:       ecr,
	}

	return extraElements, nil
}

func (f *ECRFactory) CreateFrom(cfg ECRFetcherConfig, elements ecrExtraElements) (fetching.Fetcher, error) {
	ctx := context.Background()
	identity, err := elements.identityProviderGetter.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, elements.awsConfig.Config.Region)
	kubeClient, err := elements.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	fe := &ECRFetcher{
		cfg:         cfg,
		ecrProvider: elements.ecrRepoDescriber,
		kubeClient:  kubeClient,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}
	return fe, nil
}
