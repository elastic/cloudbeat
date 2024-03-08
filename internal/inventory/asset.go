package inventory

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type assetCategory string
type assetSubCategory string
type assetType string
type assetSubType string
type assetCloudProvider string

const (
	CategoryInfrastructure assetCategory = "infrastructure"

	SubCategoryCompute assetSubCategory = "compute"

	TypeVirtualMachine assetType = "virtual-machine"

	SubTypeEC2 assetSubType = "ec2"

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
	return func(asset *AssetEvent) {

	}
}

func generateUniqueId(c AssetClassification, resourceId string) string {
	hasher := sha256.New()
	toBeHashed := fmt.Sprintf("%s-%s-%s-%s-%s", resourceId, c.Category, c.SubCategory, c.Type, c.SubStype)
	hasher.Write([]byte(toBeHashed)) //nolint:revive
	hash := hasher.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(hash)
	return encoded
}
