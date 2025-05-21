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
	"slices"
	"strings"
	"sync"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/errorhandler"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/fetching/preset"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/samber/lo"
)

const resourceChBufferSize = 10000

type AWS struct {
	IdentityProvider awslib.IdentityProviderGetter
	errorPublisher   ErrorPublisher
	errorProcessor   *AWSErrorProcessor
}

func (a *AWS) NewBenchmark(ctx context.Context, log *clog.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, _, err := a.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
	).Build(ctx, log, cfg, resourceCh, reg, a)
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
		registry.WithFetchersMap(preset.NewCisAwsFetchers(ctx, log, *awsConfig, ch, awsIdentity, a.errorPublisher)),
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

func (a *AWS) Prepare(ctx context.Context, _ cycle.Metadata) error {
	a.errorPublisher.Reset(ctx)
	a.errorProcessor.Clear()
	return nil
}

func (a *AWS) ErrorProcessor() ErrorProcessor {
	return a.errorProcessor
}

func NewAWSErrorProcessor(log *clog.Logger) *AWSErrorProcessor {
	return &AWSErrorProcessor{
		log: log,
	}
}

type AWSErrorProcessor struct {
	log                  *clog.Logger
	missingPoliciesMutex sync.Mutex
	missingPolicies      map[string]struct{}
}

func (a *AWSErrorProcessor) Process(sr status.StatusReporter, err error) {
	var mp *errorhandler.MissingCSPPermissionError
	if errors.As(err, &mp) {
		a.log.Warn("missing permission error received, changing status to degraded")

		a.missingPoliciesMutex.Lock()
		defer a.missingPoliciesMutex.Unlock()
		if a.missingPolicies == nil {
			a.missingPolicies = map[string]struct{}{}
		}
		a.missingPolicies[mp.Permission] = struct{}{}
		policies := lo.Keys(a.missingPolicies)
		slices.Sort(policies)
		sr.UpdateStatus(
			status.Degraded,
			fmt.Sprintf(awsErrorProcessorDegradedStatusMessageFMT, strings.Join(policies, " , ")),
		)
	}
}

func (a *AWSErrorProcessor) Clear() {
	a.missingPoliciesMutex.Lock()
	defer a.missingPoliciesMutex.Unlock()
	a.missingPolicies = map[string]struct{}{}
}

const awsErrorProcessorDegradedStatusMessageFMT = "missing permission on cloud provider side: %s"
