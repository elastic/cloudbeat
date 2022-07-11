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
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/gofrs/uuid"

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

// Factory can create fetcher instances based on configuration
type Factory interface {
	Create(*logp.Logger, *config.C, chan ResourceInfo) (Fetcher, error)
}

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context, CycleMetadata) error
	Stop()
}

type Condition interface {
	Condition() bool
	Name() string
}

type ResourceInfo struct {
	Resource
	CycleMetadata
}

type CycleMetadata struct {
	CycleId uuid.UUID
}

type Resource interface {
	GetMetadata() ResourceMetadata
	GetData() any
	GetElasticCommonData() any
}

type ResourceFields struct {
	ResourceMetadata
	Raw interface{} `json:"raw"`
}

type ResourceMetadata struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	SubType   string `json:"sub_type"`
	Name      string `json:"name"`
	ECSFormat string `json:"ecsFormat"`
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

type AwsBaseFetcherConfig struct {
	BaseFetcherConfig `config:",inline"`
	AwsConfig         aws.ConfigAWS `config:",inline"`
}

const KubeAPIType = "kube-api"
