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
	SubCategoryAccessManagement       AssetSubCategory = "access-management"
	SubCategoryApplication            AssetSubCategory = "application"
	SubCategoryApplicationIntegration AssetSubCategory = "application-integration"
	SubCategoryAuthorization          AssetSubCategory = "authorization"
	SubCategoryCompute                AssetSubCategory = "compute"
	SubCategoryContainer              AssetSubCategory = "container"
	SubCategoryDatabase               AssetSubCategory = "database"
	SubCategoryDigitalIdentity        AssetSubCategory = "digital-identity"
	SubCategoryIntegration            AssetSubCategory = "integration"
	SubCategoryManagement             AssetSubCategory = "management"
	SubCategoryMessaging              AssetSubCategory = "messaging"
	SubCategoryNetwork                AssetSubCategory = "network"
	SubCategoryServerless             AssetSubCategory = "serverless"
	SubCategoryServiceIdentity        AssetSubCategory = "service-identity"
	SubCategoryStorage                AssetSubCategory = "storage"
)

// AssetType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetType string

const (
	TypeAcl                 AssetType = "acl"
	TypeCloudAccount        AssetType = "cloud-account"
	TypeDisk                AssetType = "disk"
	TypeEventSource         AssetType = "event-source"
	TypeFirewall            AssetType = "firewall"
	TypeGateway             AssetType = "gateway"
	TypeInterface           AssetType = "interface"
	TypeLoadBalancer        AssetType = "load-balancer"
	TypeMessageQueue        AssetType = "message-queue"
	TypeNoSQLDatabase       AssetType = "nosql-database"
	TypeNotificationService AssetType = "notification-service"
	TypeObjectStorage       AssetType = "object-storage"
	TypePeering             AssetType = "peering"
	TypePolicy              AssetType = "policy"
	TypePrincipal           AssetType = "principal"
	TypeRegistry            AssetType = "registry"
	TypeRelationalDatabase  AssetType = "relational"
	TypeResourceGroup       AssetType = "resource-group"
	TypeRole                AssetType = "role"
	TypeScalability         AssetType = "scalability"
	TypeServerless          AssetType = "serverless"
	TypeServiceAccount      AssetType = "service-account"
	TypeServiceAccountKey   AssetType = "service-account-key"
	TypeSnapshot            AssetType = "snapshot"
	TypeStorage             AssetType = "storage"
	TypeSubnet              AssetType = "subnet"
	TypeUser                AssetType = "user"
	TypeVirtualMachine      AssetType = "virtual-machine"
	TypeVirtualNetwork      AssetType = "virtual-network"
	TypeWebApplication      AssetType = "web-application"
	TypeResourceHierarchy   AssetType = "resource-hierarchy"
	TypeOrchestration       AssetType = "orchestration"
	TypeFunction            AssetType = "function"
	TypeLoadBalancing       AssetType = "load-balancing"
	TypeIamRole             AssetType = "iam-role"
)

// AssetSubType is used to build the document index. Use only numbers, letters and dashes (-)
type AssetSubType string

