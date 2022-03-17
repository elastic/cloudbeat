package fetchers

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type FileFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestFileFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FileFactoryTestSuite))
}

func (s *FileFactoryTestSuite) SetupTest() {
	s.factory = &FileSystemFactory{}
}

func (s *FileFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config           string
		expectedPatterns []string
	}{
		{
			`
name: file-system
patterns: [
  "hello",
  "world"
]
`,
			[]string{"hello", "world"},
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := s.factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		file, ok := fetcher.(*FileSystemFetcher)
		s.True(ok)
		s.Equal(test.expectedPatterns, file.cfg.Patterns)
	}
}
