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
)

const (
	KubeAPIType    = "kube-api"
	FileSystemType = "file-system"
	ProcessType    = "process"

	EcrType                   = "aws-ecr"
	IAMType                   = "aws-iam"
	EC2Type                   = "aws-ec2"
	EC2NetworkingType         = "aws-ec2-network"
	AwsMonitoringType         = "aws-monitoring"
	NetworkNACLType           = "aws-nacl"
	TrailType                 = "aws-trail"
	MultiTrailsType           = "aws-multi-trails"
	SecurityGroupType         = "aws-security-group"
	EBSType                   = "aws-ebs"
	EBSSnapshotType           = "aws-ebs-snapshot"
	ElbType                   = "aws-elb"
	IAMUserType               = "aws-iam-user"
	IAMServerCertificateType  = "aws-iam-server-certificate"
	PwdPolicyType             = "aws-password-policy"
	S3Type                    = "aws-s3"
	KmsType                   = "aws-kms"
	SecurityHubType           = "aws-securityhub"
	VpcType                   = "aws-vpc"
	RdsType                   = "aws-rds"
	ConfigServiceResourceType = "aws-config"
	PolicyType                = "aws-policy"
	AccessAnalyzers           = "aws-access-analyzers"

	GcpMonitoringType      = "gcp-monitoring"
	GcpLoggingType         = "gcp-logging"
	GcpServiceUsage        = "gcp-service-usage"
	GcpPolicies            = "gcp-policies"
	CloudIdentity          = "identity-management"
	CloudCompute           = "cloud-compute"
	MonitoringIdentity     = "monitoring"
	LoggingIdentity        = "logging"
	CloudContainerMgmt     = "caas" // containers as a service
	CloudLoadBalancer      = "load-balancer"
	CloudContainerRegistry = "container-registry"
	CloudStorage           = "cloud-storage"
	CloudAudit             = "cloud-audit"
	CloudDatabase          = "cloud-database"
	CloudConfig            = "cloud-config"
	CloudDns               = "cloud-dns"
	KeyManagement          = "key-management"
	ProjectManagement      = "project-management"
	DataProcessing         = "data-processing"

	AzureActivityLogAlertType          = "azure-activity-log-alert"
	AzureBastionType                   = "azure-bastion"
	AzureClassicStorageAccountType     = "azure-classic-storage-account"
	AzureClassicVMType                 = "azure-classic-vm"
	AzureDiskType                      = "azure-disk"
	AzureDocumentDBDatabaseAccountType = "azure-document-db-database-account"
	AzureMySQLDBType                   = "azure-mysql-server-db"
	AzureNetworkWatchersFlowLogType    = "azure-network-watchers-flow-log"
	AzureNetworkWatchersType           = "azure-network-watcher"
	AzurePostgreSQLDBType              = "azure-postgresql-server-db"
	AzureSQLServerType                 = "azure-sql-server"
	AzureStorageAccountType            = "azure-storage-account"
	AzureVMType                        = "azure-vm"
	AzureWebSiteType                   = "azure-web-site"
)

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context, CycleMetadata) error
	Stop()
}

type Condition interface {
	Condition() bool
	Name() string
}

type ResourceInfo struct {
	Resource
	CycleMetadata
}

type CycleMetadata struct {
	Sequence int64
}

type EcsGcp struct {
	Provider         string
	ProjectId        string
	ProjectName      string
	OrganizationId   string
	OrganizationName string
}

type Resource interface {
	GetMetadata() (ResourceMetadata, error)
	GetData() any
	GetElasticCommonData() (map[string]any, error)
}

type ResourceFields struct {
	ResourceMetadata
	Raw interface{} `json:"raw"`
}

type ResourceMetadata struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	SubType             string `json:"sub_type,omitempty"`
	Name                string `json:"name,omitempty"`
	Region              string `json:"region,omitempty"`
	AwsAccountId        string `json:"aws_account_id,omitempty"`
	AwsAccountAlias     string `json:"aws_account_alias,omitempty"`
	AwsOrganizationId   string `json:"aws_organization_id,omitempty"`
	AwsOrganizationName string `json:"aws_organization_name,omitempty"`
}

type Result struct {
	Type     string      `json:"type"`
	SubType  string      `json:"subType"`
	Resource interface{} `json:"resource"`
}

type ResourceMap map[string][]Resource

type BaseFetcherConfig struct {
	Name string `config:"name"`
}

type AwsBaseFetcherConfig struct {
	BaseFetcherConfig `config:",inline"`
	AwsConfig         awssdk.ConfigAWS `config:",inline"`
}
