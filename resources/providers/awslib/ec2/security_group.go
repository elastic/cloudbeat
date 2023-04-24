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

type SecurityGroup struct {
	types.SecurityGroup
	awsAccount string
	region     string
}

func (s SecurityGroup) GetResourceArn() string {
	if s.SecurityGroup.GroupId == nil {
		return ""
	}
	return fmt.Sprintf("arn:aws:ec2:%s:%s:security-group/%s", s.region, s.awsAccount, *s.SecurityGroup.GroupId)
}

func (s SecurityGroup) GetResourceName() string {
	if s.SecurityGroup.GroupName == nil {
		return ""
	}
	return *s.SecurityGroup.GroupName
}

func (s SecurityGroup) GetResourceType() string {
	return fetching.SecurityGroupType
}

func (s SecurityGroup) GetRegion() string {
	return s.region
}
