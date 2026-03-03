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

package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
)

func TestDefaultParentResolver_GetParent(t *testing.T) {
	ctx := context.Background()
	cfgSingle := config.GcpConfig{AccountType: config.SingleAccount, ProjectId: "my-project"}
	cfgOrg := config.GcpConfig{AccountType: config.OrganizationAccount, OrganizationId: "my-org"}

	t.Run("SingleAccount delegates to project resolver", func(t *testing.T) {
		mockProject := NewMockProjectParentResolver(t)
		mockOrg := NewMockOrganizationParentResolver(t)

		mockProject.EXPECT().
			GetProjectParent(ctx, cfgSingle, mock.AnythingOfType("[]option.ClientOption")).
			Return("projects/my-project", nil)

		resolver := &defaultParentResolver{project: mockProject, org: mockOrg}
		parent, err := resolver.GetParent(ctx, cfgSingle, nil)
		require.NoError(t, err)
		assert.Equal(t, "projects/my-project", parent)
	})

	t.Run("OrganizationAccount delegates to organization resolver", func(t *testing.T) {
		mockProject := NewMockProjectParentResolver(t)
		mockOrg := NewMockOrganizationParentResolver(t)

		mockOrg.EXPECT().
			GetOrganizationParent(ctx, cfgOrg, mock.AnythingOfType("[]option.ClientOption")).
			Return("organizations/my-org", nil)

		resolver := &defaultParentResolver{project: mockProject, org: mockOrg}
		parent, err := resolver.GetParent(ctx, cfgOrg, nil)
		require.NoError(t, err)
		assert.Equal(t, "organizations/my-org", parent)
	})

	t.Run("SingleAccount returns project resolver error", func(t *testing.T) {
		mockProject := NewMockProjectParentResolver(t)
		mockOrg := NewMockOrganizationParentResolver(t)
		wantErr := errors.New("project resolution failed")

		mockProject.EXPECT().
			GetProjectParent(ctx, cfgSingle, mock.AnythingOfType("[]option.ClientOption")).
			Return("", wantErr)

		resolver := &defaultParentResolver{project: mockProject, org: mockOrg}
		parent, err := resolver.GetParent(ctx, cfgSingle, nil)
		require.ErrorIs(t, err, wantErr)
		assert.Empty(t, parent)
	})

	t.Run("OrganizationAccount returns organization resolver error", func(t *testing.T) {
		mockProject := NewMockProjectParentResolver(t)
		mockOrg := NewMockOrganizationParentResolver(t)
		wantErr := errors.New("org resolution failed")

		mockOrg.EXPECT().
			GetOrganizationParent(ctx, cfgOrg, mock.AnythingOfType("[]option.ClientOption")).
			Return("", wantErr)

		resolver := &defaultParentResolver{project: mockProject, org: mockOrg}
		parent, err := resolver.GetParent(ctx, cfgOrg, nil)
		require.ErrorIs(t, err, wantErr)
		assert.Empty(t, parent)
	})

	t.Run("invalid account type returns error", func(t *testing.T) {
		mockProject := NewMockProjectParentResolver(t)
		mockOrg := NewMockOrganizationParentResolver(t)

		resolver := &defaultParentResolver{project: mockProject, org: mockOrg}
		cfgInvalid := config.GcpConfig{AccountType: "invalid"}
		parent, err := resolver.GetParent(ctx, cfgInvalid, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid gcp account type")
		assert.Empty(t, parent)
	})
}

func TestNewDefaultParentResolver(t *testing.T) {
	// Smoke test: constructor returns a non-nil resolver that can be called.
	// Full behavior is tested via GetParent with mocks and via credentials tests.
	mockAuth := NewMockDefaultCredentialsFinder(t)
	resolver := NewDefaultParentResolver(mockAuth)
	require.NotNil(t, resolver)
}
