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

package conditions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PodNameEnvar           = "POD_NAME"
	DefaultLeaderLeaseName = "elastic-agent-cluster-leader"
	DefaultLeaderValue     = false
)

type leaseProvider struct {
	ctx    context.Context
	client kubernetes.Interface
}

func NewLeaderLeaseProvider(ctx context.Context, client kubernetes.Interface) LeaderLeaseProvider {
	return &leaseProvider{ctx, client}
}

func (l *leaseProvider) IsLeader() (bool, error) {
	leases, err := l.client.CoordinationV1().Leases("kube-system").List(l.ctx, v1.ListOptions{})
	if err != nil {
		return DefaultLeaderValue, err
	}

	for _, lease := range leases.Items {
		if lease.Name == DefaultLeaderLeaseName {
			podid := lastPart(*lease.Spec.HolderIdentity)

			if podid == l.currentPodID() {
				return true, nil
			}

			return false, nil
		}
	}

	return DefaultLeaderValue, fmt.Errorf("could not find lease %v in Kube leases", DefaultLeaderLeaseName)
}

func (l *leaseProvider) currentPodID() string {
	pod := os.Getenv(PodNameEnvar)

	return lastPart(pod)
}

func lastPart(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
