package fetchers

import (
	"encoding/gob"
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/ctxProvider"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"regexp"
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
	awsCredProvider := ctxProvider.AWSCredProvider{}
	awsCfg := awsCredProvider.GetAwsCredentials()
	ctx := context.Background()
	identityProvider := ctxProvider.NewAWSIdentityProvider(awsCfg.Config)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, awsCfg.Config.Region)
	ecrProvider := NewEcrProvider(awsCfg.Config)
	kubeClient, err := kubernetes.GetKubernetesClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	fe := &ECRFetcher{
		cfg:         cfg,
		ecrProvider: ecrProvider,
		kubeClient:  kubeClient,
		repoRegexMatchers: []*regexp.Regexp{
			regexp.MustCompile(privateRepoRegex),
			regexp.MustCompile(PublicRepoRegex),
		},
	}
	return fe, nil
}

//
//func (f *ECRFactory) CreateFrom(cfg ECRFetcherConfig) (fetching.Fetcher, error) {
//	awsCredProvider := ctxProvider.AWSCredProvider{}
//	awsCfg := awsCredProvider.GetAwsCredentials()
//	ctx := context.Background()
//	ecrProvider := NewEcrProvider(awsCfg.Config)
//	identityProvider := NewAWSIdentityProvider(awsCfg.Config)
//	identity, err := identityProvider.GetIdentity(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
//	}
//

//
//	fe := &ECRFetcher{
//		cfg:         cfg,
//		ecrProvider: ecrProvider,
//		kubeClient:  kubeClient,
//		repoRegexMatchers: []*regexp.Regexp{
//			regexp.MustCompile(privateRepoRegex),
//			regexp.MustCompile(PublicRepoRegex),
//		},
//	}
//
//	return fe, nil
//}
