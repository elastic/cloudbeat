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

package transformer

import (
	"context"
	"os"
	"testing"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
)

var bgCtx = context.Background()

type clusterNameProviderMock struct {
	clusterName string
}

func (c clusterNameProviderMock) GetClusterName(_ context.Context, _ config.Config) (string, error) {
	return c.clusterName, nil
}

func TestCommonDataProvider_FetchCommonData(t *testing.T) {
	clusterName := "my-cluster"
	cdProvider := CommonDataProvider{
		log:                 logp.NewLogger("cloudbeat_common_data_provider_test"),
		kubeClient:          k8sFake.NewSimpleClientset(),
		cfg:                 &config.Config{},
		clusterNameProvider: clusterNameProviderMock{clusterName},
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    CommonData
		wantErr bool
	}{
		{
			name: "test common data",
			args: args{
				ctx: bgCtx,
			},
			want: CommonData{
				clusterId: "testing_namespace_uid",
				nodeId:    "testing_node_uid",
			},
			wantErr: false,
		},
	}
	adjustK8sCluster(t, &cdProvider)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cdProvider.FetchCommonData(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCommonData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want.clusterId, got.GetData().clusterId, "commonData clusterId is not correct")
			assert.Equal(t, tt.want.nodeId, got.GetData().nodeId, "commonData nodeId is not correct")
			assert.Equal(t, clusterName, got.GetData().clusterName, "commonData nodeId is not correct")
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

func adjustK8sCluster(t *testing.T, cdProvider *CommonDataProvider) {
	// libbeat DiscoverKubernetesNode performs a fallback to environment variable NODE_NAME
	os.Setenv("NODE_NAME", "testing_node")
	// Need to add services
	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  "testing_namespace_uid",
		},
	}

	node := &v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "testing_node",
			UID:  "testing_node_uid",
		},
	}

	_, err := cdProvider.kubeClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	assert.NoError(t, err)

	_, err = cdProvider.kubeClient.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	assert.NoError(t, err)
}
