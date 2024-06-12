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

package lambda

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type LayerInfo struct {
	Layer  types.LayersListItem `json:"layer"`
	region string
}

func (v LayerInfo) GetResourceArn() string {
	return pointers.Deref(v.Layer.LayerArn)
}

func (v LayerInfo) GetResourceName() string {
	return pointers.Deref(v.Layer.LayerName)
}

func (v LayerInfo) GetResourceType() string {
	return fetching.LambdaLayerType
}

func (v LayerInfo) GetRegion() string {
	return v.region
}
