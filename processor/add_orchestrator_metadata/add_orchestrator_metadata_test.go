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

package add_orchestrator_metadata

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/suite"
)

type AddMetadataTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestAddOrchestratorMetadataTestSuite(t *testing.T) {
	s := new(AddMetadataTestSuite)
	s.log = logp.NewLogger("cloudbeat_add_orchestrator_metadata_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *AddMetadataTestSuite) TestAddOrchestratorMetadataRun() {
	var tests = []struct {
		clusterName string
		clusterId   string
	}{
		{
			"some-cluster-name",
			"some-cluster-id",
		},
		{
			"some-cluster-name-2",
			"some-cluster-id-2",
		},
	}
	for _, t := range tests {
		mock := &clusterHelperMock{
			id:          t.clusterId,
			clusterName: t.clusterName,
		}
		processor := &processor{
			helper: mock,
		}

		e := beat.Event{
			Fields: make(mapstr.M),
		}

		event, err := processor.Run(&e)
		s.NoError(err)

		res, err := event.GetValue("orchestrator.cluster.name")
		s.NoError(err)
		s.Equal(t.clusterName, res)

		res, err = event.GetValue("cluster_id")
		s.NoError(err)
		s.Equal(t.clusterId, res)
	}
}

func (s *AddMetadataTestSuite) TestAddOrchestratorMetadataRunWhenNoClusterName() {
	var tests = []struct {
		clusterName string
		clusterId   string
	}{
		{
			"",
			"some-cluster-id",
		},
	}
	for _, t := range tests {
		mock := &clusterHelperMock{
			id:          t.clusterId,
			clusterName: t.clusterName,
		}
		processor := &processor{
			helper: mock,
		}

		e := beat.Event{
			Fields: make(mapstr.M),
		}

		event, err := processor.Run(&e)
		s.NoError(err)

		res, err := event.GetValue("orchestrator.cluster.name")
		s.Error(err)
		s.ErrorContains(err, "key not found")
		s.Empty(res)

		res, err = event.GetValue("cluster_id")
		s.NoError(err)
		s.Equal(t.clusterId, res)

	}
}

type clusterHelperMock struct {
	id          string
	clusterName string
}

func (m *clusterHelperMock) GetClusterMetadata() ClusterMetadata {
	return ClusterMetadata{clusterName: m.clusterName, clusterId: m.id}
}
