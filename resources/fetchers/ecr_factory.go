package fetchers

import (
	"context"
	"encoding/gob"
	"fmt"
	"regexp"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	ECRType = "aws-ecr"
)

func init() {
	manager.Factories.ListFetcherFactory(ECRType, &ECRFactory{})
	gob.Register(ECRResource{})
}

type ECRFactory struct {
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
	fe := &ECRFetcher{
		cfg: cfg,
	}

	return fe, nil
}

func NewECRFetcher(awsCfg AwsFetcherConfig, cfg ECRFetcherConfig, ctx context.Context) (fetching.Fetcher, error) {
	ecrProvider := NewEcrProvider(awsCfg.Config)
	identityProvider := NewAWSIdentityProvider(awsCfg.Config)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, awsCfg.Config.Region)
	kubeClient, err := kubernetes.GetKubernetesClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})

	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}
	return &ECRFetcher{
		cfg:         cfg,
		ecrProvider: ecrProvider,
		kubeClient:  kubeClient,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}, nil
}
