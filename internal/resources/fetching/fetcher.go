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

package fetching

import (
	"context"

	awssdk "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
)

const (
	FileSystemType = "file-system"
	KubeAPIType    = "kube-api"
	ProcessType    = "process"

	// AWS subtypes
	AccessAnalyzers              = "aws-access-analyzers"
	AwsMonitoringType            = "aws-monitoring"
	ConfigServiceResourceType    = "aws-config"
	EBSSnapshotType              = "aws-ebs-snapshot"
	EBSType                      = "aws-ebs"
	EC2NetworkingType            = "aws-ec2-network"
	EC2Type                      = "aws-ec2"
	EcrType                      = "aws-ecr"
	ElbType                      = "aws-elb"
	IAMServerCertificateType     = "aws-iam-server-certificate"
	IAMType                      = "aws-iam"
	IAMUserType                  = "aws-iam-user"
	InternetGateway              = "aws-internet-gateway"
	KmsType                      = "aws-kms"
	LambdaAliasType              = "aws-lambda-function-alias"
	LambdaEventSourceMappingType = "aws-lambda-event-source-mapping"
	LambdaFunctionType           = "aws-lambda-function"
	LambdaLayerType              = "aws-lambda-layer"
	MultiTrailsType              = "aws-multi-trails"
	NatGateway                   = "aws-nat-gateway"
	NetworkInterface             = "aws-network-interface"
	NetworkNACLType              = "aws-nacl"
	PolicyType                   = "aws-policy"
	PwdPolicyType                = "aws-password-policy"
	RdsType                      = "aws-rds"
	S3Type                       = "aws-s3"
	SNSTopicType                 = "aws-sns"
	SecurityGroupType            = "aws-security-group"
	SecurityHubType              = "aws-securityhub"
	Subnet                       = "aws-subnet"
	TrailType                    = "aws-trail"
	TransitGateway               = "aws-transit-gateway"
	TransitGatewayAttachment     = "aws-transit-gateway-attachment"
	VpcPeeringConnectionType     = "aws-vpc-peering-connection"
	VpcType                      = "aws-vpc"

	// GCP subtypes
	GcpLoggingType    = "gcp-logging"
	GcpMonitoringType = "gcp-monitoring"
	GcpPolicies       = "gcp-policies"
	GcpServiceUsage   = "gcp-service-usage"

	// Azure resources group subtypes
	AzureActivityLogAlertType          = "azure-activity-log-alert"
	AzureBastionType                   = "azure-bastion"
	AzureClassicStorageAccountType     = "azure-classic-storage-account"
	AzureDiagnosticSettingsType        = "azure-diagnostic-settings"
	AzureDiskType                      = "azure-disk"
	AzureDocumentDBDatabaseAccountType = "azure-document-db-database-account"
	AzureInsightsComponentType         = "azure-insights-component"
	AzureMySQLDBType                   = "azure-mysql-server-db"
	AzureFlexibleMySQLDBType           = "azure-flexible-mysql-server-db"
	AzureNetworkWatchersFlowLogType    = "azure-network-watchers-flow-log"
	AzureNetworkWatchersType           = "azure-network-watcher"
	AzureNetworkSecurityGroupType      = "azure-network-group"
	AzurePostgreSQLDBType              = "azure-postgresql-server-db"
	AzureFlexiblePostgreSQLDBType      = "azure-flexible-postgresql-server-db"
	AzureSecurityContactsType          = "azure-security-contacts"
	AzureAutoProvisioningSettingsType  = "azure-security-auto-provisioning-settings"
	AzureSQLServerType                 = "azure-sql-server"
	AzureStorageAccountType            = "azure-storage-account"
	AzureVMType                        = "azure-vm"
	AzureVaultType                     = "azure-vault"
	AzureWebSiteType                   = "azure-web-site"

	// Azure authorizationresources group subtypes
	AzureRoleDefinitionType = "azure-role-definition"

	// Types
	CloudAudit             = "cloud-audit"
	CloudCompute           = "cloud-compute"
	CloudConfig            = "cloud-config"
	CloudContainerMgmt     = "caas" // containers as a service
	CloudContainerRegistry = "container-registry"
	CloudDatabase          = "cloud-database"
	CloudDns               = "cloud-dns"
	CloudIdentity          = "identity-management"
	CloudLoadBalancer      = "load-balancer"
	CloudStorage           = "cloud-storage"
	DataProcessing         = "data-processing"
	KeyManagement          = "key-management"
	LoggingIdentity        = "logging"
	MonitoringIdentity     = "monitoring"
	ProjectManagement      = "project-management"
)

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context, cycle.Metadata) error
	Stop()
}

type Condition interface {
	Condition() bool
	Name() string
}

type ResourceInfo struct {
	Resource
	CycleMetadata cycle.Metadata
}

type Resource interface {
	GetMetadata() (ResourceMetadata, error)
	GetData() any
	GetElasticCommonData() (map[string]any, error)
	GetIds() []string
}

type ResourceFields struct {
	ResourceMetadata
	Raw any `json:"raw,omitempty"`
}

type ResourceMetadata struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"sub_type,omitempty"`
	Name    string `json:"name,omitempty"`
	Region  string `json:"region,omitempty"`

	CloudAccountMetadata
}

type CloudAccountMetadata struct {
	AccountId        string `json:"account_id,omitempty"`
	AccountName      string `json:"account_name,omitempty"`
	OrganisationId   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
}

type Result struct {
	Type     string `json:"type"`
	SubType  string `json:"subType"`
	Resource any    `json:"resource"`
}

type ResourceMap map[string][]Resource

type BaseFetcherConfig struct {
	Name string `config:"name"`
}

type AwsBaseFetcherConfig struct {
	BaseFetcherConfig `config:",inline"`
	AwsConfig         awssdk.ConfigAWS `config:",inline"`
}
