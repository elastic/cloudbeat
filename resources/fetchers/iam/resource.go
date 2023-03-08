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

package iam

import (
	"fmt"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type IAMResource struct {
	awslib.AwsResource
	identity *awslib.Identity
}

func (r IAMResource) GetData() any {
	return r.AwsResource
}

func (r IAMResource) GetMetadata() (fetching.ResourceMetadata, error) {
	identifier := r.GetResourceArn()
	if identifier == "" {
		identifier = fmt.Sprintf("%s-%s", *r.identity.Account, r.GetResourceName())
	}

	return fetching.ResourceMetadata{
		ID:      identifier,
		Type:    fetching.CloudIdentity,
		SubType: r.GetResourceType(),
		Name:    r.GetResourceName(),
	}, nil
}
func (r IAMResource) GetElasticCommonData() any { return nil }
