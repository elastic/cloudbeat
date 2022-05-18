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
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

type args struct {
	resource  manager.ResourceMap
	metadata  CycleMetadata
	namespace string
}

type testAttr struct {
	name    string
	args    args
	wantErr bool
	mocks   []MethodMock
}

type MethodMock struct {
	methodName string
	args       []interface{}
	returnArgs []interface{}
}

const (
	opaResultsFileName = "opa_results.json"
	testIndex          = "test_index"
)

var fetcherResult = fetchers.FileSystemResource{
	FileName: "scheduler.conf",
	FileMode: "700",
	Gid:      "root",
	Uid:      "root",
	Path:     "/hostfs/etc/kubernetes/scheduler.conf",
	Inode:    "8901",
	SubType:  "file",
}

var (
	opaResults   evaluator.RuleResult
	resourcesMap = map[string][]fetching.Resource{fetchers.FileSystemType: {fetcherResult}}
	ctx          = context.Background()
)

type EventsCreatorTestSuite struct {
	suite.Suite
	cycleId         uuid.UUID
	mockedEvaluator evaluator.MockedEvaluator
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(EventsCreatorTestSuite))
}

func (s *EventsCreatorTestSuite) SetupSuite() {
	err := parseJsonfile(opaResultsFileName, &opaResults)
	if err != nil {
		logp.L().Errorf("Could not parse Json file: %v", err)
		return
	}
	s.cycleId, _ = uuid.NewV4()
}

func (s *EventsCreatorTestSuite) SetupTest() {
	s.mockedEvaluator = evaluator.MockedEvaluator{}
}

func (s *EventsCreatorTestSuite) TestTransformer_ProcessAggregatedResources() {
	var tests = []testAttr{
		{
			name: "All events propagated as expected",
			args: args{
				resource:  resourcesMap,
				metadata:  CycleMetadata{CycleId: s.cycleId},
				namespace: "kube-system",
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("Result")},
				returnArgs: []interface{}{mock.Anything, nil},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{opaResults.Findings, nil},
			},
			},
			wantErr: false,
		},
		{
			name: "Events should not be created due to a policy error",
			args: args{
				resource:  resourcesMap,
				metadata:  CycleMetadata{CycleId: s.cycleId},
				namespace: "kube-system",
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("Result")},
				returnArgs: []interface{}{mock.Anything, errors.New("policy err")},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{opaResults.Findings, nil},
			},
			},
			wantErr: true,
		},
		{
			name: "Events should not be created due to a parse error",
			args: args{
				resource:  resourcesMap,
				metadata:  CycleMetadata{CycleId: s.cycleId},
				namespace: "kube-system",
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("Result")},
				returnArgs: []interface{}{mock.Anything, nil},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{nil, errors.New("parse err")},
			},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.SetupTest()
		s.Run(tt.name, func() {
			for _, methodMock := range tt.mocks {
				s.mockedEvaluator.On(methodMock.methodName, methodMock.args...).Return(methodMock.returnArgs...)
			}

			//Need to add services
			kc := k8sfake.NewSimpleClientset()

			namespace := &v1.Namespace{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "apps/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.args.namespace,
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

			_, err := kc.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
			s.NoError(err)

			_, err = kc.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
			s.NoError(err)

			cdp := CommonDataProvider{
				kubeClient: kc,
				cfg:        config.Config{},
			}

			// libbeat DiscoverKubernetesNode performs a fallback to environment variable NODE_NAME
			os.Setenv("NODE_NAME", "testing_node")

			commonData, err := cdp.FetchCommonData(ctx)
			s.NoError(err)

			s.Equal(commonData.GetData().clusterId, "testing_namespace_uid", "commonData clusterId is not correct")
			s.Equal(commonData.GetData().nodeId, "testing_node_uid", "commonData nodeId is not correct")

			transformer := NewTransformer(ctx, &s.mockedEvaluator, commonData, testIndex)

			generatedEvents := transformer.ProcessAggregatedResources(tt.args.resource, tt.args.metadata)

			if tt.wantErr {
				s.Equal(0, len(generatedEvents))
			}

			for _, event := range generatedEvents {
				resource := event.Fields["resource"].(fetching.ResourceFields)
				s.Equal(s.cycleId, event.Fields["cycle_id"], "event cycle_id is not correct")
				s.NotEmpty(event.Timestamp, `event timestamp is missing`)
				s.NotEmpty(event.Fields["result"], "event result is missing")
				s.NotEmpty(event.Fields["rule"], "event rule is missing")
				s.NotEmpty(resource.Raw, "raw resource is missing")
				s.NotEmpty(resource.SubType, "resource sub type is missing")
				s.NotEmpty(resource.ID, "resource ID is missing")
				s.NotEmpty(resource.Type, "resource  type is missing")
				s.NotEmpty(event.Fields["type"], "resource type is missing") // for BC sake
			}
		})
	}
}

func parseJsonfile(filename string, data interface{}) error {
	fetcherDataFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fetcherDataFile.Close()

	byteValue, err := ioutil.ReadAll(fetcherDataFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, data)
	if err != nil {
		return err
	}
	return nil
}
