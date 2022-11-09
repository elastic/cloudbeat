package awslib

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type ClusterNameProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestClusterNameProviderTestSuite(t *testing.T) {
	s := new(ClusterNameProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_cluster_name_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ClusterNameProviderTestSuite) TestA() {

	instanceId := "i-0d004c69210358d6e"
	ctx := context.Background()
	provider := EKSClusterNameProvider{}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-1"))

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	clusterName, err := provider.GetClusterName(ctx, cfg, instanceId)
	s.NoError(err)

	s.NotEmpty(clusterName)
}
