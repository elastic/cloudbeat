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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

func TestNewDefaultOrganizationParentResolver(t *testing.T) {
	resolver := NewDefaultOrganizationParentResolver()
	require.NotNil(t, resolver)
}

func TestDefaultOrganizationParentResolver_GetOrganizationParent(t *testing.T) {
	ctx := context.Background()

	t.Run("returns organization parent from config when OrganizationId is set", func(t *testing.T) {
		resolver := NewDefaultOrganizationParentResolver()
		cfg := config.GcpConfig{OrganizationId: "123456789"}

		parent, err := resolver.GetOrganizationParent(ctx, cfg, nil)
		require.NoError(t, err)
		assert.Equal(t, "organizations/123456789", parent)
	})

	t.Run("returns ErrMissingOrgId when OrganizationId is empty and no clientOpts", func(t *testing.T) {
		resolver := NewDefaultOrganizationParentResolver()
		cfg := config.GcpConfig{OrganizationId: ""}

		parent, err := resolver.GetOrganizationParent(ctx, cfg, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingOrgId)
		assert.Empty(t, parent)
	})

	t.Run("returns ErrMissingOrgId when OrganizationId is empty and no audience", func(t *testing.T) {
		resolver := NewDefaultOrganizationParentResolver()
		cfg := config.GcpConfig{
			OrganizationId: "",
			GcpClientOpt:   config.GcpClientOpt{Audience: ""},
		}
		clientOpts := []option.ClientOption{option.WithRequestReason("test")}

		parent, err := resolver.GetOrganizationParent(ctx, cfg, clientOpts)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingOrgId)
		assert.Empty(t, parent)
	})

	t.Run("returns error when clientOpts and audience set but audience has no valid project number", func(t *testing.T) {
		resolver := NewDefaultOrganizationParentResolver()
		cfg := config.GcpConfig{
			OrganizationId: "",
			GcpClientOpt:   config.GcpClientOpt{Audience: "//iam.googleapis.com/locations/global/not-a-valid-audience"},
		}
		clientOpts := []option.ClientOption{option.WithRequestReason("test")}

		parent, err := resolver.GetOrganizationParent(ctx, cfg, clientOpts)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingOrgId)
		assert.ErrorContains(t, err, "audience does not contain a valid project number")
		assert.Empty(t, parent)
	})
}
