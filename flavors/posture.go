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
	"io/fs"
	"os"
	"time"

	ecr_fetcher "github.com/elastic/cloudbeat/resources/fetchers/ecr"
	eks_fetcher "github.com/elastic/cloudbeat/resources/fetchers/eks"
	elb_fetcher "github.com/elastic/cloudbeat/resources/fetchers/elb"
	filesystem_fetcher "github.com/elastic/cloudbeat/resources/fetchers/file_system"
	iam_fetcher "github.com/elastic/cloudbeat/resources/fetchers/iam"
	kube_fetcher "github.com/elastic/cloudbeat/resources/fetchers/kube"
	logging_fetcher "github.com/elastic/cloudbeat/resources/fetchers/logging"
	monitoring_fetcher "github.com/elastic/cloudbeat/resources/fetchers/monitoring"
	network_fetcher "github.com/elastic/cloudbeat/resources/fetchers/network"
	process_fetcher "github.com/elastic/cloudbeat/resources/fetchers/process"
	rds_fetcher "github.com/elastic/cloudbeat/resources/fetchers/rds"
	s3_fetcher "github.com/elastic/cloudbeat/resources/fetchers/s3"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/logging"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/configservice"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/resources/utils/user"
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

	aws_sdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudwatch/logs"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/cloudbeat/resources/providers/awslib/sns"

	cloudtrail_sdk "github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudwatch_sdk "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchlogs_sdk "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	configservice_sdk "github.com/aws/aws-sdk-go-v2/service/configservice"
	ec2_sdk "github.com/aws/aws-sdk-go-v2/service/ec2"
	ecr_sdk "github.com/aws/aws-sdk-go-v2/service/ecr"
	eks_sdk "github.com/aws/aws-sdk-go-v2/service/eks"
	elb_sdk "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	iam_sdk "github.com/aws/aws-sdk-go-v2/service/iam"
	rds_sdk "github.com/aws/aws-sdk-go-v2/service/rds"
	s3_sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	securityhub_sdk "github.com/aws/aws-sdk-go-v2/service/securityhub"
	sns_sdk "github.com/aws/aws-sdk-go-v2/service/sns"
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

	fetchersRegistry, err := initRegistry(ctx, log, c, resourceCh, le)
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

func initRegistry(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager) (fetchersManager.FetchersRegistry, error) {
	registry := fetchersManager.NewFetcherRegistry(log)

	r, err := initFetchers(ctx, log, cfg, ch)
	if err != nil {
		return nil, err
	}
	parsedList, err := fetchersManager.ParseConfigFetchers(log, cfg, ch, r)
	if err != nil {
		return nil, err
	}

	if err := registry.RegisterFetchers(parsedList, le); err != nil {
		return nil, err
	}
	return registry, nil
}

