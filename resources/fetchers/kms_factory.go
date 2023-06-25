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

package fetchers

import (
	"fmt"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	agentConfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.KmsType, &KmsFactory{
		CrossRegionFactory: &awslib.MultiRegionClientFactory[kms.Client]{},
	})
}

type KmsFactory struct {
	CrossRegionFactory awslib.CrossRegionFactory[kms.Client]
}

func (f *KmsFactory) Create(log *logp.Logger, c *agentConfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting KmsFactory.Create")

	cfg := KmsFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	return f.CreateFrom(log, cfg, ch)
}

func (f *KmsFactory) CreateFrom(log *logp.Logger, cfg KmsFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	kmsProvider := kms.NewKMSProvider(awsConfig, log, f.CrossRegionFactory)

	return &KmsFetcher{
		log:        log,
		cfg:        cfg,
		kms:        kmsProvider,
		resourceCh: ch,
	}, nil
}
