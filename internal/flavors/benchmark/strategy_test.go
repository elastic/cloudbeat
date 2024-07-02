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

package benchmark

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestGetStrategy(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent()) // GetStrategy should not start anything

	tests := []struct {
		cfg      config.Config
		wantType Strategy
		wantErr  bool
	}{
		{
			cfg:     config.Config{Benchmark: "unknown"},
			wantErr: true,
		},
		{
			cfg:      config.Config{Benchmark: config.CIS_AWS},
			wantType: &AWS{}, //nolint:exhaustruct
		},
		{
			cfg: config.Config{
				Benchmark: config.CIS_AWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{AccountType: config.SingleAccount},
				},
			},
			wantType: &AWS{}, //nolint:exhaustruct
		},
		{
			cfg: config.Config{
				Benchmark: config.CIS_AWS,
				CloudConfig: config.CloudConfig{
					Aws: config.AwsConfig{AccountType: config.OrganizationAccount},
				},
			},
			wantType: &AWSOrg{}, //nolint:exhaustruct
		},
		{
			cfg:      config.Config{Benchmark: config.CIS_EKS},
			wantType: &EKS{}, //nolint:exhaustruct
		},
		{
			cfg:      config.Config{Benchmark: config.CIS_K8S},
			wantType: &K8S{}, //nolint:exhaustruct
		},
		{
			cfg:      config.Config{Benchmark: config.CIS_GCP},
			wantType: &GCP{}, //nolint:exhaustruct
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T", tt.wantType), func(t *testing.T) {
			got, err := GetStrategy(&tt.cfg, testhelper.NewLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			assert.IsType(t, tt.wantType, got)
			require.NoError(t, got.checkDependencies())
		})
	}
}

type benchInit interface {
	initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error)
}

func testInitialize(t *testing.T, s benchInit, cfg *config.Config, wantErr string, want []string) {
	t.Helper()

	reg, dp, _, err := s.initialize(context.Background(), testhelper.NewLogger(t), cfg, make(chan fetching.ResourceInfo))
	if wantErr != "" {
		require.ErrorContains(t, err, wantErr)
		return
	}
	reg.Update()
	defer reg.Stop()

	require.NoError(t, err)
	assert.Len(t, reg.Keys(), len(want))

	eks, ok := s.(*EKS)
	if ok {
		require.NoError(t, eks.leaderElector.Run(context.Background()))
		defer eks.leaderElector.Stop()
	}
	k8s, ok := s.(*K8S)
	if ok {
		require.NoError(t, k8s.leaderElector.Run(context.Background()))
		defer k8s.leaderElector.Stop()
	}

	for _, fetcher := range want {
		ok := reg.ShouldRun(fetcher)
		assert.Truef(t, ok, "fetcher %s enabled", fetcher)
	}

	// TODO: gcp diff tests cover
	assert.NotNil(t, dp)
}

func mockKubeClient(err error) k8s.ClientGetterAPI {
	kube := k8s.MockClientGetterAPI{}
	on := kube.EXPECT().GetClient(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			k8sfake.NewSimpleClientset(
				&core_v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-name",
					},
				},
				&core_v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			), nil)
	} else {
		on.Return(nil, err)
	}
	return &kube
}

func mockAwsIdentityProvider(err error) *awslib.MockIdentityProviderGetter {
	identityProvider := &awslib.MockIdentityProviderGetter{}
	on := identityProvider.EXPECT().GetIdentity(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&cloud.Identity{
				Account: "test-account",
			},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return identityProvider
}
