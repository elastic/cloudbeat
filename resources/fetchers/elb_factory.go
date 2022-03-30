package fetchers

import (
	"encoding/gob"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/ctxProvider"
	"regexp"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	ELBType = "aws-elb"
)

func init() {
	manager.Factories.ListFetcherFactory(ELBType, &ELBFactory{})
	gob.Register(ELBResource{})
}

type ELBFactory struct {
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
	awsCredProvider := ctxProvider.AWSCredProvider{}
	awsCfg := awsCredProvider.GetAwsCredentials()
	elb := NewELBProvider(awsCfg.Config)
	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, awsCfg.Config.Region)
	kubeClient, err := kubernetes.GetKubernetesClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
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
