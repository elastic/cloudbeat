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

// AssetCategory is used to build the document index. Use only numbers, letters and dashes (-)
type AssetCategory string

const (
	CategoryInfrastructure AssetCategory = "infrastructure"
	CategoryIdentity       AssetCategory = "identity"
)

// AssetSubCategory is used to build the document index. Use only numbers, letters and dashes (-)
type AssetSubCategory string

const (
	SubCategoryAuthorization AssetSubCategory = "authorization"
	SubCategoryCompute       AssetSubCategory = "compute"
	SubCategoryStorage       AssetSubCategory = "storage"
	SubCategoryDatabase      AssetSubCategory = "database"
	SubCategoryNetwork       AssetSubCategory = "network"

	SubCategoryCloudProviderAccount AssetSubCategory = "cloud-provider-account"
)

// AssetType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetType string

const (
	TypeAcl                AssetType = "acl"
	TypeFirewall           AssetType = "firewall"
	TypeInterface          AssetType = "interface"
	TypeObjectStorage      AssetType = "object-storage"
	TypePeering            AssetType = "peering"
	TypeRelationalDatabase AssetType = "relational-database"
	TypeSubnet             AssetType = "subnet"
	TypeVirtualMachine     AssetType = "virtual-machine"
	TypeVirtualNetwork     AssetType = "virtual-network"

	TypePermissions    AssetType = "permissions"
	TypeServiceAccount AssetType = "service-account"
	TypeUser           AssetType = "user"
)

// AssetSubType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetSubType string

const (
	SubTypeEC2                      AssetSubType = "ec2"
	SubTypeS3                       AssetSubType = "s3"
	SubTypeIAM                      AssetSubType = "iam"
	SubTypeRDS                      AssetSubType = "rds"
	SubTypeEC2NetworkInterface      AssetSubType = "ec2-network-interface"
	SubTypeEC2Subnet                AssetSubType = "ec2-subnet"
	SubTypeInternetGateway          AssetSubType = "internet-gateway"
	SubTypeNatGateway               AssetSubType = "nat-gateway"
	SubTypeSecurityGroup            AssetSubType = "security-group"
	SubTypeTransitGateway           AssetSubType = "transit-gateway"
	SubTypeTransitGatewayAttachment AssetSubType = "transit-gateway-attachment"
	SubTypeVpc                      AssetSubType = "vpc"
	SubTypeVpcAcl                   AssetSubType = "vpc-acl"
	SubTypeVpcPeeringConnection     AssetSubType = "vpc-peering-connections"
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
	Category    AssetCategory    `json:"category"`
	SubCategory AssetSubCategory `json:"sub_category"`
	Type        AssetType        `json:"type"`
	SubType     AssetSubType     `json:"sub_type"`
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
