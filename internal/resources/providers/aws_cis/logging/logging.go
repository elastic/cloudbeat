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

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type Client interface {
	DescribeTrails(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log           *clog.Logger
	s3Provider    s3.S3
	trailProvider cloudtrail.TrailService
}

func NewProvider(
	ctx context.Context,
	log *clog.Logger,
	cfg aws.Config,
	multiRegionTrailFactory awslib.CrossRegionFactory[cloudtrail.Client],
	multiRegionS3Factory awslib.CrossRegionFactory[s3.Client],
	accountId string,
) *Provider {
	return &Provider{
		log:           log,
		s3Provider:    s3.NewProvider(ctx, log, cfg, multiRegionS3Factory, accountId),
		trailProvider: cloudtrail.NewProvider(ctx, log, cfg, multiRegionTrailFactory),
	}
}
