package benchmark

import (
	"context"
	"fmt"

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_gokubernetes "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/k8s"
)

type k8sBenchmarkHelper struct {
	log    *logp.Logger
	cfg    *config.Config
	client client_gokubernetes.Interface
}

func NewK8sBenchmarkHelper(log *logp.Logger, cfg *config.Config, client client_gokubernetes.Interface) *k8sBenchmarkHelper {
	return &k8sBenchmarkHelper{
		log:    log,
		cfg:    cfg,
		client: client,
	}
}

func (h *k8sBenchmarkHelper) GetK8sDataProvider(ctx context.Context, clusterNameProvider k8s.ClusterNameProviderAPI) (dataprovider.CommonDataProvider, error) {
	clusterName, err := clusterNameProvider.GetClusterName(ctx, h.cfg)
	if err != nil {
		h.log.Errorf("failed to get cluster name: %v", err)
	}

	serverVersion, err := h.client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	clusterId, err := h.getK8sClusterId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster id: %w", err)
	}

	options := []k8s.Option{
		k8s.WithLogger(h.log),
		k8s.WithClusterName(clusterName),
		k8s.WithClusterID(clusterId),
		k8s.WithClusterVersion(serverVersion.String()),
	}
	return k8s.New(options...), nil
}

func (h *k8sBenchmarkHelper) GetK8sIdProvider(ctx context.Context) (dataprovider.IdProvider, error) {
	nodeId, err := h.getK8sNodeId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node id: %w", err)
	}

	clusterId, err := h.getK8sClusterId(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster id: %w", err)
	}

	return k8s.NewIdProvider(clusterId, nodeId), nil
}

func (h *k8sBenchmarkHelper) getK8sClusterId(ctx context.Context) (string, error) {
	namespace, err := h.client.CoreV1().Namespaces().Get(ctx, "kube-system", v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get namespace data: %w", err)
	}

	return string(namespace.ObjectMeta.UID), nil
}

func (h *k8sBenchmarkHelper) getK8sNodeId(ctx context.Context) (string, error) {
	nodeName, err := kubernetes.DiscoverKubernetesNode(h.log, &kubernetes.DiscoverKubernetesNodeParams{
		ConfigHost:  "",
		Client:      h.client,
		IsInCluster: true,
		HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get node name: %w", err)
	}

	node, err := h.client.CoreV1().Nodes().Get(ctx, nodeName, v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get node data for node '%s': %w", nodeName, err)
	}

	return string(node.ObjectMeta.UID), nil
}
