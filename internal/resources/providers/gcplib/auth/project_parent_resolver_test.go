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
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
)

func TestNewDefaultProjectParentResolver(t *testing.T) {
	mockAuth := NewMockDefaultCredentialsFinder(t)
	resolver := NewDefaultProjectParentResolver(mockAuth)
	require.NotNil(t, resolver)
}

func TestDefaultProjectParentResolver_GetProjectParent(t *testing.T) {
	ctx := context.Background()

	t.Run("returns project parent from config when ProjectId is set", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{ProjectId: "my-project-id"}

		parent, err := resolver.GetProjectParent(ctx, cfg, nil)
		require.NoError(t, err)
		assert.Equal(t, "projects/my-project-id", parent)
		// FindDefaultCredentials must not be called when ProjectId is in config
		mockAuth.AssertNotCalled(t, "FindDefaultCredentials")
	})

	t.Run("returns project parent from default credentials when ProjectId and audience path not used", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{ProjectId: ""}

		mockAuth.EXPECT().
			FindDefaultCredentials(ctx).
			Return(&google.Credentials{ProjectID: "adc-project"}, nil)

		parent, err := resolver.GetProjectParent(ctx, cfg, nil)
		require.NoError(t, err)
		assert.Equal(t, "projects/adc-project", parent)
	})

	t.Run("returns error when FindDefaultCredentials fails", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{ProjectId: ""}
		wantErr := errors.New("credentials not found")

		mockAuth.EXPECT().
			FindDefaultCredentials(ctx).
			Return(nil, wantErr)

		parent, err := resolver.GetProjectParent(ctx, cfg, nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to get project ID")
		assert.ErrorIs(t, err, wantErr)
		assert.Empty(t, parent)
	})

	t.Run("returns ErrProjectNotFound when default credentials have empty ProjectID", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{ProjectId: ""}

		mockAuth.EXPECT().
			FindDefaultCredentials(ctx).
			Return(&google.Credentials{ProjectID: ""}, nil)

		parent, err := resolver.GetProjectParent(ctx, cfg, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrProjectNotFound)
		assert.Empty(t, parent)
	})

	t.Run("returns error when audience is set with clientOpts but audience has no valid project number", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{
			ProjectId: "",
			GcpClientOpt: config.GcpClientOpt{Audience: "//iam.googleapis.com/locations/global/not-a-valid-audience"},
		}
		clientOpts := []option.ClientOption{option.WithRequestReason("test")}

		parent, err := resolver.GetProjectParent(ctx, cfg, clientOpts)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrProjectNotFound)
		assert.ErrorContains(t, err, "audience does not contain a valid project number")
		assert.Empty(t, parent)
		// Should not fall back to ADC when clientOpts and Audience are present
		mockAuth.AssertNotCalled(t, "FindDefaultCredentials")
	})

	t.Run("falls back to default credentials when ProjectId is empty and no clientOpts", func(t *testing.T) {
		mockAuth := NewMockDefaultCredentialsFinder(t)
		resolver := NewDefaultProjectParentResolver(mockAuth)
		cfg := config.GcpConfig{
			ProjectId:    "",
			GcpClientOpt: config.GcpClientOpt{Audience: "//iam.googleapis.com/projects/123/locations/global"},
		}

		mockAuth.EXPECT().
			FindDefaultCredentials(ctx).
			Return(&google.Credentials{ProjectID: "fallback-project"}, nil)

		parent, err := resolver.GetProjectParent(ctx, cfg, nil)
		require.NoError(t, err)
		assert.Equal(t, "projects/fallback-project", parent)
	})
}
