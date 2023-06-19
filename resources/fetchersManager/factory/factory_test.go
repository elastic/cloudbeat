package factory

import (
	"context"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

type expectedFetchers struct {
	names []string
	count int
}

func TestNewFactory(t *testing.T) {
	logger := logp.NewLogger("test new factory")
	ch := make(chan fetching.ResourceInfo)
	le := &uniqueness.DefaultUniqueManager{}
	kubeClient := k8sfake.NewSimpleClientset()

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		want    expectedFetchers
	}{
		{
			name: "Get k8s factory",
			cfg: &config.Config{
				Benchmark: config.CIS_K8S,
			},
			want: expectedFetchers{
				names: []string{
					fetching.FileSystemType,
					fetching.KubeAPIType,
					fetching.ProcessType,
				},
				count: 3,
			},
		},
		{
			name: "Get CIS AWS factory",
			cfg: &config.Config{
				Benchmark: config.CIS_AWS,
			},
			want: expectedFetchers{
				names: []string{
					fetching.IAMType,
					fetching.KmsType,
					fetching.TrailType,
					fetching.MonitoringType,
					fetching.EC2NetworkingType,
					fetching.RdsType,
					fetching.S3Type,
				},
				count: 7,
			},
		},
		{
			name: "Get CIS EKS factory",
			cfg: &config.Config{
				Benchmark: config.CIS_EKS,
			},
			want: expectedFetchers{
				names: []string{
					fetching.FileSystemType,
					fetching.KubeAPIType,
					fetching.ProcessType,
					fetching.EcrType,
					fetching.ElbType,
				},
				count: 5,
			},
		},
		{
			name: "Non supported benchmark fail to get factory",
			cfg: &config.Config{
				Benchmark: "Non existing benchmark",
			},
			want: expectedFetchers{
				names: []string{},
				count: 0,
			},
			wantErr: true,
		},
		{
			name: "No config fail to get factory",
			cfg:  nil,
			want: expectedFetchers{
				names: []string{},
				count: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchersMap, err := NewFactory(context.TODO(), logger, tt.cfg, ch, le, kubeClient)
			assert.Equal(t, len(fetchersMap), tt.want.count)
			for fetcher := range fetchersMap {
				if _, ok := fetchersMap[fetcher]; !ok {
					t.Errorf("NewFactory() fetchersMap = %v, want %v", fetchersMap, tt.want.names)
				}
			}

			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