func initFetchers(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (map[string]fetching.Fetcher, error) {
	reg := map[string]fetching.Fetcher{}
	list, err := config.GetFetcherNames(cfg)
	if err != nil {
		return nil, err
	}

	k8sProvider := providers.KubernetesProvider{}
	k8sClient, err := k8sProvider.GetClient(log, "", kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, err
	}

	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, err
	}

	identityProvider := awslib.GetIdentityClient(awsConfig)
	identity, err := identityProvider.GetIdentity(context.Background())
	if err != nil {
		return nil, err
	}

	awsIAMService := iam_sdk.NewFromConfig(awsConfig)
	awsELBService := elb_sdk.NewFromConfig(awsConfig)
	awsEKSService := eks_sdk.NewFromConfig(awsConfig)
	awsTrailCrossRegionFactory := &awslib.MultiRegionClientFactory[cloudtrail.Client]{}
	awsCloudwatchCrossRegionFactory := &awslib.MultiRegionClientFactory[cloudwatch.Client]{}
	awsCloudwatchlogsCrossRegionFactory := &awslib.MultiRegionClientFactory[logs.Client]{}
	awsSNSCrossRegionFactory := &awslib.MultiRegionClientFactory[sns.Client]{}
	awsSecurityhubRegionFactory := &awslib.MultiRegionClientFactory[securityhub.Client]{}
	awsS3CrossRegionFactory := &awslib.MultiRegionClientFactory[s3.Client]{}
	awsConfigCrossRegionFactory := &awslib.MultiRegionClientFactory[configservice.Client]{}
	awsRDSCrossRegionFactory := &awslib.MultiRegionClientFactory[rds.Client]{}
	awsEC2CrossRegionFactory := &awslib.MultiRegionClientFactory[ec2.Client]{}
	awsECRCrossRegionFactory := &awslib.MultiRegionClientFactory[*ecr_sdk.Client]{}

	if _, ok := list[iam_fetcher.Type]; ok {
		reg[iam_fetcher.Type] = iam_fetcher.New(
			iam_fetcher.WithCloudIdentity(identity),
			iam_fetcher.WithConfig(cfg),
			iam_fetcher.WithResourceChan(ch),
			iam_fetcher.WithLogger(log),
			iam_fetcher.WithIAMProvider(iam.NewIAMProvider(log, awsIAMService)),
		)
	}

	if _, ok := list[filesystem_fetcher.Type]; ok {
		reg[filesystem_fetcher.Type] = filesystem_fetcher.New(
			filesystem_fetcher.WithConfig(cfg),
			filesystem_fetcher.WithLogger(log),
			filesystem_fetcher.WithResourceChan(ch),
			filesystem_fetcher.WithOSUser(user.NewOSUserUtil()),
		)
	}

	if _, ok := list[kube_fetcher.Type]; ok {
		reg[kube_fetcher.Type] = kube_fetcher.New(
			kube_fetcher.WithLogger(log),
			kube_fetcher.WithConfig(cfg),
			kube_fetcher.WithResourceChan(ch),
			kube_fetcher.WithKubeClientProvider(kubernetes.GetKubernetesClient),
			kube_fetcher.WithWatchers([]kubernetes.Watcher{}),
		)
	}

	if _, ok := list[monitoring_fetcher.Type]; ok {
		reg[monitoring_fetcher.Type] = monitoring_fetcher.New(
			monitoring_fetcher.WithConfig(cfg),
			monitoring_fetcher.WithLogger(log),
			monitoring_fetcher.WithResourceChan(ch),
			monitoring_fetcher.WithCloudIdentity(identity),
			monitoring_fetcher.WithMonitoringProvider(&monitoring.Provider{
				Cloudtrail:     cloudtrail.NewProvider(awsConfig, log, getCloudrailClients(awsTrailCrossRegionFactory, log, awsConfig)),
				Cloudwatch:     cloudwatch.NewProvider(log, awsConfig, getCloudwatchClients(awsCloudwatchCrossRegionFactory, log, awsConfig)),
				Cloudwatchlogs: logs.NewCloudwatchLogsProvider(log, awsConfig, getCloudwatchlogsClients(awsCloudwatchlogsCrossRegionFactory, log, awsConfig)),
				Sns:            sns.NewSNSProvider(log, awsConfig, getSNSClients(awsSNSCrossRegionFactory, log, awsConfig)),
				Log:            log,
			}),
			monitoring_fetcher.WithSecurityhubService(securityhub.NewProvider(awsConfig, log, getSecurityhubClients(awsSecurityhubRegionFactory, log, awsConfig), *identity.Account)),
		)
	}

	if _, ok := list[logging_fetcher.Type]; ok {
		reg[logging_fetcher.Type] = logging_fetcher.New(
			logging_fetcher.WithConfig(cfg),
			logging_fetcher.WithLogger(log),
			logging_fetcher.WithResourceChan(ch),
			logging_fetcher.WithConfigserviceProvider(
				configservice.NewProvider(
					log,
					awsConfig,
					getConfigserviceClients(awsConfigCrossRegionFactory, log, awsConfig),
					*identity.Account,
				),
			),
			logging_fetcher.WithLoggingProvider(logging.NewProvider(
				log,
				awsConfig,
				getCloudrailClients(awsTrailCrossRegionFactory, log, awsConfig),
				getS3Clients(awsS3CrossRegionFactory, log, awsConfig),
			)),
		)
	}

	if _, ok := list[s3_fetcher.Type]; ok {
		reg[s3_fetcher.Type] = s3_fetcher.New(
			s3_fetcher.WithConfig(cfg),
			s3_fetcher.WithLogger(log),
			s3_fetcher.WithResourceChan(ch),
			s3_fetcher.WithS3Provider(s3.NewProvider(
				awsConfig,
				log,
				getS3Clients(awsS3CrossRegionFactory, log, awsConfig),
			)),
		)
	}

	if _, ok := list[rds_fetcher.Type]; ok {
		reg[rds_fetcher.Type] = rds_fetcher.New(
			rds_fetcher.WithConfig(cfg),
			rds_fetcher.WithLogger(log),
			rds_fetcher.WithResourceChan(ch),
			rds_fetcher.WithRDSProvider(rds.NewProvider(
				log,
				awsConfig,
				getRDSClients(awsRDSCrossRegionFactory, log, awsConfig),
			)),
		)
	}

	if _, ok := list[process_fetcher.Type]; ok {
		reg[process_fetcher.Type] = process_fetcher.New(
			process_fetcher.WithConfig(cfg),
			process_fetcher.WithLogger(log),
			process_fetcher.WithResourceChan(ch),
			process_fetcher.WithFSProvider(func(dir string) fs.FS { return os.DirFS(dir) }),
		)
	}

	if _, ok := list[network_fetcher.Type]; ok {
		reg[network_fetcher.Type] = network_fetcher.New(
			network_fetcher.WithLogger(log),
			network_fetcher.WithConfig(cfg),
			network_fetcher.WithCloudIdentity(identity),
			network_fetcher.WithEC2Provider(
				ec2.NewEC2Provider(
					log,
					*identity.Account,
					awsConfig,
					getEC2Clients(awsEC2CrossRegionFactory, log, awsConfig),
				),
			),
		)
	}

	if _, ok := list[elb_fetcher.Type]; ok {
		reg[elb_fetcher.Type] = elb_fetcher.New(
			elb_fetcher.WithLogger(log),
			elb_fetcher.WithConfig(cfg),
			elb_fetcher.WithResourceChan(ch),
			elb_fetcher.WithCloudIdentity(identity),
			elb_fetcher.WithElbProvider(awslib.NewElbProvider(awsELBService)),
			elb_fetcher.WithKubeClient(k8sClient),
			elb_fetcher.WithRegexMatcher(awsConfig.Region),
		)
	}

	if _, ok := list[eks_fetcher.Type]; ok {
		reg[eks_fetcher.Type] = eks_fetcher.New(
			eks_fetcher.WithLogger(log),
			eks_fetcher.WithConfig(cfg),
			eks_fetcher.WithResourceChan(ch),
			eks_fetcher.WithEKSProvider(awslib.NewEksProvider(awsEKSService)),
		)
	}

	if _, ok := list[ecr_fetcher.Type]; ok {
		reg[ecr_fetcher.Type] = ecr_fetcher.New(
			ecr_fetcher.WithLogger(log),
			ecr_fetcher.WithConfig(cfg),
			ecr_fetcher.WithResourceChan(ch),
			ecr_fetcher.WithKubeClient(k8sClient),
			ecr_fetcher.WithECRProvider(
				awslib.NewEcrProvider(
					getEcrClients(awsECRCrossRegionFactory, log, awsConfig),
				),
				*identity.Account,
			),
		)
	}
	return reg, nil
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
		iamProvider := iam.NewIAMProvider(log, iam_sdk.NewFromConfig(awsConfig))

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

func getCloudrailClients(factory awslib.CrossRegionFactory[cloudtrail.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]cloudtrail.Client {
	f := func(cfg aws_sdk.Config) cloudtrail.Client {
		return cloudtrail_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getCloudwatchClients(factory awslib.CrossRegionFactory[cloudwatch.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]cloudwatch.Client {
	f := func(cfg aws_sdk.Config) cloudwatch.Client {
		return cloudwatch_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getCloudwatchlogsClients(factory awslib.CrossRegionFactory[logs.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]logs.Client {
	f := func(cfg aws_sdk.Config) logs.Client {
		return cloudwatchlogs_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getSecurityhubClients(factory awslib.CrossRegionFactory[securityhub.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]securityhub.Client {
	f := func(cfg aws_sdk.Config) securityhub.Client {
		return securityhub_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getSNSClients(factory awslib.CrossRegionFactory[sns.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]sns.Client {
	f := func(cfg aws_sdk.Config) sns.Client {
		return sns_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getConfigserviceClients(factory awslib.CrossRegionFactory[configservice.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]configservice.Client {
	f := func(cfg aws_sdk.Config) configservice.Client {
		return configservice_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getS3Clients(factory awslib.CrossRegionFactory[s3.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]s3.Client {
	f := func(cfg aws_sdk.Config) s3.Client {
		return s3_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getRDSClients(factory awslib.CrossRegionFactory[rds.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]rds.Client {
	f := func(cfg aws_sdk.Config) rds.Client {
		return rds_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getEC2Clients(factory awslib.CrossRegionFactory[ec2.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]ec2.Client {
	f := func(cfg aws_sdk.Config) ec2.Client {
		return ec2_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}

func getEcrClients(factory awslib.CrossRegionFactory[*ecr_sdk.Client], log *logp.Logger, cfg aws_sdk.Config) map[string]*ecr_sdk.Client {
	f := func(cfg aws_sdk.Config) *ecr_sdk.Client {
		return ecr_sdk.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ec2_sdk.NewFromConfig(cfg), cfg, f, log)
	return m.GetMultiRegionsClientMap()
}
