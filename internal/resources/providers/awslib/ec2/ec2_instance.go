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

	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

type Ec2Instance struct {
	types.Instance
	Region     string
	awsAccount string
	RootVolume *Volume
}

type SecurityGroupInfo struct {
	GroupId   *string `json:"group_id,omitempty"`
	GroupName *string `json:"group_name,omitempty"`
}

func (i Ec2Instance) GetResourceArn() string {
	if i.Instance.InstanceId == nil {
		return ""
	}
	// TODO: check if this is the correct ARN
	return fmt.Sprintf("arn:aws:ec2:%s:%s:ec2/%s", i.Region, i.awsAccount, *i.Instance.InstanceId)
}

func (i Ec2Instance) GetResourceName() string {
	for _, tag := range i.Instance.Tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}

	return ""
}

func (i Ec2Instance) GetResourceId() string {
	if i.Instance.InstanceId == nil {
		return ""
	}

	return *i.Instance.InstanceId
}

func (i Ec2Instance) GetResourceType() string {
	return fetching.EC2Type
}

// TODO: Use genertic implementation with custom functions
func (i Ec2Instance) GetResourceTags() map[string]string {
	instanceTags := make(map[string]string, len(i.Tags))
	for _, tag := range i.Tags {
		instanceTags[*tag.Key] = *tag.Value
	}
	return instanceTags
}

// TODO: Use genertic implementation with custom functions
func (i Ec2Instance) GetResourceMacAddresses() []string {
	macAddresses := make([]string, len(i.NetworkInterfaces))
	for i, iface := range i.NetworkInterfaces {
		macAddresses[i] = *iface.MacAddress
	}
	return macAddresses
}

// TODO: Use genertic implementation with custom functions
func (i Ec2Instance) GetResourceSecurityGroups() []SecurityGroupInfo {
	securityGroups := make([]SecurityGroupInfo, len(i.SecurityGroups))
	for i, group := range i.SecurityGroups {
		securityGroups[i] = SecurityGroupInfo{
			GroupId:   group.GroupId,
			GroupName: group.GroupName,
		}
	}
	return securityGroups
}
