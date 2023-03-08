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

package logging

import (
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type LoggingResource struct {
	awslib.AwsResource
}

type ConfigResource struct {
	awslib.AwsResource
}

func (r LoggingResource) GetData() any {
	return r.AwsResource
}

func (r LoggingResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.GetResourceArn(),
		Type:    fetching.CloudAudit,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
	}, nil
}
func (r LoggingResource) GetElasticCommonData() any { return nil }

func (c ConfigResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      c.GetResourceArn(),
		Type:    fetching.CloudConfig,
		SubType: c.GetResourceType(),
		Name:    c.GetResourceName(),
	}, nil
}

func (c ConfigResource) GetData() any {
	return c.AwsResource
}

func (c ConfigResource) GetElasticCommonData() any { return nil }
