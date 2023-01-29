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

package dataprovider

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

type commonDataProvider struct {
	log                 *logp.Logger
	cfg                 *config.Config
	k8sDataProviderInit func(*logp.Logger, *config.Config) EnvironmentCommonDataProvider
	awsDataProviderInit func(*logp.Logger, *config.Config) (EnvironmentCommonDataProvider, error)
}

func NewCommonDataProvider(log *logp.Logger, cfg *config.Config) CommonDataProvider {
	return commonDataProvider{log, cfg, NewK8sDataProvider, NewAwsDataProvider}
}

func (c commonDataProvider) FetchCommonData(ctx context.Context) (CommonData, error) {
	if c.cfg.Benchmark == "cis_eks" || c.cfg.Benchmark == "cis_k8s" {
		return c.k8sDataProviderInit(c.log, c.cfg).FetchData(ctx)
	}

	if c.cfg.Benchmark == "cis_aws" {
		dataProvider, err := c.awsDataProviderInit(c.log, c.cfg)
		if err != nil {
			return nil, err
		}

		return dataProvider.FetchData(ctx)
	}

	return nil, fmt.Errorf("could not get common data provider for benchmark %s", c.cfg.Benchmark)
}
