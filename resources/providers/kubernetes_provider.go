package providers

import (
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	k8s "k8s.io/client-go/kubernetes"
)

type KubernetesClientGetter interface {
	GetClient(kubeConfig string, options kubernetes.KubeClientOptions) (k8s.Interface, error)
}

type KubernetesProvider struct {
}

func (provider KubernetesProvider) GetClient(kubeConfig string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
	client, err := kubernetes.GetKubernetesClient(kubeConfig, options)
	return client, err
}
