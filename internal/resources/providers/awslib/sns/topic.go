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

package sns

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type TopicInfo struct {
	Topic         types.Topic          `json:"topic"`
	Subscriptions []types.Subscription `json:"subscriptions"`
	region        string
}

func (v TopicInfo) GetResourceArn() string {
	return pointers.Deref(v.Topic.TopicArn)
}

func (v TopicInfo) GetResourceName() string {
	elems := strings.Split(v.GetResourceArn(), ":")
	return elems[len(elems)-1]
}

func (v TopicInfo) GetResourceType() string {
	return fetching.SNSTopicType
}

func (v TopicInfo) GetRegion() string {
	return v.region
}
