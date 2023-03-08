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

package aws

import (
	"github.com/elastic/elastic-agent-libs/logp"
)

type Option func(*DataProvider)

func WithLogger(log *logp.Logger) Option {
	return func(dp *DataProvider) {
		dp.log = log
	}
}

func WithAccount(name string, id string) Option {
	return func(dp *DataProvider) {
		dp.accountId = id
		dp.accountName = name
	}
}
