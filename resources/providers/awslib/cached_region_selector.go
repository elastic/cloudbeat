package awslib

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"
)

var currentSelectorCache = &cachedRegions{}

var allSelectorCache = &cachedRegions{}

func CurrentRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&currentRegionSelector{}, currentSelectorCache)
}

func AllRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&allRegionsSelector{}, allSelectorCache)
}

type cachedRegions struct {
	regions []string
}

type cachedRegionSelector struct {
	once   *sync.Once
	lock   *sync.Mutex
	cache  *cachedRegions
	client RegionsSelector
}

func newCachedRegionSelector(selector RegionsSelector, cache *cachedRegions) *cachedRegionSelector {
	return &cachedRegionSelector{
		once:   &sync.Once{},
		lock:   &sync.Mutex{},
		cache:  cache,
		client: selector,
	}
}

func (s *cachedRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := logp.NewLogger("aws")
	log.Debug("allRegionsSelector starting...")
	var err error

	// Make sure that consequent calls to the function will keep trying to retrieve the regions list until it succeeds.
	s.lock.Lock()
	defer s.lock.Unlock()
	s.once.Do(func() {
		log.Debug("Get aws regions for the first time")
		var output []string
		output, err = s.client.Regions(ctx, cfg)
		if err != nil {
			log.Errorf("failed DescribeRegions: %v", err)
			s.once = &sync.Once{} // reset singleton upon error
			return
		}

		s.cache = &cachedRegions{}
		for _, region := range output {
			s.cache.regions = append(s.cache.regions, region)
		}
	})

	log.Debugf("enabled regions for aws account, %v", s.cache.regions)
	return s.cache.regions, err
}
