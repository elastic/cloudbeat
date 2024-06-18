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

package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Provider struct {
	log     *logp.Logger
	clients map[string]Client
}

type Client interface {
	lambda.ListAliasesAPIClient
	lambda.ListEventSourceMappingsAPIClient
	lambda.ListFunctionsAPIClient
	lambda.ListLayersAPIClient
}

func (p *Provider) ListFunctions(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Lambda Functions")
	funcs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &lambda.ListFunctionsInput{}
		all := []types.FunctionConfiguration{}
		for {
			output, err := c.ListFunctions(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.Functions...)
			if output.NextMarker == nil {
				break
			}
			input.Marker = output.NextMarker
		}

		var result []awslib.AwsResource
		for _, item := range all {
			f := &FunctionInfo{
				Function: item,
				region:   region,
			}
			aliases, err := p.ListAliases(ctx, region, f.GetResourceArn())
			if err != nil {
				p.log.Warnf("error listing aliases: %s", err)
			} else {
				f.Aliases = aliases
			}
			result = append(result, f)
		}
		return result, nil
	})
	result := lo.Flatten(funcs)
	if err != nil {
		p.log.Debugf("Fetched %d Lambda Functions", len(result))
	}
	return result, err
}

func (p *Provider) ListAliases(ctx context.Context, region, functionName string) ([]awslib.AwsResource, error) {
	p.log.Debugf("Fetching Lambda Aliases for %s function in %s", functionName, region)
	c, ok := p.clients[region]
	if !ok {
		return nil, fmt.Errorf("failed to get a client for %s region", region)
	}
	input := &lambda.ListAliasesInput{
		FunctionName: &functionName,
	}

	var result []awslib.AwsResource
	for {
		output, err := c.ListAliases(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, item := range output.Aliases {
			f := &AliasInfo{
				Alias:  item,
				region: region,
			}
			result = append(result, f)
		}
		if output.NextMarker == nil {
			break
		}
		input.Marker = output.NextMarker
	}

	p.log.Debugf("Fetched %d Lambda Aliases for %s in %s", len(result), functionName, region)
	return result, nil
}

func (p *Provider) ListEventSourceMappings(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Lambda Event Source Mappings")
	funcs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &lambda.ListEventSourceMappingsInput{}
		all := []types.EventSourceMappingConfiguration{}
		for {
			output, err := c.ListEventSourceMappings(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.EventSourceMappings...)
			if output.NextMarker == nil {
				break
			}
			input.Marker = output.NextMarker
		}

		var result []awslib.AwsResource
		for _, item := range all {
			f := &EventSourceMappingInfo{
				EventSourceMapping: item,
				region:             region,
			}
			result = append(result, f)
		}
		return result, nil
	})
	result := lo.Flatten(funcs)
	if err != nil {
		p.log.Debugf("Fetched %d Lambda Event Source Mappings", len(result))
	}
	return result, err
}

func (p *Provider) ListLayers(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Lambda Layers")
	funcs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &lambda.ListLayersInput{}
		all := []types.LayersListItem{}
		for {
			output, err := c.ListLayers(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.Layers...)
			if output.NextMarker == nil {
				break
			}
			input.Marker = output.NextMarker
		}

		var result []awslib.AwsResource
		for _, item := range all {
			f := &LayerInfo{
				Layer:  item,
				region: region,
			}
			result = append(result, f)
		}
		return result, nil
	})
	result := lo.Flatten(funcs)
	if err != nil {
		p.log.Debugf("Fetched %d Lambda Layers", len(result))
	}
	return result, err
}
