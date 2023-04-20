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

package iam

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type AnalyzersForRegion struct {
	Analyzers []types.AnalyzerSummary
	Region    string
}

func (a AnalyzersForRegion) GetMetadata() (fetching.ResourceMetadata, error) {
	id := fmt.Sprintf("access-analyzers-for-%s", a.Region)
	return fetching.ResourceMetadata{
		ID:      id,
		Type:    fetching.CloudIdentity,
		SubType: fetching.RegionAccessAnalyzers,
		Name:    id,
	}, nil
}

func (a AnalyzersForRegion) GetData() any { return a }

func (a AnalyzersForRegion) GetElasticCommonData() any { return nil }

func (p Provider) GetAccessAnalyzers(ctx context.Context) ([]AnalyzersForRegion, error) {
	out, err := awslib.MultiRegionFetch(ctx, p.accessAnalyzerClients, getAccessAnalyzersForRegion)
	if err != nil {
		return nil, err
	}
	return out, err
}

func getAccessAnalyzersForRegion(ctx context.Context, region string, c AccessAnalyzer) (AnalyzersForRegion, error) {
	analyzers := make([]types.AnalyzerSummary, 0)

	input := &accessanalyzer.ListAnalyzersInput{}
	for {
		out, err := c.ListAnalyzers(ctx, input)
		if err != nil {
			return AnalyzersForRegion{}, err
		}
		analyzers = append(analyzers, out.Analyzers...)
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	return AnalyzersForRegion{
		Analyzers: analyzers,
		Region:    region,
	}, nil
}
