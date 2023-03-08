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

package monitoring

import (
	"fmt"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/aws_cis/monitoring"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/securityhub"
)

type MonitoringResource struct {
	monitoring.Resource
	identity *awslib.Identity
}
type SecurityHubResource struct {
	securityhub.SecurityHub
}

func (r MonitoringResource) GetData() any {
	return r
}

func (r MonitoringResource) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("cloudtrail-%s", *r.identity.Account)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.MultiTrailsType,
		Name:    id,
	}, nil
}
func (r MonitoringResource) GetElasticCommonData() any { return nil }

func (s SecurityHubResource) GetData() any {
	return s
}

func (s SecurityHubResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      s.GetResourceArn(),
		Name:    s.GetResourceName(),
		Type:    fetching.MonitoringIdentity,
		SubType: fetching.SecurityHubType,
	}, nil
}

func (s SecurityHubResource) GetElasticCommonData() any { return nil }
