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
	"github.com/elastic/beats/v7/libbeat/ecs"
)

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
	Category AssetCategory `json:"category"`
	Type     AssetType     `json:"type"`
}

// AssetClassifications below are used to generate
// 'internal/inventory/ASSETS.md'. Please keep formatting consistent.
var (
	// AWS
	AssetClassificationAwsEc2Instance              = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryCompute */, Type: TypeVirtualMachine /* , SubType: SubTypeEC2 */}
	AssetClassificationAwsElbV1                    = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeLoadBalancer /* , SubType: SubTypeELBv1 */}
	AssetClassificationAwsElbV2                    = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeLoadBalancer /* , SubType: SubTypeELBv2 */}
	AssetClassificationAwsIamPolicy                = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryDigitalIdentity */, Type: TypePolicy /* , SubType: SubTypeIAMPolicy */}
	AssetClassificationAwsIamRole                  = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryDigitalIdentity */, Type: TypeRole /* , SubType: SubTypeIAMRole */}
	AssetClassificationAwsIamUser                  = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryDigitalIdentity */, Type: TypeUser /* , SubType: SubTypeIAMUser */}
	AssetClassificationAwsLambdaEventSourceMapping = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryIntegration */, Type: TypeEventSource /* , SubType: SubTypeLambdaEventSourceMapping */}
	AssetClassificationAwsLambdaFunction           = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryCompute */, Type: TypeServerless /* , SubType: SubTypeLambdaFunction */}
	AssetClassificationAwsLambdaLayer              = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryCompute */, Type: TypeServerless /* , SubType: SubTypeLambdaLayer */}
	AssetClassificationAwsInternetGateway          = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeGateway /* , SubType: SubTypeInternetGateway */}
	AssetClassificationAwsNatGateway               = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeGateway /* , SubType: SubTypeNatGateway */}
	AssetClassificationAwsNetworkAcl               = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryAuthorization */, Type: TypeAcl /* , SubType: SubTypeVpcAcl */}
	AssetClassificationAwsNetworkInterface         = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeInterface /* , SubType: SubTypeEC2NetworkInterface */}
	AssetClassificationAwsSecurityGroup            = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeFirewall /* , SubType: SubTypeSecurityGroup */}
	AssetClassificationAwsSubnet                   = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeSubnet /* , SubType: SubTypeEC2Subnet */}
	AssetClassificationAwsTransitGateway           = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeVirtualNetwork /* , SubType: SubTypeTransitGateway */}
	AssetClassificationAwsTransitGatewayAttachment = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeVirtualNetwork /* , SubType: SubTypeTransitGatewayAttachment */}
	AssetClassificationAwsVpcPeeringConnection     = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypePeering /* , SubType: SubTypeVpcPeeringConnection */}
	AssetClassificationAwsVpc                      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeVirtualNetwork /* , SubType: SubTypeVpc */}
	AssetClassificationAwsRds                      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeRelationalDatabase /* , SubType: SubTypeRDS */}
	AssetClassificationAwsS3Bucket                 = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeObjectStorage /* , SubType: SubTypeS3 */}
	AssetClassificationAwsSnsTopic                 = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryMessaging */, Type: TypeNotificationService /* , SubType: SubTypeSNSTopic */}
	// Azure
	AssetClassificationAzureAppService          = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryApplication */, Type: TypeWebApplication /* , SubType: SubTypeAzureAppService */}
	AssetClassificationAzureContainerRegistry   = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryContainer */, Type: TypeRegistry /* , SubType: SubTypeAzureContainerRegistry */}
	AssetClassificationAzureCosmosDBAccount     = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeNoSQLDatabase /* , SubType: SubTypeAzureCosmosDBAccount */}
	AssetClassificationAzureCosmosDBSQLDatabase = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeNoSQLDatabase /* , SubType: SubTypeAzureCosmosDBSQLDatabase */}
	AssetClassificationAzureDisk                = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeDisk /* , SubType: SubTypeAzureDisk */}
	AssetClassificationAzureElasticPool         = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeScalability /* , SubType: SubTypeAzureElasticPool */}
	AssetClassificationAzureResourceGroup       = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeResourceGroup /* , SubType: SubTypeAzureResourceGroup */}
	AssetClassificationAzureSQLDatabase         = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeRelationalDatabase /* , SubType: SubTypeAzureSQLDatabase */}
	AssetClassificationAzureSQLServer           = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryDatabase */, Type: TypeRelationalDatabase /* , SubType: SubTypeAzureSQLServer */}
	AssetClassificationAzureServicePrincipal    = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryDigitalIdentity */, Type: TypePrincipal /* , SubType: SubTypeAzurePrincipal */}
	AssetClassificationAzureSnapshot            = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeSnapshot /* , SubType: SubTypeAzureSnapshot */}
	AssetClassificationAzureStorageAccount      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeStorage /* , SubType: SubTypeAzureStorageAccount */}
	AssetClassificationAzureStorageBlobService  = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeObjectStorage /* , SubType: SubTypeAzureStorageBlobService */}
	AssetClassificationAzureStorageQueue        = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryApplicationIntegration */, Type: TypeMessageQueue /* , SubType: SubTypeAzureStorageQueue */}
	AssetClassificationAzureStorageQueueService = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryApplicationIntegration */, Type: TypeMessageQueue /* , SubType: SubTypeAzureStorageQueueService */}
	AssetClassificationAzureSubscription        = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeCloudAccount /* , SubType: SubTypeAzureSubscription */}
	AssetClassificationAzureTenant              = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeCloudAccount /* , SubType: SubTypeAzureTenant */}
	AssetClassificationAzureVirtualMachine      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryCompute */, Type: TypeVirtualMachine /* , SubType: SubTypeAzureVirtualMachine */}

	// GCP
	AssetClassificationGcpProject           = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeCloudAccount /* , SubType: SubTypeGcpProject */}
	AssetClassificationGcpOrganization      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeCloudAccount /* , SubType: SubTypeGcpOrganization */}
	AssetClassificationGcpFolder            = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryManagement */, Type: TypeResourceHierarchy /* , SubType: SubTypeGcpFolder */}
	AssetClassificationGcpInstance          = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryCompute */, Type: TypeVirtualMachine /* , SubType: SubTypeGcpInstance */}
	AssetClassificationGcpBucket            = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryStorage */, Type: TypeObjectStorage /* , SubType: SubTypeGcpBucket */}
	AssetClassificationGcpFirewall          = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeFirewall /* , SubType: SubTypeGcpFirewall */}
	AssetClassificationGcpSubnet            = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeSubnet /* , SubType: SubTypeGcpSubnet */}
	AssetClassificationGcpServiceAccount    = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryServiceIdentity */, Type: TypeServiceAccount /* , SubType: SubTypeGcpServiceAccount */}
	AssetClassificationGcpServiceAccountKey = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryServiceIdentity */, Type: TypeServiceAccountKey /* , SubType: SubTypeGcpServiceAccountKey */}

	AssetClassificationGcpGkeCluster      = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryContainer */, Type: TypeOrchestration /* , SubType: SubTypeGcpGkeCluster */}
	AssetClassificationGcpForwardingRule  = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryNetwork */, Type: TypeLoadBalancing /* , SubType: SubTypeGcpForwardingRule */}
	AssetClassificationGcpIamRole         = AssetClassification{Category: CategoryIdentity /* , SubCategory: SubCategoryAccessManagement */, Type: TypeIamRole /* , SubType: SubTypeGcpIamRole */}
	AssetClassificationGcpCloudFunction   = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryServerless */, Type: TypeFunction /* , SubType: SubTypeGcpCloudFunction */}
	AssetClassificationGcpCloudRunService = AssetClassification{Category: CategoryInfrastructure /* , SubCategory: SubCategoryContainer */, Type: TypeServerless /* , SubType: SubTypeGcpCloudRunService */}
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Entity        Entity
	Network       *ecs.Network
	Cloud         *ecs.Cloud
	Host          *ecs.Host
	User          *ecs.User
	Labels        map[string]string
	RawAttributes *any
}

