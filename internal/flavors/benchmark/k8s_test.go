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
	"errors"
	"testing"

	"github.com/elastic/cloudbeat/internal/config"
	k8sprovider "github.com/elastic/cloudbeat/internal/dataprovider/providers/k8s"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestK8S_Initialize(t *testing.T) {
	testhelper.SkipLong(t)

	t.Setenv("NODE_NAME", "node-name")
	tests := []struct {
		name           string
		cfg            config.Config
		clientProvider k8sprovider.ClientGetterAPI
		want           []string
		wantErr        string
	}{
		{
			name:    "nothing initialized",
			wantErr: "kubernetes client provider is uninitialized",
		},
		{
			name:           "kubernetes provider error",
			clientProvider: mockKubeClient(errors.New("some error")),
			wantErr:        "some error",
		},
		{
			name:           "no error",
			clientProvider: mockKubeClient(nil),
			want: []string{
				fetching.FileSystemType,
				fetching.KubeAPIType,
				fetching.ProcessType,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &K8S{
				ClientProvider: tt.clientProvider,
				leaderElector:  nil,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}
