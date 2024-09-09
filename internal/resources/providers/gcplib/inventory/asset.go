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

const (
	StorageBucketAssetType         = "storage.googleapis.com/Bucket"
	ComputeFirewallAssetType       = "compute.googleapis.com/Firewall"
	ComputeInstanceAssetType       = "compute.googleapis.com/Instance"
	ComputeNetworkAssetType        = "compute.googleapis.com/Network"
	ComputeBackendServiceAssetType = "compute.googleapis.com/RegionBackendService"
	ComputeSubnetworkAssetType     = "compute.googleapis.com/Subnetwork"
	ComputeDiskAssetType           = "compute.googleapis.com/Disk"
	DnsManagedZoneAssetType        = "dns.googleapis.com/ManagedZone"
	BigqueryDatasetAssetType       = "bigquery.googleapis.com/Dataset"
	BigqueryTableAssetType         = "bigquery.googleapis.com/Table"
	CrmProjectAssetType            = "cloudresourcemanager.googleapis.com/Project"
	CrmOrgAssetType                = "cloudresourcemanager.googleapis.com/Organization"
	CrmFolderAssetType             = "cloudresourcemanager.googleapis.com/Folder"
	ApiKeysKeyAssetType            = "apikeys.googleapis.com/Key"
	CloudKmsCryptoKeyAssetType     = "cloudkms.googleapis.com/CryptoKey"
	IamServiceAccountAssetType     = "iam.googleapis.com/ServiceAccount"
	IamServiceAccountKeyAssetType  = "iam.googleapis.com/ServiceAccountKey"
	SqlDatabaseInstanceAssetType   = "sqladmin.googleapis.com/Instance"
	LogBucketAssetType             = "logging.googleapis.com/LogBucket"
	LogSinkAssetType               = "logging.googleapis.com/LogSink"
	DataprocClusterAssetType       = "dataproc.googleapis.com/Cluster"
	MonitoringLogMetricAssetType   = "logging.googleapis.com/LogMetric"
	MonitoringAlertPolicyAssetType = "monitoring.googleapis.com/AlertPolicy"
	DnsPolicyAssetType             = "dns.googleapis.com/Policy"
	ServiceUsageAssetType          = "serviceusage.googleapis.com/Service"
	GkeClusterAssetType            = "container.googleapis.com/Cluster"
	ComputeForwardingRuleAssetType = "compute.googleapis.com/ForwardingRule"
	IamRoleAssetType               = "iam.googleapis.com/Role"
	CloudFunctionAssetType         = "cloudfunctions.googleapis.com/CloudFunction"
	CloudRunService                = "run.googleapis.com/Service"
)
