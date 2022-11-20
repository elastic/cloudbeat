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

package fetchersManager

import (
	"context"
	"fmt"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

type awsTestFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
}

type awsTestFetcher struct {
	resourceCh chan fetching.ResourceInfo
	cfg        awsTestFetcherConfig
}

func (f *awsTestFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.resourceCh <- fetching.ResourceInfo{
		Resource:      awsTestResource{AwsConfig: f.cfg.AwsConfig},
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f awsTestFetcher) Stop() {
}

type awsTestResource struct {
	AwsConfig aws.ConfigAWS
}

func (a awsTestResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{}, nil
}

func (a awsTestResource) GetData() any {
	return a.AwsConfig
}

func (a awsTestResource) GetElasticCommonData() any {
	return nil
}

type awsTestFactory struct{}

func (n *awsTestFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	cfg := awsTestFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return &awsTestFetcher{ch, cfg}, nil
}

func awsMockedFetcherConfig(s *FactoriesTestSuite, awsConfig aws.ConfigAWS) *agentconfig.C {
	c := agentconfig.NewConfig()
	err := c.Merge(awsConfig)
	s.NoError(err)

	return c
}

func (s *FactoriesTestSuite) TestCreateFetcherWithAwsCredentials() {
	tests := []struct {
		fetcherName string
		awsConfig   aws.ConfigAWS
	}{
		{
			"some_fetcher",
			aws.ConfigAWS{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				SessionToken:    "session",
			},
		},
	}

	for _, test := range tests {
		s.F.RegisterFactory(test.fetcherName, &awsTestFactory{})
		c := awsMockedFetcherConfig(s, test.awsConfig)

		f, err := s.F.CreateFetcher(s.log, test.fetcherName, c, s.resourceCh)
		s.NoError(err)
		err = f.Fetch(context.TODO(), fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(1, len(results))
		s.NoError(err)

		result := results[0].GetData().(aws.ConfigAWS)
		s.Equal(test.awsConfig.AccessKeyID, result.AccessKeyID)
		s.Equal(test.awsConfig.SecretAccessKey, result.SecretAccessKey)
		s.Equal(test.awsConfig.SessionToken, result.SessionToken)
	}
}

func (s *FactoriesTestSuite) TestRegisterFetchersWithAwsCredentials() {
	tests := []struct {
		fetcherName string
		awsConfig   aws.ConfigAWS
	}{
		{
			"some_fetcher",
			aws.ConfigAWS{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				SessionToken:    "session",
			},
		},
		{
			"another_fetcher",
			aws.ConfigAWS{
				AccessKeyID:     "new_key",
				SecretAccessKey: "new_secret",
				SessionToken:    "new_session",
			},
		},
	}

	for _, test := range tests {
		s.F = newFactories()
		s.F.RegisterFactory(test.fetcherName, &awsTestFactory{})
		reg := NewFetcherRegistry(s.log)
		conf := createEksAgentConfig(s, test.awsConfig, test.fetcherName)
		parsedList, err := s.F.ParseConfigFetchers(s.log, conf, s.resourceCh)
		s.Equal(test.fetcherName, parsedList[0].name)
		s.NoError(err)

		err = reg.RegisterFetchers(parsedList, nil)
		s.NoError(err)
		s.Equal(1, len(reg.Keys()))

		err = reg.Run(context.Background(), test.fetcherName, fetching.CycleMetadata{})
		s.NoError(err)

		results := testhelper.CollectResources(s.resourceCh)
		result := results[0].GetData().(aws.ConfigAWS)
		s.Equal(test.awsConfig.AccessKeyID, result.AccessKeyID)
		s.Equal(test.awsConfig.SecretAccessKey, result.SecretAccessKey)
		s.Equal(test.awsConfig.SessionToken, result.SessionToken)
	}
}

func createEksAgentConfig(s *FactoriesTestSuite, awsConfig aws.ConfigAWS, fetcherName string) *config.Config {
	conf := &config.Config{
		Type:       config.InputTypeEks,
		AWSConfig:  awsConfig,
		RuntimeCfg: nil,
		Fetchers:   []*agentconfig.C{agentconfig.MustNewConfigFrom(fmt.Sprint("name: ", fetcherName))},
	}

	return conf
}
