package fetchers

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"regexp"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	ELBType = "aws-elb"
)

func init() {
	awsConfigProvider := awslib.ConfigProvider{}
	config := awsConfigProvider.GetConfig()
	elb := awslib.NewELBProvider(config.Config)
	kubeGetter := providers.KubernetesProvider{}

	manager.Factories.ListFetcherFactory(ELBType,
		&ELBFactory{
			awsConfig:              config,
			balancerDescriber:      elb,
			kubernetesClientGetter: kubeGetter,
		},
	)
}

type ELBFactory struct {
	balancerDescriber      awslib.ELBLoadBalancerDescriber
	awsConfig              awslib.Config
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
	elb := f.balancerDescriber
	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, f.awsConfig.Config.Region)
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
