package providers

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	k8s "k8s.io/client-go/kubernetes"
)

type ClusterNameProviderAPI interface {
	GetClusterName(ctx context.Context, cfg config.Config) (string, error)
}

type ClusterNameProvider struct {
	KubernetesClusterNameProvider KubernetesClusterNameProviderApi
	EKSMetadataProvider           awslib.MetadataProvider
	EKSClusterNameProvider        awslib.ClusterNameProvider
	KubeClient                    k8s.Interface
	AwsConfigProvider             awslib.ConfigProviderAPI
}

func (provider ClusterNameProvider) GetClusterName(ctx context.Context, cfg config.Config) (string, error) {
	switch cfg.Type {
	case config.InputTypeVanillaK8s:
		return provider.KubernetesClusterNameProvider.GetClusterName(cfg, provider.KubeClient)
	case config.InputTypeEks:
		awsConfig, err := provider.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.AWSConfig)
		if err != nil {
			return "", fmt.Errorf("failed to initialize aws configuration for identifying the cluster name: %v", err)
		}
		metadata, err := provider.EKSMetadataProvider.GetMetadata(ctx, awsConfig)
		if err != nil {
			return "", fmt.Errorf("failed to get the ec2 metadata required for identifying the cluster name: %v", err)
		}
		instanceId := metadata.InstanceID
		return provider.EKSClusterNameProvider.GetClusterName(ctx, awsConfig, instanceId)
	default:
		panic(fmt.Sprintf("unknown cluster type: %s", cfg.Type))
	}
}
