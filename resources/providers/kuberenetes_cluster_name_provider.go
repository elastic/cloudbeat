package providers

import (
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes/metadata"
	agentcfg "github.com/elastic/elastic-agent-libs/config"
	k8s "k8s.io/client-go/kubernetes"
)

type KubernetesClusterNameProviderApi interface {
	GetClusterName(cfg *config.Config, client k8s.Interface) (string, error)
}
type KubernetesClusterNameProvider struct {
}

func (provider KubernetesClusterNameProvider) GetClusterName(cfg *config.Config, client k8s.Interface) (string, error) {
	agentConfig, err := agentcfg.NewConfigFrom(cfg)
	clusterIdentifier, err := metadata.GetKubernetesClusterIdentifier(agentConfig, client)
	if err != nil {
		return "", fmt.Errorf("fail to resolve the name of the cluster, error %v", err)
	}

	return clusterIdentifier.Name, nil
}
