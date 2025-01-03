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

// AssetCategory is used to build the document index.
type AssetCategory string

const (
	CategoryAccessManagement       AssetCategory = "Access Management"
	CategoryAccount                AssetCategory = "Account"
	CategoryContainerRegistry      AssetCategory = "Container Registry"
	CategoryContainerService       AssetCategory = "Container Service"
	CategoryDatabase               AssetCategory = "Database"
	CategoryFaaS                   AssetCategory = "FaaS"
	CategoryFirewall               AssetCategory = "Firewall"
	CategoryGateway                AssetCategory = "Gateway"
	CategoryHost                   AssetCategory = "Host"
	CategoryIdentity               AssetCategory = "Identity"
	CategoryInfrastructure         AssetCategory = "Infrastructure"
	CategoryLoadBalancer           AssetCategory = "Load Balancer"
	CategoryMessagingService       AssetCategory = "Messaging Service"
	CategoryNetworking             AssetCategory = "Networking"
	CategoryOrchestrator           AssetCategory = "Orchestrator"
	CategoryOrganization           AssetCategory = "Organization"
	CategoryPrivateEndpoint        AssetCategory = "Private Endpoint"
	CategoryServiceAccount         AssetCategory = "Service Account"
	CategoryServiceUsageTechnology AssetCategory = "Service Usage Technology"
	CategorySnapshot               AssetCategory = "Snapshot"
	CategoryStorageBucket          AssetCategory = "Storage Bucket"
	CategorySubnet                 AssetCategory = "Subnet"
	CategoryVolume                 AssetCategory = "Volume"
	CategoryWebService             AssetCategory = "Web Service"
)

