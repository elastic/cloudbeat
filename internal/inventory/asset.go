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

// AssetCategory is used to build the document index. Use only numbers, letters and dashes (-)
type AssetCategory string

const (
	CategoryIdentity       AssetCategory = "identity"
	CategoryInfrastructure AssetCategory = "infrastructure"
)

// AssetSubCategory is used to build the document index. Use only numbers, letters and dashes (-)
type AssetSubCategory string

const (
	SubCategoryAuthorization        AssetSubCategory = "authorization"
	SubCategoryCloudProviderAccount AssetSubCategory = "cloud-provider-account"
	SubCategoryCompute              AssetSubCategory = "compute"
	SubCategoryDatabase             AssetSubCategory = "database"
	SubCategoryIntegration          AssetSubCategory = "integration"
	SubCategoryMessaging            AssetSubCategory = "messaging"
	SubCategoryNetwork              AssetSubCategory = "network"
	SubCategoryStorage              AssetSubCategory = "storage"
)

// AssetType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetType string

const (
	TypeAcl                 AssetType = "acl"
	TypeEventSource         AssetType = "event-type"
	TypeFirewall            AssetType = "firewall"
	TypeInterface           AssetType = "interface"
	TypeLoadBalancer        AssetType = "load-balancer"
	TypeNotificationService AssetType = "notification-service"
	TypeObjectStorage       AssetType = "object-storage"
	TypePeering             AssetType = "peering"
	TypePermissions         AssetType = "permissions"
	TypeRelationalDatabase  AssetType = "relational-database"
	TypeServerless          AssetType = "serverless"
	TypeServiceAccount      AssetType = "service-account"
	TypeSubnet              AssetType = "subnet"
	TypeUser                AssetType = "user"
	TypeVirtualMachine      AssetType = "virtual-machine"
	TypeVirtualNetwork      AssetType = "virtual-network"
)

// AssetSubType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetSubType string

const (
	SubTypeEC2                      AssetSubType = "ec2"
	SubTypeS3                       AssetSubType = "s3"
	SubTypeIAM                      AssetSubType = "iam"
	SubTypeEC2NetworkInterface      AssetSubType = "ec2-network-interface"
	SubTypeEC2Subnet                AssetSubType = "ec2-subnet"
	SubTypeELBv1                    AssetSubType = "elastic-load-balancer"
	SubTypeELBv2                    AssetSubType = "elastic-load-balancer-v2"
	SubTypeInternetGateway          AssetSubType = "internet-gateway"
	SubTypeLambdaAlias              AssetSubType = "lambda-function-alias"
	SubTypeLambdaEventSourceMapping AssetSubType = "lambda-event-source-mapping"
	SubTypeLambdaFunction           AssetSubType = "lambda-function"
	SubTypeLambdaLayer              AssetSubType = "lambda-layer"
	SubTypeNatGateway               AssetSubType = "nat-gateway"
	SubTypeRDS                      AssetSubType = "rds"
	SubTypeSecurityGroup            AssetSubType = "security-group"
	SubTypeTransitGateway           AssetSubType = "transit-gateway"
	SubTypeTransitGatewayAttachment AssetSubType = "transit-gateway-attachment"
	SubTypeVpc                      AssetSubType = "vpc"
	SubTypeSNSTopic                 AssetSubType = "sns-topic"
	SubTypeVpcAcl                   AssetSubType = "vpc-acl"
	SubTypeVpcPeeringConnection     AssetSubType = "vpc-peering-connections"
)

const (
	AwsCloudProvider = "aws"
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Entity           Entity
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

// Entity contains the identifiers of the asset
type Entity struct {
	Identifiers EntityIdentifiers `json:"identifiers"`
	Name        string            `json:"name"`
	AssetClassification
	Tags map[string]string `json:"tags"`
	Raw  any               `json:"raw"`
}

// Specific identifiers of the asset
type EntityIdentifiers struct {
	Arns []string `json:"arns,omitempty"`
	Ids  []string `json:"ids,omitempty"`
}

// AssetNetwork contains network information
type AssetNetwork struct {
	Ipv6Address         *string  `json:"ipv6_address,omitempty"`
	NetworkId           *string  `json:"network_id,omitempty"`
	NetworkInterfaceIds []string `json:"network_interface_ids,omitempty"`
	PrivateDnsName      *string  `json:"private_dns_name,omitempty"`
	PrivateIpAddress    *string  `json:"private_ip_address,omitempty"`
	PublicDnsName       *string  `json:"public_dns_name,omitempty"`
	PublicIpAddress     *string  `json:"public_ip_address,omitempty"`
	RouteTableIds       []string `json:"route_table_ids,omitempty"`
	SecurityGroupIds    []string `json:"security_group_ids,omitempty"`
	SubnetIds           []string `json:"subnet_ids,omitempty"`
	TransitGatewayIds   []string `json:"transit_gateway_ids,omitempty"`
	VpcIds              []string `json:"vpc_ids,omitempty"`
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

func NewAssetEvent(c AssetClassification, identifiers EntityIdentifiers, name string, enrichers ...AssetEnricher) AssetEvent {
	a := AssetEvent{
		Entity: Entity{
			Identifiers:         identifiers,
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
		a.Entity.Raw = &raw
	}
}

func WithTags(tags map[string]string) AssetEnricher {
	return func(a *AssetEvent) {
		if len(tags) == 0 {
			return
		}

		a.Entity.Tags = tags
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
