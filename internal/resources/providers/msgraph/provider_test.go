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
	"testing"

	graphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPageIterator_ServicePrincipalable_NoPanicWithSubtype is the targeted regression test.
// The MS Graph API returns AgentIdentityBlueprintPrincipal objects on tenants
// with Microsoft 365 Copilot/Agents. The old PageIterator[*ServicePrincipal] panicked because
// its internal convertToPage does value.Index(i).Interface().(*ServicePrincipal) — an unsafe
// assertion that fails when the concrete type is *AgentIdentityBlueprintPrincipal.
//
// Using PageIterator[ServicePrincipalable] instead makes the assertion succeed for any type
// that satisfies the interface, including all current and future ServicePrincipal subtypes.
func TestPageIterator_ServicePrincipalable_NoPanicWithSubtype(t *testing.T) {
	agentPrincipal := models.NewAgentIdentityBlueprintPrincipal()
	regularPrincipal := models.NewServicePrincipal()

	response := models.NewServicePrincipalCollectionResponse()
	response.SetValue([]models.ServicePrincipalable{agentPrincipal, regularPrincipal})

	// nil adapter is safe when there is no @odata.nextLink (single page).
	pageIterator, err := graphcore.NewPageIterator[models.ServicePrincipalable](
		response,
		nil,
		models.CreateServicePrincipalCollectionResponseFromDiscriminatorValue,
	)
	require.NoError(t, err)

	var collected []models.ServicePrincipalable
	require.NotPanics(t, func() {
		iterErr := pageIterator.Iterate(context.Background(), func(item models.ServicePrincipalable) bool {
			collected = append(collected, item)
			return true
		})
		require.NoError(t, iterErr)
	})

	require.Len(t, collected, 2)
	assert.IsType(t, agentPrincipal, collected[0], "first item should be *AgentIdentityBlueprintPrincipal")
	assert.IsType(t, regularPrincipal, collected[1], "second item should be *ServicePrincipal")
}
