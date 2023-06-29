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
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type AWS struct{}

func (A *AWS) Run(context.Context) error { return nil }

func (A *AWS) InitRegistry(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, dependencies *Dependencies) (registry.Registry, error) {
	awsConfig, awsIdentity, err := getCisAwsConfig(ctx, cfg, dependencies)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS config: %w", err)
	}

	fm, err := factory.NewCisAwsFactory(log, awsConfig, ch, awsIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to create EKS: %w", err)
	}
	return registry.NewRegistry(log, fm), nil
}

func (A *AWS) Stop() {}

func getCisAwsConfig(ctx context.Context, cfg *config.Config, dependencies *Dependencies) (awssdk.Config, *awslib.Identity, error) {
	// Initialize AWS config with the default region rather than the ec2 region.
	// This is because in CSPM we create a client per region.
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identity, err := dependencies.AWSIdentity(ctx, awsConfig)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	return awsConfig, identity, err
}
