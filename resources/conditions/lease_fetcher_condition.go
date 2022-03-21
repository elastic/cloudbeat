package conditions

import (
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type LeaderLeaseProvider interface {
	IsLeader() (bool, error)
}

type LeaseFetcherCondition struct {
	provider LeaderLeaseProvider
}

func NewLeaseFetcherCondition(provider LeaderLeaseProvider) fetching.Condition {
	return &LeaseFetcherCondition{
		provider: provider,
	}
}

func (c *LeaseFetcherCondition) Condition() bool {
	l, err := c.provider.IsLeader()
	if err != nil {
		logp.L().Errorf("could not read leader value, using default value %v: %v", l, err)
	}
	return l
}

func (c *LeaseFetcherCondition) Name() string {
	return "leader_election_conditional_fetcher"
}
