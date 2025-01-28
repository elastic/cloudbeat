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

package awslib

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type currentRegionSelector struct {
	client MetadataProvider
}

func (s *currentRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := clog.NewLogger("aws")
	log.Info("Getting current region of the instance")

	if s.client == nil {
		s.client = &Ec2MetadataProvider{}
	}

	metadata, err := s.client.GetMetadata(ctx, cfg)
	if err != nil {
		log.Errorf("Failed getting current region: %v", err)
		return nil, err
	}

	log.Infof("Current region of aws instance, %v", metadata.Region)
	return []string{metadata.Region}, nil
}
