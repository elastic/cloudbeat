package fetchers

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/aws_providers"
	"github.com/elastic/cloudbeat/resources/providers"
	"regexp"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	ELBType = "aws-elb"
)

func init() {
	awsCredProvider := aws_providers.AWSCredProvider{}
	awsCred := awsCredProvider.GetAwsCredentials()
	elb := aws_providers.NewELBProvider(awsCred.Config)
	kubeGetter := providers.KubernetesProvider{}

	manager.Factories.ListFetcherFactory(ELBType,
		&ELBFactory{
			balancerDescriber:      elb,
			awsCredProvider:        awsCredProvider,
			kubernetesClientGetter: kubeGetter,
		},
	)
}

type ELBFactory struct {
	balancerDescriber      aws_providers.ELBLoadBalancerDescriber
	awsCredProvider        aws_providers.AwsCredentialsGetter
	kubernetesClientGetter providers.KubernetesClientGetter
}

func (f *ELBFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := ELBFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *ELBFactory) CreateFrom(cfg ELBFetcherConfig) (fetching.Fetcher, error) {
	awsCfg := f.awsCredProvider.GetAwsCredentials()
	elb := f.balancerDescriber
	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, awsCfg.Config.Region)
	kubeClient, err := f.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes: %w", err)
	}

	return &ELBFetcher{
		elbProvider:     elb,
		cfg:             cfg,
		kubeClient:      kubeClient,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(loadBalancerRegex)},
	}, nil
}
