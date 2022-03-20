package fetchers

import (
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type KubeFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestKubeFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(KubeFactoryTestSuite))
}

func (s *KubeFactoryTestSuite) SetupTest() {
	s.factory = &KubeFactory{}
}

func (s *KubeFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config           string
		expectedInterval time.Duration
	}{
		{
			`
name: kube-api
interval: 500
`,
			time.Second * 500,
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := s.factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		kube, ok := fetcher.(*KubeFetcher)
		s.True(ok)
		s.Equal(test.expectedInterval, kube.cfg.Interval)
	}
}
