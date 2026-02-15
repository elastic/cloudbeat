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

package common

import (
	"github.com/go-viper/mapstructure/v2"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/version"
)

type DataProvider struct {
	info map[string]any
	cfg  *config.Config
}

func New(cloudbeatVersionInfo version.CloudbeatVersionInfo, cfg *config.Config) (*DataProvider, error) {
	m := map[string]any{}
	err := mapstructure.Decode(cloudbeatVersionInfo, &m)
	if err != nil {
		return nil, err
	}

	return &DataProvider{
		info: m,
		cfg:  cfg,
	}, nil
}

func (c *DataProvider) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{}
	m["cloudbeat"] = c.info

	if c.cfg != nil {
		m["cloud_security_posture.package_policy"] = map[string]any{
			"id":       c.cfg.PackagePolicyId,
			"revision": c.cfg.PackagePolicyRevision,
		}
	}

	return m, nil
}
