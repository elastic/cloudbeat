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
	"context"
	"sync"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib/auth"
)

var ancestors = []string{
	"projects/5",
	"folders/4",
	"folders/3",
	"folders/2",
	"organizations/1",
}

type ResourceManagerTestSuite struct {
	suite.Suite
}

func TestResourceManagerTestSuite(t *testing.T) {
	s := new(ResourceManagerTestSuite)
	suite.Run(t, s)
}

func (s *ResourceManagerTestSuite) NewMockResourceManagerWrapper() *ResourceManagerWrapper {
	return &ResourceManagerWrapper{
		config: auth.GcpFactoryConfig{
			Parent:     "organizations/1",
			ClientOpts: []option.ClientOption{},
		},
		accountNamesCache: sync.Map{},
		getProjectDisplayName: func(_ context.Context, _ string) string {
			return "projectName"
		},
		getOrganizationDisplayName: func(_ context.Context, _ string) string {
			return "orgName"
		},
	}
}
func (s *ResourceManagerTestSuite) TestGetCloudMetadataOrg() {
	crm := s.NewMockResourceManagerWrapper()
	result := crm.GetCloudMetadata(context.Background(), &assetpb.Asset{
		Name:      "projects/1",
		Ancestors: ancestors,
	})
	s.Equal("5", result.AccountId)
	s.Equal("projectName", result.AccountName)
	s.Equal("1", result.OrganisationId)
	s.Equal("orgName", result.OrganizationName)
}
func (s *ResourceManagerTestSuite) TestGetCloudMetadataProject() {
	crm := s.NewMockResourceManagerWrapper()
	crm.config.Parent = "projects/5"
	result := crm.GetCloudMetadata(context.Background(), &assetpb.Asset{
		Name:      "projects/1",
		Ancestors: ancestors,
	})
	s.Equal("5", result.AccountId)
	s.Equal("projectName", result.AccountName)
	s.Equal("1", result.OrganisationId)
	s.Empty(result.OrganizationName) // no org name when Parent is a project
}

func (s *ResourceManagerTestSuite) TestGetOrganizationId() {
	s.Equal("1", getOrganizationId(ancestors))
}
func (s *ResourceManagerTestSuite) TestGetProjectId() {
	s.Equal("5", getProjectId(ancestors))
}
