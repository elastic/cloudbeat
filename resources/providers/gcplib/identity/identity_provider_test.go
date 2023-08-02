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

package identity

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"google.golang.org/api/cloudresourcemanager/v3"

	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

func TestIdentityProvider_GetIdentity(t *testing.T) {
	tests := []struct {
		name    string
		service func() ResourceManager
		want    *cloud.Identity
		wantErr bool
	}{
		{
			name: "failed to get project info",
			service: func() ResourceManager {
				m := MockResourceManager{}
				m.EXPECT().projectsGet(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("permission denied"))
				return &m
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "Project info returned successfully",
			service: func() ResourceManager {
				m := MockResourceManager{}
				m.EXPECT().projectsGet(mock.Anything, mock.Anything).Return(&cloudresourcemanager.Project{
					DisplayName: "my proj",
					ProjectId:   "test-proj",
				}, nil)
				return &m
			},
			want: &cloud.Identity{
				Provider:     "gcp",
				Account:      "test-proj",
				AccountAlias: "my proj",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				service: tt.service(),
			}

			got, err := p.GetIdentity(context.Background(), &auth.GcpFactoryConfig{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIdentity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIdentity() got = %v, want %v", got, tt.want)
			}
		})
	}
}
