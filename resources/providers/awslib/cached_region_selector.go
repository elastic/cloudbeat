package awslib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dgraph-io/ristretto"
	"github.com/elastic/elastic-agent-libs/logp"
)

var ristrettoCache *ristretto.Cache

func init() {
	var err error
	ristrettoCache, err = newCachedRegions()
	if err != nil {
		panic(fmt.Errorf("Unable to init region-selector cache: %w", err))
	}
}

func newCachedRegions() (*ristretto.Cache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: 100,
		MaxCost:     10000,
		BufferItems: 64,
	})
}

func CurrentRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&currentRegionSelector{}, "CurrentRegionSelectorCache", 0)
}

func AllRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&allRegionSelector{}, "AllRegionSelectorCache", 720*time.Hour)
}

type cachedRegions struct {
	regions []string
}

type cachedRegionSelector struct {
	lock   *sync.Mutex
	cache  *ristretto.Cache
	keep   time.Duration
	key    string
	client RegionsSelector
}

func newCachedRegionSelector(selector RegionsSelector, cache string, keep time.Duration) *cachedRegionSelector {
	return &cachedRegionSelector{
		lock:   &sync.Mutex{},
		cache:  ristrettoCache,
		keep:   keep,
		key:    cache,
		client: selector,
	}
}

func (s *cachedRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := logp.NewLogger("aws")

	cachedObject := s.getCache()
	if cachedObject != nil {
		return cachedObject, nil
	}

	// Make sure that consequent calls to the function will keep trying to retrieve the regions list until it succeeds.
	s.lock.Lock()
	defer s.lock.Unlock()
	cachedObject = s.getCache()
	if cachedObject != nil {
		return cachedObject, nil
	}

	log.Debug("RegionsSelector starting to retrieve regions")
	var output []string
	output, err := s.client.Regions(ctx, cfg)
	if err != nil {
		log.Errorf("Failed getting regions: %v", err)
		return nil, err
	}

	if !s.setCache(output) {
		log.Errorf("Failed setting regions cache")
	}
	return output, nil
}

func (s *cachedRegionSelector) setCache(list []string) bool {
	cache := &cachedRegions{
		regions: list,
	}

	return ristrettoCache.SetWithTTL(s.key, cache, 1, s.keep)
}
func (s *cachedRegionSelector) getCache() []string {
	cachedObject, ok := ristrettoCache.Get(s.key)
	if !ok {
		return nil
	}

	cachedList, ok := cachedObject.(*cachedRegions)
	if !ok {
		return nil
	}

	return cachedList.regions
}
