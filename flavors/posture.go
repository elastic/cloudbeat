// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/version"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	aws_dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/aws"
	k8s_dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/pipeline"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/transformer"
	"github.com/elastic/cloudbeat/uniqueness"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// posture configuration.
type posture struct {
	flavorBase
	data       *fetchersManager.Data
	evaluator  evaluator.Evaluator
	resourceCh chan fetching.ResourceInfo
	leader     uniqueness.Manager
	dataStop   fetchersManager.Stop
}

// NewPosture creates an instance of posture.
func NewPosture(_ *beat.Beat, cfg *agentconfig.C) (*posture, error) {
	log := logp.NewLogger("posture")

	ctx, cancel := context.WithCancel(context.Background())

	c, err := config.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	log.Info("Config initiated with cycle period of ", c.Period)

	resourceCh := make(chan fetching.ResourceInfo, resourceChBuffer)

	le := uniqueness.NewLeaderElector(log, c, &providers.KubernetesProvider{})

	fetchersRegistry, err := initRegistry(log, c, resourceCh, le)
	if err != nil {
		cancel()
		return nil, err
	}

	// TODO: timeout should be configurable and not hard-coded. Setting to 10 minutes for now to account for CSPM fetchers
	// 	https://github.com/elastic/cloudbeat/issues/653
	data, err := fetchersManager.NewData(log, c.Period, time.Minute*10, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log, c)
	if err != nil {
		cancel()
		return nil, err
	}

	// namespace will be passed as param from fleet on https://github.com/elastic/security-team/issues/2383 and it's user configurable
	resultsIndex := config.Datastream("", config.ResultsDatastreamIndexPrefix)

	cdp, err := GetCommonDataProvider(ctx, log, *c)
	if err != nil {
		cancel()
		return nil, err
	}

	t := transformer.NewTransformer(log, cdp, resultsIndex)

	base := flavorBase{
		ctx:         ctx,
		cancel:      cancel,
		config:      c,
		transformer: t,
		log:         log,
	}

	bt := &posture{
		flavorBase: base,
		evaluator:  eval,
		data:       data,
		resourceCh: resourceCh,
		leader:     le,
	}
	return bt, nil
}

// Run starts posture.
func (bt *posture) Run(b *beat.Beat) error {
	bt.log.Info("posture is running! Hit CTRL-C to stop it")

	if err := bt.leader.Run(bt.ctx); err != nil {
		return err
	}

	bt.dataStop = bt.data.Run(bt.ctx)

	procs, err := bt.configureProcessors(bt.config.Processors)
	if err != nil {
		return err
	}
	bt.log.Debugf("posture configured %d processors", len(bt.config.Processors))

	// Connect publisher (with beat's processors)
	if bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			Processor: procs,
		},
	}); err != nil {
		return err
	}

	// Creating the data pipeline
	findingsCh := pipeline.Step(bt.log, bt.resourceCh, bt.evaluator.Eval)
	eventsCh := pipeline.Step(bt.log, findingsCh, bt.transformer.CreateBeatEvents)

	var eventsToSend []beat.Event
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-bt.ctx.Done():
			bt.log.Warn("Posture context is done")
			return nil

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-ticker.C:
			if len(eventsToSend) == 0 {
				continue
			}

			bt.log.Infof("Publishing %d posture events to elasticsearch, time interval reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil

		// Flush events to ES when reaching a certain threshold
		case events := <-eventsCh:
			eventsToSend = append(eventsToSend, events...)
			if len(eventsToSend) < eventsThreshold {
				continue
			}

			bt.log.Infof("Publishing %d posture events to elasticsearch, buffer threshold reached", len(eventsToSend))
			bt.client.PublishAll(eventsToSend)
			eventsToSend = nil
		}
	}
}

func initRegistry(log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager) (fetchersManager.FetchersRegistry, error) {
	registry := fetchersManager.NewFetcherRegistry(log)

	parsedList, err := fetchersManager.Factories.ParseConfigFetchers(log, cfg, ch)
	if err != nil {
		return nil, err
	}

	if err := registry.RegisterFetchers(parsedList, le); err != nil {
		return nil, err
	}

	return registry, nil
}

// Stop stops posture.
func (bt *posture) Stop() {
	if bt.dataStop != nil {
		bt.dataStop(bt.ctx, shutdownGracePeriod)
	}
	bt.evaluator.Stop(bt.ctx)
	bt.leader.Stop()
	close(bt.resourceCh)
	if err := bt.client.Close(); err != nil {
		bt.log.Fatal("Cannot close client", err)
	}

	bt.cancel()
}

// configureProcessors configure processors to be used by the beat
func (bt *posture) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}

func GetCommonDataProvider(ctx context.Context, log *logp.Logger, cfg config.Config) (dataprovider.CommonDataProvider, error) {
	if cfg.Benchmark == config.CIS_EKS || cfg.Benchmark == config.CIS_K8S {
		kubeClient, err := providers.KubernetesProvider{}.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
		if err != nil {
			return nil, err
		}

		clusterNameProvider := providers.ClusterNameProvider{
			KubernetesClusterNameProvider: providers.KubernetesClusterNameProvider{},
			EKSMetadataProvider:           awslib.Ec2MetadataProvider{},
			EKSClusterNameProvider:        awslib.EKSClusterNameProvider{},
			KubeClient:                    kubeClient,
			AwsConfigProvider: awslib.ConfigProvider{
				MetadataProvider: awslib.Ec2MetadataProvider{},
			},
		}
		name, err := clusterNameProvider.GetClusterName(ctx, &cfg, log)
		if err != nil {
			log.Errorf("failed to get cluster name: %v", err)
		}
		v, err := kubeClient.Discovery().ServerVersion()
		if err != nil {
			return nil, err
		}
		node, err := kubernetes.DiscoverKubernetesNode(log, &kubernetes.DiscoverKubernetesNodeParams{
			ConfigHost:  "",
			Client:      kubeClient,
			IsInCluster: true,
			HostUtils:   &kubernetes.DefaultDiscoveryUtils{},
		})
		if err != nil {
			return nil, err
		}
		n, err := kubeClient.CoreV1().Nodes().Get(ctx, node, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		ns, err := kubeClient.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		options := []k8s_dataprovider.Option{
			k8s_dataprovider.WithConfig(&cfg),
			k8s_dataprovider.WithLogger(log),
			k8s_dataprovider.WithClusterName(name),
			k8s_dataprovider.WithClusterID(string(ns.ObjectMeta.UID)),
			k8s_dataprovider.WithNodeID(string(n.ObjectMeta.UID)),
			k8s_dataprovider.WithVersionInfo(version.CloudbeatVersionInfo{
				Version: version.CloudbeatVersion(),
				Policy:  version.PolicyVersion(),
				Kubernetes: version.Version{
					Version: v.Major + "." + v.Minor,
				},
			}),
		}
		return k8s_dataprovider.New(options...), nil
	}

	if cfg.Benchmark == config.CIS_AWS {
		awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
		}

		identityClient := awslib.GetIdentityClient(awsConfig)
		iamProvider := iam.NewIAMProvider(log, awsConfig)

		identity, err := identityClient.GetIdentity(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get AWS identity: %w", err)
		}

		alias, err := iamProvider.GetAccountAlias(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get AWS account alias: %w", err)
		}
		return aws_dataprovider.New(
			aws_dataprovider.WithLogger(log),
			aws_dataprovider.WithAccount(alias, *identity.Account),
		), nil
	}
	return nil, fmt.Errorf("could not get common data provider for benchmark %s", cfg.Benchmark)
}
