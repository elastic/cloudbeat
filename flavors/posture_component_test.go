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

//go:build component
// +build component

package flavors

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	rds_sdk "github.com/aws/aws-sdk-go-v2/service/rds"
	s3_sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/cloudbeat/config"
	aws_dataprovider "github.com/elastic/cloudbeat/dataprovider/providers/aws"
	"github.com/elastic/cloudbeat/evaluator"
	posture_testing "github.com/elastic/cloudbeat/flavors/testing"
	rds_fetcher "github.com/elastic/cloudbeat/resources/fetchers/rds"
	s3_fetcher "github.com/elastic/cloudbeat/resources/fetchers/s3"
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	awslib_rds "github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	awslib_s3 "github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/transformer"
	"github.com/elastic/cloudbeat/uniqueness"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Cloudbeat(t *testing.T) {
	pwd, _ := os.Getwd()
	// Setup cloudbeat configuration
	cfg := config.Config{
		BundlePath: path.Join(pwd, "..", "bundle.tar.gz"),
		Benchmark:  config.CIS_AWS,
		Fetchers: []*agentconfig.C{
			agentconfig.MustNewConfigFrom(mapstr.M{"name": "aws-s3"}),
			agentconfig.MustNewConfigFrom(mapstr.M{"name": "aws-rds"}),
		},
		Processors: processors.PluginConfig{},
	}

	// Create posture instance
	// This code replace completly the NewPosture as all the dependencies initiated here
	cloudbeat, err := newTestingCloudbeat(t, cfg)
	assert.NoError(t, err)

	// Create testing beater
	// beater hold the connection to Elasticsearch and Fleet
	// those connections are mocked for assertion
	beater, collectPublished := newTestingBeat()

	// Start cloudbeat
	// error not checked and ignored in golangci-lint
	go cloudbeat.Run(beater)
	// assert.NoError(t, err)

	// Sleep few seconds, than stop cloudbeat and "collect" all the published events
	<-time.After(time.Second * 2)
	cloudbeat.Stop()

	published := collectPublished()
	// Assert the outgoing events
	assert.Len(t, published, 10)

	rules := lo.Keys(
		lo.GroupBy(published, func(e beat.Event) string {
			r, err := e.Fields.GetValue("rule")
			assert.NoError(t, err)
			rule := r.(evaluator.Rule)
			return rule.Benchmark.Rule_Number
		}),
	)
	assert.ElementsMatch(t, rules, []string{
		"2.1.1",
		"2.1.3",
		"2.3.1",
		"2.3.2",
	})

	resources := lo.Keys(
		lo.GroupBy(published, func(e beat.Event) string {
			r, err := e.Fields.GetValue("resource")
			assert.NoError(t, err)
			info := r.(fetching.ResourceFields)
			return info.SubType
		}),
	)

	assert.ElementsMatch(t, resources, []string{
		"aws-s3",
		"aws-rds",
	})

	passed := lo.GroupBy(published, func(e beat.Event) string {
		r, err := e.Fields.GetValue("result")
		assert.NoError(t, err)
		result := r.(evaluator.Result)
		return result.Evaluation
	})

	assert.Len(t, passed["passed"], 5)
	assert.Len(t, passed["failed"], 5)
}

