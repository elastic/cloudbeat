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
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	IAMType = "aws-iam"
)

func init() {

	manager.Factories.ListFetcherFactory(IAMType, &IAMFactory{
		extraElements: getIamExtraElements,
	})
}

type IAMFactory struct {
	extraElements func(*logp.Logger) (IAMExtraElements, error)
}

type IAMExtraElements struct {
	iamProvider awslib.IAMRolePermissionGetter
}

func (f *IAMFactory) Create(log *logp.Logger, c *config.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting IAMFactory.Create")

	cfg := IAMFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	elements, err := f.extraElements(log)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, elements, ch)
}

func getIamExtraElements(log *logp.Logger) (IAMExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return IAMExtraElements{}, err
	}
	provider := awslib.NewIAMProvider(log, awsConfig.Config)

	return IAMExtraElements{
		iamProvider: provider,
	}, nil
}

func (f *IAMFactory) CreateFrom(log *logp.Logger, cfg IAMFetcherConfig, elements IAMExtraElements, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	return &IAMFetcher{
		log:         log,
		cfg:         cfg,
		iamProvider: elements.iamProvider,
		resourceCh:  ch,
	}, nil

}
