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

package flavors

import (
	"context"
	"testing"
	"time"

	awssdk_s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/evaluator"
	beater_testing "github.com/elastic/cloudbeat/flavors/testing"
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
	awslib_s3 "github.com/elastic/cloudbeat/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/transformer"
	"github.com/elastic/cloudbeat/uniqueness"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Cloudbeat(t *testing.T) {
	cfg := config.Config{
		BundlePath: "/Users/olegsuharevich/workspace/elastic/cloudbeat/bundle.tar.gz",
		Benchmark:  config.CIS_AWS,
		Fetchers: []*agentconfig.C{
			agentconfig.MustNewConfigFrom(mapstr.M{"name": "aws-s3"}),
		},
		Processors: processors.PluginConfig{},
	}
	cloudbeat, err := newTestingCloudbeat(cfg)
	assert.NoError(t, err)
	go func() {
		time.Sleep(time.Second * 3)
		cloudbeat.Stop()
	}()
	err = cloudbeat.Run(newTestingBeat())
	assert.NoError(t, err)
}

func newTestingCloudbeat(cfg config.Config) (*posture, error) {
	ctx, cancel := context.WithCancel(context.Background())
	log := logp.NewLogger("test")
	resourceChan := make(chan fetching.ResourceInfo)
	leader := uniqueness.MockManager{}
	leader.EXPECT().Run(mock.Anything).Return(nil)

	s3Mock := &awslib_s3.MockClient{}

	s3Mock.EXPECT().ListBuckets(mock.Anything, &awssdk_s3.ListBucketsInput{}).Return(nil, nil)

	reg, err := initRegistry(ctx, log, &cfg, resourceChan, &leader)
	if err != nil {
		cancel()
		return nil, err
	}
	data, err := fetchersManager.NewData(log, time.Second, time.Second*5, reg)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx, log, &cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	t := transformer.NewTransformer(log, nil, "test-index")
	return &posture{
		flavorBase: flavorBase{
			ctx:         ctx,
			cancel:      cancel,
			config:      &cfg,
			log:         log,
			transformer: t,
		},
		leader:     &leader,
		data:       data,
		evaluator:  eval,
		resourceCh: resourceChan,
	}, nil
}

func newTestingBeat() *beat.Beat {
	p := &beater_testing.MockPipeline{}
	c := &beater_testing.MockClient{}
	p.EXPECT().Connect().Return(c, nil)
	c.EXPECT().PublishAll(mock.Anything)
	return &beat.Beat{
		Publisher: p,
	}
}
