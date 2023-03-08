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

package monitoring

import (
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Option func(*MonitoringFetcher)

func WithLogger(log *logp.Logger) Option {
	return func(i *MonitoringFetcher) {
		i.log = log
	}
}

func WithConfig(c *config.Config) Option {
	return func(mf *MonitoringFetcher) {
		cfg := MonitoringFetcherConfig{}
		if err := config.UnpackInto(c, Type, &cfg); err != nil {
			panic(err)
		}
		mf.cfg = cfg
	}
}

func WithResourceChan(ch chan fetching.ResourceInfo) Option {
	return func(mf *MonitoringFetcher) {
		mf.resourceCh = ch
	}
}

func WithCloudIdentity(identity *awslib.Identity) Option {
	return func(mf *MonitoringFetcher) {
		mf.cloudIdentity = identity
	}
}

func WithMonitoringProvider(c monitoring.Client) Option {
	return func(mf *MonitoringFetcher) {
		mf.provider = c
	}
}

func WithSecurityhubService(svc securityhub.Service) Option {
	return func(mf *MonitoringFetcher) {
		mf.securityhub = svc
	}
}
