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

type NACLInfo struct {
	types.NetworkAcl
	awsAccount string
	region     string
}

func (r NACLInfo) GetResourceArn() string {
	if r.NetworkAclId == nil {
		return ""
	}
	//arn:aws:ec2:region:account-id:network-acl/network-acl-id
	return fmt.Sprintf("arn:aws:ec2:%s:%s:network-acl/%s", r.region, r.awsAccount, *r.NetworkAclId)
}

func (r NACLInfo) GetResourceName() string {
	if r.NetworkAclId == nil {
		return ""
	}
	return *r.NetworkAclId
}

func (r NACLInfo) GetResourceType() string {
	return fetching.NetworkNACLType
}
