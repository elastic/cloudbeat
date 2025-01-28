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

package azurefetcher

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/microsoft/kiota-abstractions-go/store"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestActiveDirectoryFetcher_Fetch(t *testing.T) {
	servicePrincipal := &models.ServicePrincipal{
		DirectoryObject: models.DirectoryObject{
			Entity: models.Entity{},
		},
	}
	appOwnerOrganizationId, _ := uuid.NewUUID()
	values := map[string]any{
		"id":                     pointers.Ref("id"),
		"displayName":            pointers.Ref("dn"),
		"appOwnerOrganizationId": &appOwnerOrganizationId,
	}
	store := store.NewInMemoryBackingStore()
	for k, v := range values {
		_ = store.Set(k, v)
	}
	servicePrincipal.SetBackingStore(store)

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAzureServicePrincipal,
			[]string{"id"},
			"dn",
			inventory.WithRawAsset(values),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AzureCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id: appOwnerOrganizationId.String(),
				},
				Service: &inventory.AssetCloudService{
					Name: "Azure",
				},
			}),
		),
	}

	// setup
	logger := clog.NewLogger("azurefetcher_test")
	provider := newMockActivedirectoryProvider(t)

	provider.EXPECT().ListServicePrincipals(mock.Anything).Return(
		[]*models.ServicePrincipal{servicePrincipal}, nil,
	)

	fetcher := newActiveDirectoryFetcher(logger, provider)
	// test & compare
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestActiveDirectoryFetcher_FetchError(t *testing.T) {
	// set up log capture
	var log *clog.Logger
	logCaptureBuf := &bytes.Buffer{}
	{
		replacement := zap.WrapCore(func(zapcore.Core) zapcore.Core {
			return zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(logCaptureBuf),
				zapcore.DebugLevel,
			)
		})
		log = clog.NewLogger("test").WithOptions(replacement)
	}

	provider := newMockActivedirectoryProvider(t)
	provider.EXPECT().ListServicePrincipals(mock.Anything).Return(
		[]*models.ServicePrincipal{}, fmt.Errorf("! error listing service principals"),
	)

	fetcher := newActiveDirectoryFetcher(log, provider)

	// collect
	ch := make(chan inventory.AssetEvent)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	fetcher.Fetch(ctx, ch)
	<-ctx.Done()
	close(ch)

	received := []inventory.AssetEvent{}
	for event := range ch {
		received = append(received, event)
	}

	require.Empty(t, received, "expected error, not AssetEvents")
	require.NotEmpty(t, logCaptureBuf, "expected logs, but captured none")
	require.Contains(t, logCaptureBuf.String(), "error listing service principals", "expected message not found")
}
