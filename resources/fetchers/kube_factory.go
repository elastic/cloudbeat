package fetchers

import (
	"encoding/gob"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

type KubeFactory struct {
}

func init() {
	manager.Factories.ListFetcherFactory(fetching.KubeAPIType, &KubeFactory{})
	gob.Register(K8sResource{})
	gob.Register(ECRResource{})
	gob.Register(ELBResource{})
	gob.Register(EKSResource{})
	gob.Register(IAMResource{})
	gob.Register(kubernetes.Pod{})
	gob.Register(kubernetes.Secret{})
	gob.Register(kubernetes.Role{})
	gob.Register(kubernetes.RoleBinding{})
	gob.Register(kubernetes.ClusterRole{})
	gob.Register(kubernetes.ClusterRoleBinding{})
	gob.Register(kubernetes.NetworkPolicy{})
	gob.Register(kubernetes.PodSecurityPolicy{})
}

func (f *KubeFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := KubeApiFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *KubeFactory) CreateFrom(cfg KubeApiFetcherConfig) (fetching.Fetcher, error) {
	fe := &KubeFetcher{
		cfg:      cfg,
		watchers: make([]kubernetes.Watcher, 0),
	}

	return fe, nil
}
