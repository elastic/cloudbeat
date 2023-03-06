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

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type (
	Service interface {
		Describe(ctx context.Context) ([]SecurityHub, error)
	}

	SecurityHub struct {
		Enabled   bool
		Region    string
		AccountId string
		*securityhub.DescribeHubOutput
	}
)

func (s SecurityHub) GetResourceArn() string {
	if s.DescribeHubOutput == nil || s.HubArn == nil {
		return s.GetResourceName()
	}
	return *s.HubArn
}

func (s SecurityHub) GetResourceName() string {
	return fmt.Sprintf("securityhub-%s-%s", s.Region, s.AccountId)
}

func (s SecurityHub) GetResourceType() string {
	return fetching.SecurityHubType
}
