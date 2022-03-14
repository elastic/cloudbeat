package resources

import (
	"context"
	"errors"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/conditions"
	"github.com/elastic/cloudbeat/resources/fetchers"
)

func init() {
	Factories.ListFetcherFactory(fetchers.KubeAPIType, &fetchers.KubeFactory{})
	Factories.ListFetcherFactory(fetchers.ProcessType, &fetchers.ProcessFactory{})
	Factories.ListFetcherFactory(fetchers.FileSystemType, &fetchers.FileSystemFactory{})
}

var Factories = newFactories()

type FetcherFactory interface {
	Create(*common.Config) (fetchers.Fetcher, error)
}

type factories struct {
	m map[string]FetcherFactory
}

func newFactories() factories {
	return factories{m: make(map[string]FetcherFactory)}
}

func (fa *factories) ListFetcherFactory(name string, f FetcherFactory) {
	_, ok := fa.m[name]
	if ok {
		logp.L().Warnf("fetcher %q factory method overwritten", name)
	}

	fa.m[name] = f
}

func (fa *factories) CreateFetcher(name string, c *common.Config) (fetchers.Fetcher, error) {
	factory, ok := fa.m[name]
	if !ok {
		return nil, errors.New("fetcher factory could not be found")
	}

	return factory.Create(c)
}

func (fa *factories) ConfigFetchers(registry FetchersRegistry, cfg config.Config) error {
	parsedList := fa.ParseConfigFetchers(cfg)
	for _, p := range parsedList {
		if p.err != nil {
			return p.err
		}

		c := fa.getConditions(p.name)
		registry.Register(p.name, p.f, c...)
	}

	return nil
}

func (fa *factories) getConditions(name string) []fetchers.FetcherCondition {
	c := make([]fetchers.FetcherCondition, 0)
	switch name {
	case fetchers.KubeAPIType:
		client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
		if err != nil {
			leaseProvider := conditions.NewLeaderLeaseProvider(context.Background(), client)
			condition := conditions.NewLeaseFetcherCondition(leaseProvider)
			c = append(c, condition)
		}
	}

	return c
}

type ParsedFetcher struct {
	name string
	f    fetchers.Fetcher
	err  error
}

func (fa *factories) ParseConfigFetchers(cfg config.Config) []*ParsedFetcher {
	arr := []*ParsedFetcher{}
	for _, fcfg := range cfg.Fetchers {
		p := fa.ParseConfigFetcher(fcfg)
		arr = append(arr, p)
	}

	return arr
}

func (fa *factories) ParseConfigFetcher(fcfg *common.Config) *ParsedFetcher {
	gen := fetchers.BaseFetcherConfig{}
	err := fcfg.Unpack(&gen)
	if err != nil {
		return &ParsedFetcher{gen.Name, nil, err}
	}

	f, err := fa.CreateFetcher(gen.Name, fcfg)
	if err != nil {
		return &ParsedFetcher{gen.Name, f, err}
	}

	return &ParsedFetcher{gen.Name, f, err}
}
