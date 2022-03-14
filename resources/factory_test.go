package resources

import (
	"context"
	"testing"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/stretchr/testify/suite"
)

type FactoriesTestSuite struct {
	suite.Suite
	F factories
}

type numberFetcherFactory struct {
}

func (n *numberFetcherFactory) Create(c *common.Config) (fetchers.Fetcher, error) {
	x, _ := c.Int("num", 1)
	return &numberFetcher{int(x), false}, nil
}

func numberConfig(number int) *common.Config {
	c := common.NewConfig()
	c.SetInt("num", 1, int64(number))
	return c
}

func TestFactoriesTestSuite(t *testing.T) {
	suite.Run(t, new(FactoriesTestSuite))
}

func (s *FactoriesTestSuite) SetupTest() {
	s.F = newFactories()
}

func (s *FactoriesTestSuite) TestListFetcher() {
	var tests = []struct {
		key string
	}{
		{"duplicate_fetcher"},
		{"duplicate_fetcher"},
		{"other_fetcher"},
		{"new_fetcher"},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
	}

	s.Contains(s.F.m, "duplicate_fetcher")
	s.Contains(s.F.m, "other_fetcher")
	s.Contains(s.F.m, "new_fetcher")
}

func (s *FactoriesTestSuite) TestCreateFetcher() {
	var tests = []struct {
		key   string
		value int
	}{
		{"duplicate_fetcher", 1},
		{"duplicate_fetcher", 2},
		{"other_fetcher", 4},
		{"new_fetcher", 6},
	}

	for _, test := range tests {
		s.F.ListFetcherFactory(test.key, &numberFetcherFactory{})
		c := numberConfig(test.value)
		f, err := s.F.CreateFetcher(test.key, c)
		s.NoError(err)
		res, err := f.Fetch(context.TODO())
		s.NoError(err)

		s.Equal(1, len(res))
		s.Equal(test.value, res[0].GetData())
	}
}
