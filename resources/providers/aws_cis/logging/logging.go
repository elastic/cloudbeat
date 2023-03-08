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

package logging

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Client interface {
	DescribeTrails(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log           *logp.Logger
	s3Provider    s3.S3
	trailProvider cloudtrail.TrailService
}

func NewProvider(
	log *logp.Logger,
	cfg aws.Config,
	multiRegionTrailClients map[string]cloudtrail.Client,
	multiRegionS3Clients map[string]s3.Client,
) *Provider {
	return &Provider{
		log:           log,
		s3Provider:    s3.NewProvider(cfg, log, multiRegionS3Clients),
		trailProvider: cloudtrail.NewProvider(cfg, log, multiRegionTrailClients),
	}
}
