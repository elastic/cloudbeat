package fetchers

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
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

	manager.Factories.ListFetcherFactory(ELBType,
		&ELBFactory{
			extraElements: getElbExtraElements,
		},
	)
}

type ELBFactory struct {
	extraElements func() (elbExtraElements, error)
}

type elbExtraElements struct {
	balancerDescriber      awslib.ELBLoadBalancerDescriber
	awsConfig              awslib.Config
	kubernetesClientGetter providers.KubernetesClientGetter
}

func (f *ELBFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	logp.L().Info("ELB factory has started")
	cfg := ELBFetcherConfig{}
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

func getElbExtraElements() (elbExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return elbExtraElements{}, err
	}
	elb := awslib.NewELBProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	return elbExtraElements{
		balancerDescriber:      elb,
		awsConfig:              awsConfig,
		kubernetesClientGetter: kubeGetter,
	}, err
}

func (f *ELBFactory) CreateFrom(cfg ELBFetcherConfig, elements elbExtraElements) (fetching.Fetcher, error) {
	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, elements.awsConfig.Config.Region)
	kubeClient, err := elements.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes: %w", err)
	}

	return &ELBFetcher{
		elbProvider:     elements.balancerDescriber,
		cfg:             cfg,
		kubeClient:      kubeClient,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(loadBalancerRegex)},
	}, nil
}
