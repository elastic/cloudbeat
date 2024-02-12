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

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type AccessAnalyzer struct {
	types.AnalyzerSummary
	Region string
}

type AccessAnalyzers struct {
	Analyzers []AccessAnalyzer
	Regions   []string
}

func (a AccessAnalyzers) GetResourceArn() string {
	return ""
}

func (a AccessAnalyzers) GetResourceName() string {
	return "account-access-analyzers"
}

func (a AccessAnalyzers) GetResourceType() string {
	return fetching.AccessAnalyzers
}

func (a AccessAnalyzers) GetRegion() string {
	return awslib.GlobalRegion
}

type analyzersForRegion struct {
	analyzers  []types.AnalyzerSummary
	regionName string
}

func (p Provider) GetAccessAnalyzers(ctx context.Context) (awslib.AwsResource, error) {
	analyzers, err := awslib.MultiRegionFetch(ctx, p.accessAnalyzerClients, getAccessAnalyzersForRegion)
	if err != nil {
		return nil, err
	}

	accessAnalyzers := make([]AccessAnalyzer, 0)
	regions := make([]string, 0)
	for _, region := range analyzers {
		regions = append(regions, region.regionName)
		for _, analyzer := range region.analyzers {
			accessAnalyzers = append(accessAnalyzers, AccessAnalyzer{
				AnalyzerSummary: analyzer,
				Region:          region.regionName,
			})
		}
	}

	return AccessAnalyzers{
		Analyzers: accessAnalyzers,
		Regions:   regions,
	}, err
}

func getAccessAnalyzersForRegion(ctx context.Context, region string, c AccessAnalyzerClient) (analyzersForRegion, error) {
	analyzers := make([]types.AnalyzerSummary, 0)

	input := &accessanalyzer.ListAnalyzersInput{}
	for {
		out, err := c.ListAnalyzers(ctx, input)
		if err != nil {
			return analyzersForRegion{}, err
		}
		analyzers = append(analyzers, out.Analyzers...)
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	return analyzersForRegion{
		analyzers:  analyzers,
		regionName: region,
	}, nil
}
