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

package fetching

import (
	"context"

	"github.com/elastic/beats/v7/libbeat/common"
)

// Factory can create fetcher instances based on configuration
type Factory interface {
	Create(*common.Config) (Fetcher, error)
}

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context) ([]Resource, error)
	Stop()
}

type Condition interface {
	Condition() bool
	Name() string
}

type Resource interface {
	GetID() (string, error)
	GetData() interface{}
}

type Result struct {
	Type string `json:"type"`
	// Golang 1.18 will introduce generics which will be useful for typing the resource field
	Resource interface{} `json:"resource"`
}

type ResourceMap map[string][]Resource

type BaseFetcherConfig struct {
	Name string `config:"name"`
}

const KubeAPIType = "kube-api"
