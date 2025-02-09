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
	"github.com/elastic/cloudbeat/internal/ecs"
	"github.com/samber/lo"
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
	CategoryFileSystemService      AssetCategory = "File System Service"
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
	AssetClassificationAwsEc2Instance              = AssetClassification{CategoryHost, "AWS EC2 Instance"}
	AssetClassificationAwsElbV1                    = AssetClassification{CategoryLoadBalancer, "AWS Elastic Load Balancer"}
	AssetClassificationAwsElbV2                    = AssetClassification{CategoryLoadBalancer, "AWS Elastic Load Balancer v2"}
	AssetClassificationAwsIamPolicy                = AssetClassification{CategoryAccessManagement, "AWS IAM Policy"}
	AssetClassificationAwsIamRole                  = AssetClassification{CategoryServiceAccount, "AWS IAM Role"}
	AssetClassificationAwsIamUser                  = AssetClassification{CategoryIdentity, "AWS IAM User"}
	AssetClassificationAwsLambdaEventSourceMapping = AssetClassification{CategoryFaaS, "AWS Lambda Event Source Mapping"}
	AssetClassificationAwsLambdaFunction           = AssetClassification{CategoryFaaS, "AWS Lambda Function"}
	AssetClassificationAwsLambdaLayer              = AssetClassification{CategoryFaaS, "AWS Lambda Layer"}
	AssetClassificationAwsInternetGateway          = AssetClassification{CategoryGateway, "AWS Internet Gateway"}
	AssetClassificationAwsNatGateway               = AssetClassification{CategoryGateway, "AWS NAT Gateway"}
	AssetClassificationAwsNetworkAcl               = AssetClassification{CategoryNetworking, "AWS EC2 Network ACL"}
	AssetClassificationAwsNetworkInterface         = AssetClassification{CategoryNetworking, "AWS EC2 Network Interface"}
	AssetClassificationAwsSecurityGroup            = AssetClassification{CategoryFirewall, "AWS EC2 Security Group"}
	AssetClassificationAwsSubnet                   = AssetClassification{CategoryNetworking, "AWS EC2 Subnet"}
	AssetClassificationAwsTransitGateway           = AssetClassification{CategoryGateway, "AWS Transit Gateway"}
	AssetClassificationAwsTransitGatewayAttachment = AssetClassification{CategoryGateway, "AWS Transit Gateway Attachment"}
	AssetClassificationAwsVpcPeeringConnection     = AssetClassification{CategoryNetworking, "AWS VPC Peering Connection"}
	AssetClassificationAwsVpc                      = AssetClassification{CategoryNetworking, "AWS VPC"}
	AssetClassificationAwsRds                      = AssetClassification{CategoryDatabase, "AWS RDS Instance"}
	AssetClassificationAwsS3Bucket                 = AssetClassification{CategoryStorageBucket, "AWS S3 Bucket"}
	AssetClassificationAwsSnsTopic                 = AssetClassification{CategoryMessagingService, "AWS SNS Topic"}

	// Azure
	AssetClassificationAzureAppService           = AssetClassification{CategoryWebService, "Azure App Service"}
	AssetClassificationAzureContainerRegistry    = AssetClassification{CategoryContainerRegistry, "Azure Container Registry"}
	AssetClassificationAzureCosmosDBAccount      = AssetClassification{CategoryInfrastructure, "Azure Cosmos DB Account"}
	AssetClassificationAzureCosmosDBSQLDatabase  = AssetClassification{CategoryInfrastructure, "Azure Cosmos DB SQL Database"}
	AssetClassificationAzureDisk                 = AssetClassification{CategoryVolume, "Azure Disk"}
	AssetClassificationAzureElasticPool          = AssetClassification{CategoryDatabase, "Azure Elastic Pool"}
	AssetClassificationAzureResourceGroup        = AssetClassification{CategoryAccessManagement, "Azure Resource Group"}
	AssetClassificationAzureSQLDatabase          = AssetClassification{CategoryDatabase, "Azure SQL Database"}
	AssetClassificationAzureSQLServer            = AssetClassification{CategoryDatabase, "Azure SQL Server"}
	AssetClassificationAzureServicePrincipal     = AssetClassification{CategoryIdentity, "Azure Principal"}
	AssetClassificationAzureSnapshot             = AssetClassification{CategorySnapshot, "Azure Snapshot"}
	AssetClassificationAzureStorageAccount       = AssetClassification{CategoryPrivateEndpoint, "Azure Storage Account"}
	AssetClassificationAzureStorageBlobContainer = AssetClassification{CategoryStorageBucket, "Azure Storage Blob Container"}
	AssetClassificationAzureStorageBlobService   = AssetClassification{CategoryServiceUsageTechnology, "Azure Storage Blob Service"}
	AssetClassificationAzureStorageFileService   = AssetClassification{CategoryFileSystemService, "Azure Storage File Service"}
	AssetClassificationAzureStorageFileShare     = AssetClassification{CategoryFileSystemService, "Azure Storage File Share"}
	AssetClassificationAzureStorageQueue         = AssetClassification{CategoryMessagingService, "Azure Storage Queue"}
	AssetClassificationAzureStorageQueueService  = AssetClassification{CategoryMessagingService, "Azure Storage Queue Service"}
	AssetClassificationAzureStorageTable         = AssetClassification{CategoryDatabase, "Azure Storage Table"}
	AssetClassificationAzureStorageTableService  = AssetClassification{CategoryServiceUsageTechnology, "Azure Storage Table Service"}
	AssetClassificationAzureSubscription         = AssetClassification{CategoryAccessManagement, "Azure Subscription"}
	AssetClassificationAzureTenant               = AssetClassification{CategoryAccessManagement, "Azure Tenant"}
	AssetClassificationAzureVirtualMachine       = AssetClassification{CategoryHost, "Azure Virtual Machine"}

	// GCP
	AssetClassificationGcpProject           = AssetClassification{CategoryAccount, "GCP Project"}
	AssetClassificationGcpOrganization      = AssetClassification{CategoryOrganization, "GCP Organization"}
	AssetClassificationGcpFolder            = AssetClassification{CategoryOrganization, "GCP Folder"}
	AssetClassificationGcpInstance          = AssetClassification{CategoryHost, "GCP Compute Instance"}
	AssetClassificationGcpBucket            = AssetClassification{CategoryStorageBucket, "GCP Bucket"}
	AssetClassificationGcpFirewall          = AssetClassification{CategoryFirewall, "GCP Firewall"}
	AssetClassificationGcpSubnet            = AssetClassification{CategorySubnet, "GCP Subnet"}
	AssetClassificationGcpServiceAccount    = AssetClassification{CategoryAccessManagement, "GCP Service Account"}
	AssetClassificationGcpServiceAccountKey = AssetClassification{CategoryAccessManagement, "GCP Service Account Key"}
	AssetClassificationGcpGkeCluster        = AssetClassification{CategoryOrchestrator, "GCP Kubernetes Engine (GKE) Cluster"}
	AssetClassificationGcpForwardingRule    = AssetClassification{CategoryLoadBalancer, "GCP Load Balancing Forwarding Rule"}
	AssetClassificationGcpIamRole           = AssetClassification{CategoryServiceUsageTechnology, "GCP IAM Role"}
	AssetClassificationGcpCloudFunction     = AssetClassification{CategoryFaaS, "GCP Cloud Function"}
	AssetClassificationGcpCloudRunService   = AssetClassification{CategoryContainerService, "GCP Cloud Run Service"}
)

// AssetEvent holds the whole asset
type AssetEvent struct {
	Entity        Entity
	Event         *ecs.Event
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
		Event: &ecs.Event{
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
		ids = lo.Filter(ids, func(id string, _ int) bool {
			return id != ""
		})

		if len(ids) == 0 {
			a.Entity.relatedEntityId = nil
			return
		}

		a.Entity.relatedEntityId = lo.Uniq(ids)
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
