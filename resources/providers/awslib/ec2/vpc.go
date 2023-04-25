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

package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type VpcInfo struct {
	Vpc        types.Vpc       `json:"vpc"`
	FlowLogs   []types.FlowLog `json:"flow_logs"`
	awsAccount string
	region     string
}

func (v VpcInfo) GetResourceArn() string {
	if v.Vpc.VpcId == nil {
		return ""
	}
	return fmt.Sprintf("arn:aws:ec2:%s:%s:vpc/%s", v.region, v.awsAccount, *v.Vpc.VpcId)
}

func (v VpcInfo) GetResourceName() string {
	if v.Vpc.VpcId == nil {
		return ""
	}
	return *v.Vpc.VpcId
}

func (v VpcInfo) GetResourceType() string {
	return fetching.VpcType
}

func (v VpcInfo) GetRegion() string {
	return v.region
}
