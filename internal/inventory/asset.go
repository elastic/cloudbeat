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

type AssetClassification struct {
	Category      assetCategory      `json:"category"`
	SubCategory   assetSubCategory   `json:"subCategory"`
	Type          assetType          `json:"type"`
	SubStype      assetSubType       `json:"subStype"`
	CloudProvider assetCloudProvider `json:"assetCloudProvider"`
}

type Asset struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	ResourceId string `json:"resourceId"`
	AssetClassification
}

func NewAsset(c AssetClassification, resourceId string, name string) Asset {
	return Asset{
		Id:                  generateUniqueId(c, resourceId),
		Name:                name,
		ResourceId:          resourceId,
		AssetClassification: c,
	}
}

func generateUniqueId(c AssetClassification, resourceId string) string {
	hasher := sha256.New()
	toBeHashed := fmt.Sprintf("%s-%s-%s-%s-%s-%s", resourceId, c.CloudProvider, c.Category, c.SubCategory, c.Type, c.SubStype)
	hasher.Write([]byte(toBeHashed))
	hash := hasher.Sum(nil)
	encoded := base64.StdEncoding.EncodeToString(hash)
	return encoded
}
