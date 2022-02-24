// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"github.com/elastic/beats/v7/libbeat/processors"
	"time"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cis_kubernetes_benchmark.findings"
const MetadataDatastreamIndexPrefix = ".logs-cis_kubernetes_benchmark.metadata"

type Config struct {
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Files      []string                `config:"files"`
	Processors processors.PluginConfig `config:"processors"`
}

var DefaultConfig = Config{
	Period: 10 * time.Second,
	Files: []string{
		"/hostfs/etc/kubernetes/scheduler.conf",
		"/hostfs/etc/kubernetes/controller-manager.conf",
		"/hostfs/etc/kubernetes/admin.conf",
		"/hostfs/etc/kubernetes/kubelet.conf",
		"/hostfs/etc/kubernetes/manifests/etcd.yaml",
		"/hostfs/etc/kubernetes/manifests/kube-apiserver.yaml",
		"/hostfs/etc/kubernetes/manifests/kube-controller-manager.yaml",
		"/hostfs/etc/kubernetes/manifests/kube-scheduler.yaml",
		"/hostfs/etc/systemd/system/kubelet.service.d/10-kubeadm.conf",
		"/hostfs/var/lib/kubelet/config.yaml",
		"/hostfs/var/lib/etcd/**",
		"/hostfs/etc/kubernetes/pki/**",
	},
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}
