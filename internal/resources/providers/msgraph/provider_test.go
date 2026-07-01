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

package msgraph

import (
	"context"
	"errors"
	"testing"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubAdapter is a minimal RequestAdapter for pagination tests. It only implements Send;
// all other methods are satisfied by the embedded interface (calling them would panic,
// but pagination only requires Send).
type stubAdapter struct {
	abstractions.RequestAdapter
	sendFn func() (absser.Parsable, error)
}

func (s *stubAdapter) Send(_ context.Context, _ *abstractions.RequestInformation, _ absser.ParsableFactory, _ abstractions.ErrorMappings) (absser.Parsable, error) {
	return s.sendFn()
}

func newPage(items []models.ServicePrincipalable, nextLink *string) *models.ServicePrincipalCollectionResponse {
	r := models.NewServicePrincipalCollectionResponse()
	r.SetValue(items)
	r.SetOdataNextLink(nextLink)
	return r
}

func ptr(s string) *string { return &s }

// noopAdapter satisfies NewPageIterator's non-nil reqAdapter requirement for tests with no next
// link; Send is never expected to be called in that case.
func noopAdapter() *stubAdapter {
	return &stubAdapter{sendFn: func() (absser.Parsable, error) {
		panic("adapter should not be called when there is no next link")
	}}
}

func TestPageIterator_EmptyResponse(t *testing.T) {
	// nil value and an empty slice are both valid; both should yield zero items.
	for _, items := range [][]models.ServicePrincipalable{nil, {}} {
		pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
			newPage(items, nil), noopAdapter(),
			models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
		)
		require.NoError(t, err)

		var collected []models.ServicePrincipalable
		require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
			collected = append(collected, item)
			return true
		}))
		assert.Empty(t, collected)
	}
}

func TestPageIterator_RegularPrincipalOnly(t *testing.T) {
	sp := models.NewServicePrincipal()
	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		newPage([]models.ServicePrincipalable{sp}, nil), noopAdapter(),
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
		collected = append(collected, item)
		return true
	}))
	require.Len(t, collected, 1)
	assert.IsType(t, sp, collected[0])
}

// TestPageIterator_NoPanicWithSubtype is a regression test.
// AgentIdentityBlueprintPrincipal is a ServicePrincipal subtype provisioned automatically
// on tenants with Microsoft 365 Copilot/Agents. PageIterator[*ServicePrincipal] panicked on
// these items because convertToPage does value.Index(i).Interface().(T) — an unsafe type
// assertion that fails when the concrete type is *AgentIdentityBlueprintPrincipal.
// PageIterator[ServicePrincipalable] succeeds for any type that satisfies the interface.
func TestPageIterator_NoPanicWithSubtype(t *testing.T) {
	agent := models.NewAgentIdentityBlueprintPrincipal()
	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		newPage([]models.ServicePrincipalable{agent}, nil), noopAdapter(),
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NotPanics(t, func() {
		require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
			collected = append(collected, item)
			return true
		}))
	})
	require.Len(t, collected, 1)
	assert.IsType(t, agent, collected[0])
}

func TestPageIterator_MixedTypes(t *testing.T) {
	agent := models.NewAgentIdentityBlueprintPrincipal()
	regular := models.NewServicePrincipal()
	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		newPage([]models.ServicePrincipalable{agent, regular}, nil), noopAdapter(),
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
		collected = append(collected, item)
		return true
	}))
	require.Len(t, collected, 2)
	assert.IsType(t, agent, collected[0])
	assert.IsType(t, regular, collected[1])
}

// TestPageIterator_Pagination_SubtypeOnSecondPage verifies the fix holds across page
// boundaries. convertToPage is called for every fetched page, so a subtype on page 2
// would also have triggered the original panic.
func TestPageIterator_Pagination_SubtypeOnSecondPage(t *testing.T) {
	sp := models.NewServicePrincipal()
	agent := models.NewAgentIdentityBlueprintPrincipal()

	adapter := &stubAdapter{sendFn: func() (absser.Parsable, error) {
		return newPage([]models.ServicePrincipalable{agent}, nil), nil
	}}
	page1 := newPage([]models.ServicePrincipalable{sp}, ptr("https://graph.microsoft.com/v1.0/servicePrincipals?$skiptoken=X"))

	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		page1, adapter,
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
		collected = append(collected, item)
		return true
	}))
	require.Len(t, collected, 2)
	assert.IsType(t, sp, collected[0])
	assert.IsType(t, agent, collected[1])
}

func TestPageIterator_Pagination_AdapterError(t *testing.T) {
	sp := models.NewServicePrincipal()
	adapter := &stubAdapter{sendFn: func() (absser.Parsable, error) {
		return nil, errors.New("network error fetching page 2")
	}}
	page1 := newPage([]models.ServicePrincipalable{sp}, ptr("https://graph.microsoft.com/v1.0/servicePrincipals?$skiptoken=X"))

	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		page1, adapter,
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	err = pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
		collected = append(collected, item)
		return true
	})
	require.ErrorContains(t, err, "network error fetching page 2")
	assert.Len(t, collected, 1, "items from page 1 should have been collected before the error")
}

func TestPageIterator_EarlyStop(t *testing.T) {
	sp1, sp2 := models.NewServicePrincipal(), models.NewServicePrincipal()
	// A next link is present to confirm that early stop also prevents the adapter from being called.
	adapterCalled := false
	adapter := &stubAdapter{sendFn: func() (absser.Parsable, error) {
		adapterCalled = true
		return newPage(nil, nil), nil
	}}
	page1 := newPage([]models.ServicePrincipalable{sp1, sp2}, ptr("https://graph.microsoft.com/v1.0/servicePrincipals?$skiptoken=X"))

	pi, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		page1, adapter,
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NoError(t, pi.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
		collected = append(collected, item)
		return false // stop after the first item
	}))
	assert.Len(t, collected, 1)
	assert.False(t, adapterCalled, "adapter should not be called when iteration is stopped early")
}