// Entity contains the identifiers of the asset
type Entity struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	AssetClassification

	// non exported fields
	relatedEntityId []string
}

// AssetEnricher functional builder function
type AssetEnricher func(asset *AssetEvent)

func NewAssetEvent(c AssetClassification, id string, name string, enrichers ...AssetEnricher) AssetEvent {
	a := AssetEvent{
		Entity: Entity{
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
		a.RawAttributes = &raw
	}
}

func WithRelatedAssetIds(ids []string) AssetEnricher {
	return func(a *AssetEvent) {
		a.Entity.relatedEntityId = ids
	}
}

func WithLabels(labels map[string]string) AssetEnricher {
	return func(a *AssetEvent) {
		if len(labels) == 0 {
			return
		}

		a.Labels = labels
	}
}

func WithNetwork(network ecs.Network) AssetEnricher {
	return func(a *AssetEvent) {
		a.Network = &network
	}
}

func WithCloud(cloud ecs.Cloud) AssetEnricher {
	return func(a *AssetEvent) {
		a.Cloud = &cloud
	}
}

func WithHost(host ecs.Host) AssetEnricher {
	return func(a *AssetEvent) {
		a.Host = &host
	}
}

func WithUser(user ecs.User) AssetEnricher {
	return func(a *AssetEvent) {
		a.User = &user
	}
}

func EmptyEnricher() AssetEnricher {
	return func(_ *AssetEvent) {}
}