// AssetType is used to build the document index.
type AssetType string

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
	AssetClassificationAwsEc2Instance              = AssetClassification{Category: CategoryHost, Type: "AWS EC2 Instance"}
	AssetClassificationAwsElbV1                    = AssetClassification{Category: CategoryLoadBalancer, Type: "AWS Elastic Load Balancer"}
	AssetClassificationAwsElbV2                    = AssetClassification{Category: CategoryLoadBalancer, Type: "AWS Elastic Load Balancer v2"}
	AssetClassificationAwsIamPolicy                = AssetClassification{Category: CategoryAccessManagement, Type: "AWS IAM Policy"}
	AssetClassificationAwsIamRole                  = AssetClassification{Category: CategoryServiceAccount, Type: "AWS IAM Role"}
	AssetClassificationAwsIamUser                  = AssetClassification{Category: CategoryIdentity, Type: "AWS IAM User"}
	AssetClassificationAwsLambdaEventSourceMapping = AssetClassification{Category: CategoryFaaS, Type: "AWS Lambda Event Source Mapping"}
	AssetClassificationAwsLambdaFunction           = AssetClassification{Category: CategoryFaaS, Type: "AWS Lambda Function"}
	AssetClassificationAwsLambdaLayer              = AssetClassification{Category: CategoryFaaS, Type: "AWS Lambda Layer"}
	AssetClassificationAwsInternetGateway          = AssetClassification{Category: CategoryGateway, Type: "AWS Internet Gateway"}
	AssetClassificationAwsNatGateway               = AssetClassification{Category: CategoryGateway, Type: "AWS NAT Gateway"}
	AssetClassificationAwsNetworkAcl               = AssetClassification{Category: CategoryNetworking, Type: "AWS EC2 Network ACL"}
	AssetClassificationAwsNetworkInterface         = AssetClassification{Category: CategoryNetworking, Type: "AWS EC2 Network Interface"}
	AssetClassificationAwsSecurityGroup            = AssetClassification{Category: CategoryFirewall, Type: "AWS EC2 Security Group"}
	AssetClassificationAwsSubnet                   = AssetClassification{Category: CategoryNetworking, Type: "AWS EC2 Subnet"}
	AssetClassificationAwsTransitGateway           = AssetClassification{Category: CategoryGateway, Type: "AWS Transit Gateway"}
	AssetClassificationAwsTransitGatewayAttachment = AssetClassification{Category: CategoryGateway, Type: "AWS Transit Gateway Attachment"}
	AssetClassificationAwsVpcPeeringConnection     = AssetClassification{Category: CategoryNetworking, Type: "AWS VPC Peering Connection"}
	AssetClassificationAwsVpc                      = AssetClassification{Category: CategoryNetworking, Type: "AWS VPC"}
	AssetClassificationAwsRds                      = AssetClassification{Category: CategoryDatabase, Type: "AWS RDS Instance"}
	AssetClassificationAwsS3Bucket                 = AssetClassification{Category: CategoryStorageBucket, Type: "AWS S3 Bucket"}
	AssetClassificationAwsSnsTopic                 = AssetClassification{Category: CategoryMessagingService, Type: "AWS SNS Topic"}

	// Azure
	AssetClassificationAzureAppService          = AssetClassification{Category: CategoryWebService, Type: "Azure App Service"}
	AssetClassificationAzureContainerRegistry   = AssetClassification{Category: CategoryContainerRegistry, Type: "Azure Container Registry"}
	AssetClassificationAzureCosmosDBAccount     = AssetClassification{Category: CategoryInfrastructure, Type: "Azure Cosmos DB Account"}
	AssetClassificationAzureCosmosDBSQLDatabase = AssetClassification{Category: CategoryInfrastructure, Type: "Azure Cosmos DB SQL Database"}
	AssetClassificationAzureDisk                = AssetClassification{Category: CategoryVolume, Type: "Azure Disk"}
	AssetClassificationAzureElasticPool         = AssetClassification{Category: CategoryDatabase, Type: "Azure Elastic Pool"}
	AssetClassificationAzureResourceGroup       = AssetClassification{Category: CategoryAccessManagement, Type: "Azure Resource Group"}
	AssetClassificationAzureSQLDatabase         = AssetClassification{Category: CategoryDatabase, Type: "Azure SQL Database"}
	AssetClassificationAzureSQLServer           = AssetClassification{Category: CategoryDatabase, Type: "Azure SQL Server"}
	AssetClassificationAzureServicePrincipal    = AssetClassification{Category: CategoryIdentity, Type: "Azure Principal"}
	AssetClassificationAzureSnapshot            = AssetClassification{Category: CategorySnapshot, Type: "Azure Snapshot"}
	AssetClassificationAzureStorageAccount      = AssetClassification{Category: CategoryPrivateEndpoint, Type: "Azure Storage Account"}
	AssetClassificationAzureStorageBlobService  = AssetClassification{Category: CategoryStorageBucket, Type: "Azure Storage Blob Service"}
	AssetClassificationAzureStorageQueue        = AssetClassification{Category: CategoryMessagingService, Type: "Azure Storage Queue"}
	AssetClassificationAzureStorageQueueService = AssetClassification{Category: CategoryMessagingService, Type: "Azure Storage Queue Service"}
	AssetClassificationAzureSubscription        = AssetClassification{Category: CategoryAccessManagement, Type: "Azure Subscription"}
	AssetClassificationAzureTenant              = AssetClassification{Category: CategoryAccessManagement, Type: "Azure Tenant"}
	AssetClassificationAzureVirtualMachine      = AssetClassification{Category: CategoryHost, Type: "Azure Virtual Machine"}

	// GCP
	AssetClassificationGcpProject           = AssetClassification{Category: CategoryAccount, Type: "GCP Project"}
	AssetClassificationGcpOrganization      = AssetClassification{Category: CategoryOrganization, Type: "GCP Organization"}
	AssetClassificationGcpFolder            = AssetClassification{Category: CategoryOrganization, Type: "GCP Folder"}
	AssetClassificationGcpInstance          = AssetClassification{Category: CategoryHost, Type: "GCP Compute Instance"}
	AssetClassificationGcpBucket            = AssetClassification{Category: CategoryStorageBucket, Type: "GCP Bucket"}
	AssetClassificationGcpFirewall          = AssetClassification{Category: CategoryFirewall, Type: "GCP Firewall"}
	AssetClassificationGcpSubnet            = AssetClassification{Category: CategorySubnet, Type: "GCP Subnet"}
	AssetClassificationGcpServiceAccount    = AssetClassification{Category: CategoryAccessManagement, Type: "GCP Service Account"}
	AssetClassificationGcpServiceAccountKey = AssetClassification{Category: CategoryAccessManagement, Type: "GCP Service Account Key"}
	AssetClassificationGcpGkeCluster        = AssetClassification{Category: CategoryOrchestrator, Type: "GCP Kubernetes Engine (GKE) Cluster"}
	AssetClassificationGcpForwardingRule    = AssetClassification{Category: CategoryLoadBalancer, Type: "GCP Load Balancing Forwarding Rule"}
	AssetClassificationGcpIamRole           = AssetClassification{Category: CategoryServiceUsageTechnology, Type: "GCP IAM Role"}
	AssetClassificationGcpCloudFunction     = AssetClassification{Category: CategoryFaaS, Type: "GCP Cloud Function"}
	AssetClassificationGcpCloudRunService   = AssetClassification{Category: CategoryContainerService, Type: "GCP Cloud Run Service"}
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Entity        Entity
	Event         ecs.Event
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
		Event: ecs.Event{
			Kind: "asset",
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