const (
	SubTypeAzureAppService          AssetSubType = "azure-app-service"
	SubTypeAzureContainerRegistry   AssetSubType = "azure-container-registry"
	SubTypeAzureCosmosDBAccount     AssetSubType = "azure-cosmos-db-account"
	SubTypeAzureCosmosDBSQLDatabase AssetSubType = "azure-cosmos-db-sql-database"
	SubTypeAzureDisk                AssetSubType = "azure-disk"
	SubTypeAzureElasticPool         AssetSubType = "azure-elastic-pool"
	SubTypeAzurePrincipal           AssetSubType = "azure-principal"
	SubTypeAzureResourceGroup       AssetSubType = "azure-resource-group"
	SubTypeAzureSQLDatabase         AssetSubType = "azure-sql-database"
	SubTypeAzureSQLServer           AssetSubType = "azure-sql-server"
	SubTypeAzureSnapshot            AssetSubType = "azure-snapshot"
	SubTypeAzureStorageAccount      AssetSubType = "azure-storage-account"
	SubTypeAzureStorageBlobService  AssetSubType = "azure-storage-blob-service"
	SubTypeAzureStorageQueue        AssetSubType = "azure-storage-queue"
	SubTypeAzureStorageQueueService AssetSubType = "azure-storage-queue-service"
	SubTypeAzureSubscription        AssetSubType = "azure-subscription"
	SubTypeAzureTenant              AssetSubType = "azure-tenant"
	SubTypeAzureVirtualMachine      AssetSubType = "azure-virtual-machine"
	SubTypeEC2                      AssetSubType = "ec2-instance"
	SubTypeEC2NetworkInterface      AssetSubType = "ec2-network-interface"
	SubTypeEC2Subnet                AssetSubType = "ec2-subnet"
	SubTypeELBv1                    AssetSubType = "elastic-load-balancer"
	SubTypeELBv2                    AssetSubType = "elastic-load-balancer-v2"
	SubTypeIAMPolicy                AssetSubType = "iam-policy"
	SubTypeIAMRole                  AssetSubType = "iam-role"
	SubTypeIAMUser                  AssetSubType = "iam-user"
	SubTypeInternetGateway          AssetSubType = "internet-gateway"
	SubTypeLambdaAlias              AssetSubType = "lambda-function-alias"
	SubTypeLambdaEventSourceMapping AssetSubType = "lambda-event-source-mapping"
	SubTypeLambdaFunction           AssetSubType = "lambda-function"
	SubTypeLambdaLayer              AssetSubType = "lambda-layer"
	SubTypeNatGateway               AssetSubType = "nat-gateway"
	SubTypeRDS                      AssetSubType = "rds-instance"
	SubTypeS3                       AssetSubType = "s3-bucket"
	SubTypeSNSTopic                 AssetSubType = "sns-topic"
	SubTypeSecurityGroup            AssetSubType = "ec2-security-group"
	SubTypeTransitGateway           AssetSubType = "transit-gateway"
	SubTypeTransitGatewayAttachment AssetSubType = "transit-gateway-attachment"
	SubTypeVpc                      AssetSubType = "vpc"
	SubTypeVpcAcl                   AssetSubType = "s3-access-control-list"
	SubTypeVpcPeeringConnection     AssetSubType = "vpc-peering-connection"
	SubTypeGcpProject               AssetSubType = "gcp-project"
	SubTypeGcpInstance              AssetSubType = "gcp-instance"
	SubTypeGcpSubnet                AssetSubType = "gcp-subnet"
	SubTypeGcpFirewall              AssetSubType = "gcp-firewall"
	SubTypeGcpBucket                AssetSubType = "gcp-bucket"
	SubTypeGcpOrganization          AssetSubType = "gcp-organization"
	SubTypeGcpFolder                AssetSubType = "gcp-folder"
	SubTypeGcpServiceAccount        AssetSubType = "gcp-service-account"
	SubTypeGcpServiceAccountKey     AssetSubType = "gcp-service-account-key"
	SubTypeGcpGkeCluster            AssetSubType = "gke-cluster"
	SubTypeGcpForwardingRule        AssetSubType = "gcp-forwarding-rule"
	SubTypeGcpCloudFunction         AssetSubType = "gcp-cloud-function"
	SubTypeGcpCloudRunService       AssetSubType = "gcp-cloud-run-service"
	SubTypeGcpIamRole               AssetSubType = "gcp-iam-role"
)

const (
	AwsCloudProvider   = "aws"
	AzureCloudProvider = "azure"
	GcpCloudProvider   = "gcp"
)

// AssetClassification holds the taxonomy of an asset
type AssetClassification struct {
	Category    AssetCategory    `json:"category"`
	SubCategory AssetSubCategory `json:"sub_category"`
	Type        AssetType        `json:"type"`
	SubType     AssetSubType     `json:"sub_type"`
}

