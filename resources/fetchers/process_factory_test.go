package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/common"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type ProcessFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}
type ProcessConfigTestValidator struct {
	processName string
	validate    func([]string)
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
		processValidators []ProcessConfigTestValidator
	}{
		{
			`
name: process
directory: /hostfs
processes:
 etcd:
 kubelet:
  cmd-arguments:
  - config
`,
			"/hostfs",
			[]ProcessConfigTestValidator{
				{
					processName: "kubelet",
					validate: func(cmd []string) {
						s.Len(cmd, 1)
						s.Contains(cmd, "config")
					},
				},
				{
					processName: "etcd",
					validate: func(cmd []string) {
						s.Nil(cmd)
					},
				}},
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
		s.NotNil(process.cfg.Fs)

		s.Equal(len(test.processValidators), len(process.cfg.RequiredProcesses))
		for _, validator := range test.processValidators {
			validator.validate(process.cfg.RequiredProcesses[validator.processName].CommandArguments)
		}
	}
}
