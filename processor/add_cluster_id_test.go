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

package add_cluster_id

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/stretchr/testify/suite"
)

type AddClusterIdTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestAddClusterIdTestSuite(t *testing.T) {
	s := new(AddClusterIdTestSuite)
	s.log = logp.NewLogger("cloudbeat_add_cluster_id_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *AddClusterIdTestSuite) TestClusterIdProcessor() {
	tests := []string{
		"abc",
		"some-cluster-id",
	}

	for _, t := range tests {
		mock := &clusterHelperMock{
			id: t,
		}

		processor := &addClusterID{
			helper: mock,
			config: defaultConfig(),
		}

		e := beat.Event{
			Fields: make(common.MapStr),
		}
		event, err := processor.Run(&e)
		s.NoError(err)

		res, err := event.GetValue("cluster_id")
		s.NoError(err)
		s.Equal(t, res)
	}
}

type clusterHelperMock struct {
	id string
}

func (m *clusterHelperMock) ClusterId() string {
	return m.id
}