// AssetClassifications below are used to generate
// 'internal/inventory/ASSETS.md'. Please keep formatting consistent.
var (
	// AWS
	AssetClassificationAwsEc2Instance              = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryCompute, Type: TypeVirtualMachine, SubType: SubTypeEC2}
	AssetClassificationAwsElbV1                    = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeLoadBalancer, SubType: SubTypeELBv1}
	AssetClassificationAwsElbV2                    = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeLoadBalancer, SubType: SubTypeELBv2}
	AssetClassificationAwsIamPolicy                = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryDigitalIdentity, Type: TypePolicy, SubType: SubTypeIAMPolicy}
	AssetClassificationAwsIamRole                  = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryDigitalIdentity, Type: TypeRole, SubType: SubTypeIAMRole}
	AssetClassificationAwsIamUser                  = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryDigitalIdentity, Type: TypeUser, SubType: SubTypeIAMUser}
	AssetClassificationAwsLambdaEventSourceMapping = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryIntegration, Type: TypeEventSource, SubType: SubTypeLambdaEventSourceMapping}
	AssetClassificationAwsLambdaFunction           = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryCompute, Type: TypeServerless, SubType: SubTypeLambdaFunction}
	AssetClassificationAwsLambdaLayer              = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryCompute, Type: TypeServerless, SubType: SubTypeLambdaLayer}
	AssetClassificationAwsInternetGateway          = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeGateway, SubType: SubTypeInternetGateway}
	AssetClassificationAwsNatGateway               = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeGateway, SubType: SubTypeNatGateway}
	AssetClassificationAwsNetworkAcl               = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryAuthorization, Type: TypeAcl, SubType: SubTypeVpcAcl}
	AssetClassificationAwsNetworkInterface         = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeInterface, SubType: SubTypeEC2NetworkInterface}
	AssetClassificationAwsSecurityGroup            = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeFirewall, SubType: SubTypeSecurityGroup}
	AssetClassificationAwsSubnet                   = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeSubnet, SubType: SubTypeEC2Subnet}
	AssetClassificationAwsTransitGateway           = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeVirtualNetwork, SubType: SubTypeTransitGateway}
	AssetClassificationAwsTransitGatewayAttachment = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeVirtualNetwork, SubType: SubTypeTransitGatewayAttachment}
	AssetClassificationAwsVpcPeeringConnection     = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypePeering, SubType: SubTypeVpcPeeringConnection}
	AssetClassificationAwsVpc                      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeVirtualNetwork, SubType: SubTypeVpc}
	AssetClassificationAwsRds                      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeRelationalDatabase, SubType: SubTypeRDS}
	AssetClassificationAwsS3Bucket                 = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeObjectStorage, SubType: SubTypeS3}
	AssetClassificationAwsSnsTopic                 = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryMessaging, Type: TypeNotificationService, SubType: SubTypeSNSTopic}
	// Azure
	AssetClassificationAzureAppService          = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryApplication, Type: TypeWebApplication, SubType: SubTypeAzureAppService}
	AssetClassificationAzureContainerRegistry   = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryContainer, Type: TypeRegistry, SubType: SubTypeAzureContainerRegistry}
	AssetClassificationAzureCosmosDBAccount     = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeNoSQLDatabase, SubType: SubTypeAzureCosmosDBAccount}
	AssetClassificationAzureCosmosDBSQLDatabase = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeNoSQLDatabase, SubType: SubTypeAzureCosmosDBSQLDatabase}
	AssetClassificationAzureDisk                = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeDisk, SubType: SubTypeAzureDisk}
	AssetClassificationAzureElasticPool         = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeScalability, SubType: SubTypeAzureElasticPool}
	AssetClassificationAzureResourceGroup       = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeResourceGroup, SubType: SubTypeAzureResourceGroup}
	AssetClassificationAzureSQLDatabase         = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeRelationalDatabase, SubType: SubTypeAzureSQLDatabase}
	AssetClassificationAzureSQLServer           = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryDatabase, Type: TypeRelationalDatabase, SubType: SubTypeAzureSQLServer}
	AssetClassificationAzureServicePrincipal    = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryDigitalIdentity, Type: TypePrincipal, SubType: SubTypeAzurePrincipal}
	AssetClassificationAzureSnapshot            = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeSnapshot, SubType: SubTypeAzureSnapshot}
	AssetClassificationAzureStorageAccount      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeStorage, SubType: SubTypeAzureStorageAccount}
	AssetClassificationAzureStorageBlobService  = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeObjectStorage, SubType: SubTypeAzureStorageBlobService}
	AssetClassificationAzureStorageQueue        = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryApplicationIntegration, Type: TypeMessageQueue, SubType: SubTypeAzureStorageQueue}
	AssetClassificationAzureStorageQueueService = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryApplicationIntegration, Type: TypeMessageQueue, SubType: SubTypeAzureStorageQueueService}
	AssetClassificationAzureSubscription        = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeCloudAccount, SubType: SubTypeAzureSubscription}
	AssetClassificationAzureTenant              = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeCloudAccount, SubType: SubTypeAzureTenant}
	AssetClassificationAzureVirtualMachine      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryCompute, Type: TypeVirtualMachine, SubType: SubTypeAzureVirtualMachine}

	// GCP
	AssetClassificationGcpProject           = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeCloudAccount, SubType: SubTypeGcpProject}
	AssetClassificationGcpOrganization      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeCloudAccount, SubType: SubTypeGcpOrganization}
	AssetClassificationGcpFolder            = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryManagement, Type: TypeResourceHierarchy, SubType: SubTypeGcpFolder}
	AssetClassificationGcpInstance          = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryCompute, Type: TypeVirtualMachine, SubType: SubTypeGcpInstance}
	AssetClassificationGcpBucket            = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryStorage, Type: TypeObjectStorage, SubType: SubTypeGcpBucket}
	AssetClassificationGcpFirewall          = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeFirewall, SubType: SubTypeGcpFirewall}
	AssetClassificationGcpSubnet            = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeSubnet, SubType: SubTypeGcpSubnet}
	AssetClassificationGcpServiceAccount    = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryServiceIdentity, Type: TypeServiceAccount, SubType: SubTypeGcpServiceAccount}
	AssetClassificationGcpServiceAccountKey = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryServiceIdentity, Type: TypeServiceAccountKey, SubType: SubTypeGcpServiceAccountKey}

	AssetClassificationGcpGkeCluster      = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryContainer, Type: TypeOrchestration, SubType: SubTypeGcpGkeCluster}
	AssetClassificationGcpForwardingRule  = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryNetwork, Type: TypeLoadBalancing, SubType: SubTypeGcpForwardingRule}
	AssetClassificationGcpIamRole         = AssetClassification{Category: CategoryIdentity, SubCategory: SubCategoryAccessManagement, Type: TypeIamRole, SubType: SubTypeGcpIamRole}
	AssetClassificationGcpCloudFunction   = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryServerless, Type: TypeFunction, SubType: SubTypeGcpCloudFunction}
	AssetClassificationGcpCloudRunService = AssetClassification{Category: CategoryInfrastructure, SubCategory: SubCategoryContainer, Type: TypeServerless, SubType: SubTypeGcpCloudRunService}
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

