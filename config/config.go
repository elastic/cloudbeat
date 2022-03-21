// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/processors"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cis_kubernetes_benchmark.findings"
const MetadataDatastreamIndexPrefix = ".logs-cis_kubernetes_benchmark.metadata"

type Config struct {
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Processors processors.PluginConfig `config:"processors"`
	Fetchers   []*common.Config        `config:"fetchers"`
}

var DefaultConfig = Config{
	Period: 10 * time.Second,
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}
