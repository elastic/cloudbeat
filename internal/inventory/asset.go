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
)

// assetSubCategory is used to build the document index. Use only numbers, letters and dashes (-)
type assetSubCategory string

const (
	SubCategoryCompute assetSubCategory = "compute"
)

// assetType is used to build the document index. Use only numbers, letters and dashes (-)
type assetType string

const (
	TypeVirtualMachine assetType = "virtual-machine"
)

// assetSubType is used to build the document index. Use only numbers, letters and dashes (-)
type assetSubType string

const (
	SubTypeEC2 assetSubType = "ec2"
)

type assetCloudProvider string

const (
	AwsCloudProvider assetCloudProvider = "aws"
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Asset   Asset
	Network *AssetNetwork
	Cloud   *AssetCloud
	Host    *AssetHost
	IAM     *AssetIAM
}

// AssetClassification holds the taxonomy of an asset
type AssetClassification struct {
	Category    assetCategory    `json:"category"`
	SubCategory assetSubCategory `json:"subCategory"`
	Type        assetType        `json:"type"`
	SubStype    assetSubType     `json:"subStype"`
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
	NetworkId        *string `json:"networkId"`
	SubnetId         *string `json:"subnetId"`
	Ipv6Address      *string `json:"ipv6Address"`
	PublicIpAddress  *string `json:"publicIpAddress"`
	PrivateIpAddress *string `json:"privateIpAddress"`
	PublicDnsName    *string `json:"publicDnsName"`
	PrivateDnsName   *string `json:"privateDnsName"`
}

// AssetCloud contains information about the cloud provider
type AssetCloud struct {
	Provider assetCloudProvider `json:"provider"`
	Region   string             `json:"region"`
}

// AssetHost contains information of the asset in case it is a host
type AssetHost struct {
	Architecture    string  `json:"architecture"`
	ImageId         *string `json:"imageId"`
	InstanceType    string  `json:"instanceType"`
	Platform        string  `json:"platform"`
	PlatformDetails *string `json:"platformDetails"`
}

type AssetIAM struct {
	Id  *string `json:"id"`
	Arn *string `json:"arn"`
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

func EmptyEnricher() AssetEnricher {
	return func(_ *AssetEvent) {}
}

func generateUniqueId(c AssetClassification, resourceId string) string {
	hasher := sha256.New()
	toBeHashed := fmt.Sprintf("%s-%s-%s-%s-%s", resourceId, c.Category, c.SubCategory, c.Type, c.SubStype)
	hasher.Write([]byte(toBeHashed)) //nolint:revive
	hash := hasher.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(hash)
	return encoded
}
