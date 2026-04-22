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
	"testing"

	clientfeatures "k8s.io/client-go/features"
	"k8s.io/client-go/kubernetes/fake"
)

// noWatchListGates wraps the default feature gates and disables WatchListClient.
// With k8s v0.35+, WatchListClient is enabled by default, but the fake client used
// in tests does not support it (it never sends the required bookmark event), causing
// WaitForCacheSync to hang indefinitely.
type noWatchListGates struct {
	clientfeatures.Gates
}

func (g noWatchListGates) Enabled(key clientfeatures.Feature) bool {
	if key == clientfeatures.WatchListClient {
		return false
	}
	return g.Gates.Enabled(key)
}

func TestMain(m *testing.M) {
	clientfeatures.ReplaceFeatureGates(noWatchListGates{clientfeatures.FeatureGates()})
	// Pre-warm the k8s type converter (sync.Once in fake.NewClientset → NewTypeConverter).
	// This runs before m.Run() so the slow first-call cost doesn't count against
	// the per-test timeout imposed by the pre-commit go-test hook (-timeout 100ms).
	fake.NewClientset()
	m.Run()
}
