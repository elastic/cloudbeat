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

package configservice

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	configSDK "github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Client interface {
	DescribeConfigurationRecorders(ctx context.Context, params *configSDK.DescribeConfigurationRecordersInput, optFns ...func(*configSDK.Options)) (*configSDK.DescribeConfigurationRecordersOutput, error)
	DescribeConfigurationRecorderStatus(ctx context.Context, params *configSDK.DescribeConfigurationRecorderStatusInput, optFns ...func(*configSDK.Options)) (*configSDK.DescribeConfigurationRecorderStatusOutput, error)
}

type ConfigService interface {
	DescribeConfigRecorders(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log          *logp.Logger
	awsAccountId string
	clients      map[string]Client
}

type Config struct {
	// Currently, only one configuration recorder per region is supported but might change in the future, hence the array.
	Recorders []Recorder `json:"recorders"`
	accountId string
	region    string
}

type Recorder struct {
	types.ConfigurationRecorder
	Status []types.ConfigurationRecorderStatus `json:"statuses"`
}

func NewProvider(ctx context.Context, log *logp.Logger, cfg aws.Config, factory awslib.CrossRegionFactory[Client], accountId string) *Provider {
	f := func(cfg aws.Config) Client {
		return configSDK.NewFromConfig(cfg)
	}

	m := factory.NewMultiRegionClients(ctx, awslib.AllRegionSelector(), cfg, f, log)
	return &Provider{
		log:          log,
		clients:      m.GetMultiRegionsClientMap(),
		awsAccountId: accountId,
	}
}

func (c Config) GetResourceArn() string {
	return fmt.Sprintf("config-service-%s-%s", c.region, c.accountId)
}

func (c Config) GetResourceName() string {
	return fmt.Sprintf("config-service-%s-%s", c.region, c.accountId)
}

func (c Config) GetResourceType() string {
	return fetching.ConfigServiceResourceType
}

func (c Config) GetRegion() string {
	return c.region
}
