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

package cloudtrail

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type TrailInfo struct {
	Trail          types.Trail
	Status         *cloudtrail.GetTrailStatusOutput
	EventSelectors []types.EventSelector
}

func (t TrailInfo) GetResourceArn() string {
	if t.Trail.TrailARN == nil {
		return ""
	}
	return *t.Trail.TrailARN
}

func (t TrailInfo) GetResourceName() string {
	if t.Trail.Name == nil {
		return ""
	}
	return *t.Trail.Name
}

func (t TrailInfo) GetResourceType() string {
	return fetching.TrailType
}
