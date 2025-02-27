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
	"context"
	"errors"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/preset"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

const resourceChBufferSize = 10000

type AWS struct {
	IdentityProvider awslib.IdentityProviderGetter
}

func (a *AWS) NewBenchmark(ctx context.Context, log *clog.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, _, err := a.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
	).Build(ctx, log, cfg, resourceCh, reg)
}

//revive:disable-next-line:function-result-limit
func (a *AWS) initialize(ctx context.Context, log *clog.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := a.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	var (
		awsConfig   *awssdk.Config
		awsIdentity *cloud.Identity
		err         error
	)

	awsConfig, awsIdentity, err = a.getIdentity(ctx, cfg)
	if err != nil && cfg.CloudConfig.Aws.Cred.DefaultRegion == "" {
		log.Warn("failed to initialize identity; retrying to check AWS Gov Cloud regions")
		cfg.CloudConfig.Aws.Cred.DefaultRegion = awslib.DefaultGovRegion
		awsConfig, awsIdentity, err = a.getIdentity(ctx, cfg)
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get AWS Identity: %w", err)
	}
	log.Info("successfully retrieved AWS Identity")

	return registry.NewRegistry(
		log,
		registry.WithFetchersMap(preset.NewCisAwsFetchers(ctx, log, *awsConfig, ch, awsIdentity)),
	), cloud.NewDataProvider(cloud.WithAccount(*awsIdentity)), nil, nil
}

func (a *AWS) getIdentity(ctx context.Context, cfg *config.Config) (*awssdk.Config, *cloud.Identity, error) {
	var awsConfig *awssdk.Config
	var err error

	if cfg.CloudConfig.Aws.CloudConnectors {
		awsConfig, err = awslib.InitializeAWSConfigCloudConnectors(ctx, cfg.CloudConfig.Aws)
	} else {
		awsConfig, err = awslib.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	awsIdentity, err := a.IdentityProvider.GetIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	return awsConfig, awsIdentity, nil
}

func (a *AWS) checkDependencies() error {
	if a.IdentityProvider == nil {
		return errors.New("aws identity provider is uninitialized")
	}
	return nil
}
