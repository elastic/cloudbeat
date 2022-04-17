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
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	EKSType = "aws-eks"
)

func init() {
	manager.Factories.ListFetcherFactory(EKSType, &EKSFactory{
		extraElements: getEksExtraElements,
	})
}

type EKSFactory struct {
	extraElements func() (eksExtraElements, error)
}

type eksExtraElements struct {
	eksProvider awslib.EksClusterDescriber
}

func (f *EKSFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	logp.L().Info("EKS factory has started")
	cfg := EKSFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	elements, err := f.extraElements()
	if err != nil {
		return nil, err
	}
	return f.CreateFrom(cfg, elements)
}

func getEksExtraElements() (eksExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return eksExtraElements{}, err
	}

	eks := awslib.NewEksProvider(awsConfig.Config)

	return eksExtraElements{eksProvider: eks}, nil
}

func (f *EKSFactory) CreateFrom(cfg EKSFetcherConfig, elements eksExtraElements) (fetching.Fetcher, error) {
	fe := &EKSFetcher{
		cfg:         cfg,
		eksProvider: elements.eksProvider,
	}

	return fe, nil
}
