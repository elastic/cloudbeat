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

package dataprovider

import (
	"context"
	"github.com/elastic/cloudbeat/version"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

var bgCtx = context.Background()

func TestCommonDataProvider_FetchCommonData(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name          string
		args          args
		want          CommonK8sData
		wantErr       bool
		k8sCommonData CommonK8sData
	}{
		{
			name: "test common data",
			args: args{
				ctx: bgCtx,
			},
			want: CommonK8sData{
				clusterId: "testing_namespace_uid",
				nodeId:    "testing_node_uid",
				serverVersion: version.Version{
					Version: "testing_version",
				},
			},
			k8sCommonData: CommonK8sData{
				clusterId: "testing_namespace_uid",
				nodeId:    "testing_node_uid",
				serverVersion: version.Version{
					Version: "testing_version",
				},
			},
			wantErr: false,
		},
		{
			name: "test common data without k8s",
			args: args{
				ctx: bgCtx,
			},
			want: CommonK8sData{
				clusterId:     "",
				nodeId:        "",
				serverVersion: version.Version{},
			},
			k8sCommonData: CommonK8sData{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sDataProviderMock := &mockK8sDataProvider{}
			k8sDataProviderMock.EXPECT().CollectK8sData(mock.Anything).Return(&tt.k8sCommonData)
			cdProvider := createCommonDataProvider(k8sDataProviderMock)

			got, err := cdProvider.FetchCommonData(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCommonData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want.clusterId, got.GetData().clusterId, "commonData clusterId is not correct")
			assert.Equal(t, tt.want.nodeId, got.GetData().nodeId, "commonData nodeId is not correct")
			assert.Equal(t, tt.want.serverVersion.Version, got.GetData().versionInfo.Kubernetes.Version, "k8s server version is empty")
			assert.NotEmpty(t, got.GetData().versionInfo.Version, "Beat's version is empty")
		})
	}
}

func TestCommonData_GetResourceId(t *testing.T) {
	type CDFields struct {
		clusterId string
		nodeId    string
	}
	type args struct {
		metadata fetching.ResourceMetadata
	}
	tests := []struct {
		name       string
		commonData CDFields
		args       args
		want       string
	}{
		{
			name: "Get kube api resource id",
			commonData: CDFields{
				clusterId: "cluster-test",
				nodeId:    "nodeid-test",
			},
			args: args{
				metadata: fetching.ResourceMetadata{
					ID:      "uuid-test",
					Type:    fetchers.K8sObjType,
					SubType: "pod",
					Name:    "pod-test-123",
				},
			},
			want: "uuid-test",
		},
		{
			name: "Get FS resource id",
			commonData: CDFields{
				clusterId: "cluster-test",
				nodeId:    "nodeid-test",
			},
			args: args{
				metadata: fetching.ResourceMetadata{
					ID:      "1234",
					Type:    "file",
					SubType: "file",
					Name:    "/etc/passwd",
				},
			},
			want: uuid.NewV5(uuidNamespace, "cluster-test"+"nodeid-test"+"1234").String(),
		},
		{
			name: "Get AWS resource id",
			commonData: CDFields{
				clusterId: "cluster-test",
			},
			args: args{
				metadata: fetching.ResourceMetadata{
					ID:   "1234",
					Type: "load-balancer",
					Name: "aws-loadbalancer",
				},
			},
			want: uuid.NewV5(uuidNamespace, "cluster-test"+"1234").String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := CommonData{
				clusterId: tt.commonData.clusterId,
				nodeId:    tt.commonData.nodeId,
			}
			if got := cd.GetResourceId(tt.args.metadata); got != tt.want {
				t.Errorf("GetResourceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createCommonDataProvider(mock *mockK8sDataProvider) CommonDataProvider {
	return CommonDataProvider{
		log:             logp.NewLogger("cloudbeat_common_data_provider_test"),
		cfg:             &config.Config{},
		k8sDataProvider: mock,
	}
}
