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

package fetchers

import (
	"context"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMonitoringFactory_Create(t *testing.T) {
	awsconfig := &awslib.MockConfigProviderAPI{}
	awsconfig.EXPECT().InitializeAWSConfig(mock.Anything, mock.Anything).
		Call.
		Return(func(ctx context.Context, config aws.ConfigAWS) *awssdk.Config {
			return CreateSdkConfig(config, "us1-east")
		},
			func(ctx context.Context, config aws.ConfigAWS) error {
				return nil
			},
		)
	f := &MonitoringFactory{
		AwsConfigProvider: awsconfig,
	}
	cfg, err := agentconfig.NewConfigFrom(awsConfig)
	assert.NoError(t, err)
	fetcher, err := f.Create(logp.NewLogger("test"), cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher)
	monitoringFetcher, ok := fetcher.(*MonitoringFetcher)
	assert.True(t, ok)
	assert.Equal(t, monitoringFetcher.cfg.AwsConfig.AccessKeyID, "key")
	assert.Equal(t, monitoringFetcher.cfg.AwsConfig.SecretAccessKey, "secret")
}