func newTestingCloudbeat(t *testing.T, cfg config.Config) (*posture, error) {
	ctx, cancel := context.WithCancel(context.Background())
	log := logp.NewLogger("test")
	resourceChan := make(chan fetching.ResourceInfo)
	leader := uniqueness.MockManager{}
	leader.EXPECT().Run(mock.Anything).Return(nil)
	leader.EXPECT().Stop().Return()
	// awsConfig, _ := aws_config.LoadDefaultConfig(context.Background())
	awsConfig := aws.Config{
		Region: awslib.DefaultRegion,
	}
	awsAccountName := "test-account-name"
	awsAccountID := "test-account-id"
	apiOption := mockAWSCalls(map[string]interface{}{
		"aws/s3/bucket_list.json":                           &s3_sdk.ListBucketsOutput{},
		"aws/s3/get_bucket_location/bucket_1.json":          &s3_sdk.GetBucketLocationOutput{},
		"aws/s3/get_bucket_location/bucket_2.json":          &s3_sdk.GetBucketLocationOutput{},
		"aws/s3/get_bucket_location/bucket_3.json":          &s3_sdk.GetBucketLocationOutput{},
		"aws/s3/get_bucket_encryption/bucket_1.json":        &s3_sdk.GetBucketEncryptionOutput{},
		"aws/s3/get_bucket_encryption/bucket_2.json":        &s3_sdk.GetBucketEncryptionOutput{},
		"aws/s3/get_bucket_encryption/bucket_3.json":        &s3_sdk.GetBucketEncryptionOutput{},
		"aws/s3/get_bucket_policy/bucket_1.json":            &s3_sdk.GetBucketPolicyOutput{},
		"aws/s3/get_bucket_policy/bucket_2.json":            &s3_sdk.GetBucketPolicyOutput{},
		"aws/s3/get_bucket_policy/bucket_3.json":            &s3_sdk.GetBucketPolicyOutput{},
		"aws/s3/get_bucket_versioning/bucket_1.json":        &s3_sdk.GetBucketVersioningOutput{},
		"aws/s3/get_bucket_versioning/bucket_2.json":        &s3_sdk.GetBucketVersioningOutput{},
		"aws/s3/get_bucket_versioning/bucket_3.json":        &s3_sdk.GetBucketVersioningOutput{},
		"aws/rds/us-east-1_describe_db_instances.json":      &rds_sdk.DescribeDBInstancesOutput{},
		"aws/rds/us-west-2_describe_db_instances.json":      &rds_sdk.DescribeDBInstancesOutput{},
		"aws/rds/eu-central-1_describe_db_instances.json":   &rds_sdk.DescribeDBInstancesOutput{},
		"aws/rds/ap-southeast-2_describe_db_instances.json": &rds_sdk.DescribeDBInstancesOutput{},
	}, t)
	reg, err := initRegistry(ctx, log, &cfg, resourceChan, &leader, map[string]fetching.Fetcher{
		s3_fetcher.Type: s3_fetcher.New(
			s3_fetcher.WithS3Provider(awslib_s3.NewProvider(awsConfig, log, awsFakeS3Clients(func(o *s3_sdk.Options) {
				o.APIOptions = append(o.APIOptions, apiOption)
			}))),
			s3_fetcher.WithResourceChan(resourceChan),
			s3_fetcher.WithLogger(log),
			s3_fetcher.WithConfig(&cfg),
		),
		rds_fetcher.Type: rds_fetcher.New(
			rds_fetcher.WithRDSProvider(awslib_rds.NewProvider(log, awsConfig, awsFakeRDSClients(func(o *rds_sdk.Options) {
				o.APIOptions = append(o.APIOptions, apiOption)
			}))),
			rds_fetcher.WithConfig(&cfg),
			rds_fetcher.WithLogger(log),
			rds_fetcher.WithResourceChan(resourceChan),
		),
	})
	if err != nil {
		cancel()
		return nil, err
	}
	data, err := fetchersManager.NewData(log, time.Second*30, time.Second*5, reg)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log, &cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	return &posture{
		flavorBase: flavorBase{
			ctx:                 ctx,
			cancel:              cancel,
			config:              &cfg,
			log:                 log,
			flushInterval:       time.Second,
			shutdownGracePeriod: time.Second,
			transformer: transformer.NewTransformer(log, aws_dataprovider.New(
				aws_dataprovider.WithLogger(log),
				aws_dataprovider.WithAccount(awsAccountName, awsAccountID),
			), "test-index"),
		},
		leader:     &leader,
		data:       data,
		evaluator:  eval,
		resourceCh: resourceChan,
	}, nil
}

func newTestingBeat() (*beat.Beat, func() []beat.Event) {
	p := &posture_testing.MockPipeline{}
	c := &posture_testing.MockClient{}
	p.EXPECT().ConnectWith(mock.Anything).Return(c, nil)
	c.EXPECT().PublishAll(mock.Anything).Return()
	c.EXPECT().Close().Return(nil)
	published := func() []beat.Event {
		published := []beat.Event{}
		for _, call := range c.Calls {
			for _, arg := range call.Arguments {
				published = append(published, arg.([]beat.Event)...)
			}
		}
		return published
	}
	return &beat.Beat{Publisher: p}, published
}

func awsMockedConfigs() map[string]aws.Config {
	return map[string]aws.Config{
		awslib.DefaultRegion: {
			Region: awslib.DefaultRegion,
		},
		"us-west-2": {
			Region: "us-west-2",
		},
		"eu-central-1": {
			Region: "eu-central-1",
		},
		"ap-southeast-2": {
			Region: "ap-southeast-2",
		},
	}
}

func awsFakeS3Clients(opts ...func(*s3_sdk.Options)) map[string]awslib_s3.Client {
	res := map[string]awslib_s3.Client{}
	for region, cfg := range awsMockedConfigs() {
		res[region] = s3_sdk.NewFromConfig(cfg, opts...)
	}
	return res
}

func awsFakeRDSClients(opts ...func(*rds_sdk.Options)) map[string]awslib_rds.Client {
	res := map[string]awslib_rds.Client{}
	for region, cfg := range awsMockedConfigs() {
		res[region] = rds_sdk.NewFromConfig(cfg, opts...)
	}
	return res
}