// Asset contains the identifiers of the asset
type Asset struct {
	Id              []string `json:"id"`
	RelatedEntityId []string `json:"related_entity_id"`
	Name            string   `json:"name"`
	AssetClassification
	Tags map[string]string `json:"tags"`
	Raw  any               `json:"raw"`
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
	AvailabilityZone *string                `json:"availability_zone,omitempty"`
	Provider         string                 `json:"provider,omitempty"`
	Region           string                 `json:"region,omitempty"`
	Account          AssetCloudAccount      `json:"account"`
	Organization     AssetCloudOrganization `json:"organization,omitempty"`
	Instance         *AssetCloudInstance    `json:"instance,omitempty"`
	Machine          *AssetCloudMachine     `json:"machine,omitempty"`
	Project          *AssetCloudProject     `json:"project,omitempty"`
	Service          *AssetCloudService     `json:"service,omitempty"`
}

type AssetCloudAccount struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type AssetCloudOrganization struct {
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

func NewAssetEvent(c AssetClassification, ids []string, name string, enrichers ...AssetEnricher) AssetEvent {
	a := AssetEvent{
		Asset: Asset{
			Id:                  removeEmpty(ids),
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

func WithRelatedAssetIds(ids []string) AssetEnricher {
	return func(a *AssetEvent) {
		a.Asset.RelatedEntityId = ids
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
