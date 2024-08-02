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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

// https://github.com/elastic/integrations/tree/main/packages/cloud_security_posture/data_stream/findings/agent/stream
const (
	CIS_K8S       = "cis_k8s"
	CIS_EKS       = "cis_eks"
	CIS_AWS       = "cis_aws"
	CIS_GCP       = "cis_gcp"
	CIS_AZURE     = "cis_azure"
	ProviderAWS   = "aws"
	ProviderAzure = "azure"
	ProviderGCP   = "gcp"
)

var SupportedCIS = []string{CIS_AWS, CIS_K8S, CIS_EKS, CIS_GCP, CIS_AZURE}
