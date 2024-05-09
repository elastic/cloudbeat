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

package inventory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// assetCategory is used to build the document index. Use only numbers, letters and dashes (-)
type assetCategory string

const (
	CategoryInfrastructure assetCategory = "infrastructure"
	CategoryIdentity       assetCategory = "identity"
)

// assetSubCategory is used to build the document index. Use only numbers, letters and dashes (-)
type assetSubCategory string

const (
	SubCategoryCompute  assetSubCategory = "compute"
	SubCategoryStorage  assetSubCategory = "storage"
	SubCategoryDatabase assetSubCategory = "database"

	SubCategoryCloudProviderAccount assetSubCategory = "cloud-provider-account"
)

// assetType is used to build the document index. Use only numbers, letters and dashes (-)
type assetType string

const (
	TypeVirtualMachine assetType = "virtual-machine"
	TypeObjectStorage  assetType = "object-storage"
	TypeDatabase       assetType = "database"

	TypeUser           assetType = "user"
	TypeServiceAccount assetType = "service-account"
	TypePermissions    assetType = "permissions"
)

// assetSubType is used to build the document index. Use only numbers, letters and dashes (-)
type assetSubType string

const (
	SubTypeEC2 assetSubType = "ec2"
	SubTypeS3  assetSubType = "s3"
	SubTypeIAM assetSubType = "iam"
	SubTypeRDS assetSubType = "rds"
)

const (
	AwsCloudProvider = "aws"
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Asset            Asset
	Network          *AssetNetwork
	Cloud            *AssetCloud
	Host             *AssetHost
	IAM              *AssetIAM
	ResourcePolicies []AssetResourcePolicy
}

// AssetClassification holds the taxonomy of an asset
type AssetClassification struct {
	Category    assetCategory    `json:"category"`
	SubCategory assetSubCategory `json:"sub_category"`
	Type        assetType        `json:"type"`
	SubType     assetSubType     `json:"sub_type"`
}

// Asset contains the identifiers of the asset
type Asset struct {
	UUID string `json:"uuid"`
	Id   string `json:"id"`
	Name string `json:"name"`
	AssetClassification
	Tags map[string]string `json:"tags"`
	Raw  any               `json:"raw"`
}

// AssetNetwork contains network information
type AssetNetwork struct {
	NetworkId        *string `json:"network_id"`
	SubnetId         *string `json:"subnet_id"`
	Ipv6Address      *string `json:"ipv6_address"`
	PublicIpAddress  *string `json:"public_ip_address"`
	PrivateIpAddress *string `json:"private_ip_address"`
	PublicDnsName    *string `json:"public_dns_name"`
	PrivateDnsName   *string `json:"private_dns_name"`
}

// AssetCloud contains information about the cloud provider
type AssetCloud struct {
	AvailabilityZone *string             `json:"availability_zone,omitempty"`
	Provider         string              `json:"provider,omitempty"`
	Region           string              `json:"region,omitempty"`
	Account          AssetCloudAccount   `json:"account"`
	Instance         *AssetCloudInstance `json:"instance,omitempty"`
	Machine          *AssetCloudMachine  `json:"machine,omitempty"`
	Project          *AssetCloudProject  `json:"project,omitempty"`
	Service          *AssetCloudService  `json:"service,omitempty"`
}

type AssetCloudAccount struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type AssetCloudInstance struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type AssetCloudMachine struct {
	MachineType string `json:"machine_type,omitempty"`
}

type AssetCloudProject struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type AssetCloudService struct {
	Name string `json:"name,omitempty"`
}

// AssetHost contains information of the asset in case it is a host
type AssetHost struct {
	Architecture    string  `json:"architecture"`
	ImageId         *string `json:"imageId"`
	InstanceType    string  `json:"instance_type"`
	Platform        string  `json:"platform"`
	PlatformDetails *string `json:"platform_details"`
}

type AssetIAM struct {
	Id  *string `json:"id"`
	Arn *string `json:"arn"`
}

// AssetResourcePolicy maps security policies applied directly on resources
type AssetResourcePolicy struct {
	Version    *string        `json:"version,omitempty"`
	Id         *string        `json:"id,omitempty"`
	Effect     string         `json:"effect,omitempty"`
	Principal  map[string]any `json:"principal,omitempty"`
	Action     []string       `json:"action,omitempty"`
	NotAction  []string       `json:"notAction,omitempty"`
	Resource   []string       `json:"resource,omitempty"`
	NoResource []string       `json:"noResource,omitempty"`
	Condition  map[string]any `json:"condition,omitempty"`
}

// AssetEnricher functional builder function
type AssetEnricher func(asset *AssetEvent)

func NewAssetEvent(c AssetClassification, id string, name string, enrichers ...AssetEnricher) AssetEvent {
	a := AssetEvent{
		Asset: Asset{
			UUID:                generateUniqueId(c, id),
			Id:                  id,
			Name:                name,
			AssetClassification: c,
		},
	}

	for _, enrich := range enrichers {
		enrich(&a)
	}

	return a
}

func WithRawAsset(raw any) AssetEnricher {
	return func(a *AssetEvent) {
		a.Asset.Raw = &raw
	}
}

func WithTags(tags map[string]string) AssetEnricher {
	return func(a *AssetEvent) {
		if len(tags) == 0 {
			return
		}

		a.Asset.Tags = tags
	}
}

func WithNetwork(network AssetNetwork) AssetEnricher {
	return func(a *AssetEvent) {
		a.Network = &network
	}
}

func WithCloud(cloud AssetCloud) AssetEnricher {
	return func(a *AssetEvent) {
		a.Cloud = &cloud
	}
}

func WithHost(host AssetHost) AssetEnricher {
	return func(a *AssetEvent) {
		a.Host = &host
	}
}

func WithIAM(iam AssetIAM) AssetEnricher {
	return func(a *AssetEvent) {
		a.IAM = &iam
	}
}

func WithResourcePolicies(policies ...AssetResourcePolicy) AssetEnricher {
	return func(a *AssetEvent) {
		if len(policies) == 0 {
			return
		}

		a.ResourcePolicies = policies
	}
}

func EmptyEnricher() AssetEnricher {
	return func(_ *AssetEvent) {}
}

func generateUniqueId(c AssetClassification, resourceId string) string {
	hasher := sha256.New()
	toBeHashed := fmt.Sprintf("%s-%s-%s-%s-%s", resourceId, c.Category, c.SubCategory, c.Type, c.SubType)
	hasher.Write([]byte(toBeHashed)) //nolint:revive
	hash := hasher.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(hash)
	return encoded
}
