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

package k8s

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

func TestK8sDataProvider_EnrichEvent(t *testing.T) {
	options := []Option{
		WithClusterName("test_cluster"),
		WithLogger(testhelper.NewLogger(t)),
		WithConfig(&config.Config{
			Benchmark: config.CIS_K8S,
		}),
	}

	k := New(options...)
	e := &beat.Event{
		Fields: mapstr.M{},
	}
	err := k.EnrichEvent(e, fetching.ResourceMetadata{})
	assert.NoError(t, err)
	v, err := e.Fields.GetValue(clusterNameField)
	assert.NoError(t, err)
	assert.Equal(t, "test_cluster", v)
}
