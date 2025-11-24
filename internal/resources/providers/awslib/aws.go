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

package awslib

import (
	"errors"

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const (
	DefaultRegion    = "us-east-1"
	DefaultGovRegion = "us-gov-east-1"
	GlobalRegion     = "global"
)

var ErrClientNotFound = errors.New("aws client not found")

type AwsResource interface {
	GetResourceArn() string
	GetResourceName() string
	GetResourceType() string
	GetRegion() string
}

func GetDefaultClient[T any](list map[string]T) (T, error) {
	c, ok := list[DefaultRegion]
	if ok {
		return c, nil
	}

	c, ok = list[DefaultGovRegion]
	if ok {
		return c, nil
	}

	return c, ErrClientNotFound
}

func GetClient[T any](region *string, list map[string]T) (T, error) {
	if region == nil {
		return GetDefaultClient(list)
	}

	c, ok := list[pointers.Deref(region)]
	if !ok {
		return c, ErrClientNotFound
	}

	return c, nil
}

type OrgIAMRoleNamesProvider interface {
	RootRoleName() string
	MemberRoleName() string
}

type BenchmarkOrgIAMRoleNamesProvider struct{}

func (BenchmarkOrgIAMRoleNamesProvider) RootRoleName() string   { return orgBenchmarkRootRole }
func (BenchmarkOrgIAMRoleNamesProvider) MemberRoleName() string { return orgBenchmarkMemberRole }

type AssetDiscoveryOrgIAMRoleNamesProvider struct{}

func (AssetDiscoveryOrgIAMRoleNamesProvider) RootRoleName() string { return assetInventoryRootRole }
func (AssetDiscoveryOrgIAMRoleNamesProvider) MemberRoleName() string {
	return assetInventoryMemberRole
}

const (
	orgBenchmarkRootRole     = "cloudbeat-root"
	orgBenchmarkMemberRole   = "cloudbeat-securityaudit"
	assetInventoryRootRole   = "cloudbeat-asset-inventory-root"
	assetInventoryMemberRole = "cloudbeat-asset-inventory-securityaudit"
)
