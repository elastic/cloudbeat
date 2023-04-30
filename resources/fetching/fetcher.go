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
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	KubeAPIType = "kube-api"

	EcrType                   = "aws-ecr"
	IAMType                   = "aws-iam"
	EC2Type                   = "aws-ec2"
	EC2NetworkingType         = "aws-ec2-network"
	MonitoringType            = "aws-monitoring"
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
	EksType                   = "aws-eks"
	S3Type                    = "aws-s3"
	KmsType                   = "aws-kms"
	SecurityHubType           = "aws-securityhub"
	VpcType                   = "aws-vpc"
	RdsType                   = "aws-rds"
	ConfigServiceResourceType = "aws-config"
	PolicyType                = "aws-policy"
	AccessAnalyzers           = "aws-access-analyzers"

	CloudIdentity          = "identity-management"
	EC2Identity            = "cloud-compute"
	MonitoringIdentity     = "monitoring"
	CloudContainerMgmt     = "caas" // containers as a service
	CloudLoadBalancer      = "load-balancer"
	CloudContainerRegistry = "container-registry"
	CloudStorage           = "cloud-storage"
	CloudAudit             = "cloud-audit"
	CloudDatabase          = "cloud-database"
	CloudConfig            = "cloud-config"
	KeyManagement          = "key-management"
)

// Factory can create fetcher instances based on configuration
type Factory interface {
	Create(*logp.Logger, *config.C, chan ResourceInfo) (Fetcher, error)
}

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

type Resource interface {
	GetMetadata() (ResourceMetadata, error)
	GetData() any
	GetElasticCommonData() any
}

type ResourceFields struct {
	ResourceMetadata
	Raw interface{} `json:"raw"`
}

type ResourceMetadata struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	SubType   string `json:"sub_type,omitempty"`
	Name      string `json:"name,omitempty"`
	ECSFormat string `json:"ecsFormat,omitempty"`
	Region    string `json:"region,omitempty"`
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
