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

package s3

import (
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Option func(*S3Fetcher)

func WithLogger(log *logp.Logger) Option {
	return func(i *S3Fetcher) {
		i.log = log
	}
}

func WithConfig(c *config.Config) Option {
	return func(e *S3Fetcher) {
		cfg := S3FetcherConfig{}
		if err := config.UnpackInto(c, Type, &cfg); err != nil {
			panic(err)
		}
		e.cfg = cfg
	}
}

func WithResourceChan(ch chan fetching.ResourceInfo) Option {
	return func(i *S3Fetcher) {
		i.resourceCh = ch
	}
}

func WithS3Provider(p s3.S3) Option {
	return func(sf *S3Fetcher) {
		sf.s3 = p
	}
}
