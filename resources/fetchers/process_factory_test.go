package fetchers

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type ProcessFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestProcessFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessFactoryTestSuite))
}

func (s *ProcessFactoryTestSuite) SetupTest() {
	s.factory = &ProcessFactory{}
}

func (s *ProcessFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config            string
		expectedDirectory string
	}{
		{
			`
name: process
directory: /hostfs
`,
			"/hostfs",
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := s.factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		process, ok := fetcher.(*ProcessesFetcher)
		s.True(ok)
		s.Equal(test.expectedDirectory, process.cfg.Directory)
	}
}
