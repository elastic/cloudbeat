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

package inventory

import (
	"context"
	"fmt"
	"sync"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"

	gcplib "github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

type Provider struct {
	log    *logp.Logger
	client *GcpClientWrapper
	ctx    context.Context
	Config gcplib.GcpFactoryConfig
}

type Iterator interface {
	Next() (*assetpb.Asset, error)
}

type GcpClientWrapper struct {
	Close      func() error
	ListAssets func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator
}

type InventoryService interface {
	// List all content types of the given assets types
	ListAllAssetTypesByName(assets []string) ([]*assetpb.Asset, error)

	// Close the GCP asset client
	Close() error
}

func NewAssetsInventoryProvider(ctx context.Context, log *logp.Logger, gcpConfig gcplib.GcpFactoryConfig) (*Provider, error) {
	client, err := asset.NewClient(ctx, gcpConfig.ClientOpts...)

	if err != nil {
		return nil, err
	}

	// We wrap the client so we can mock it in tests
	wrapper := &GcpClientWrapper{
		Close: client.Close,
		ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
			return client.ListAssets(ctx, req, opts...)
		},
	}

	return &Provider{
		log:    log,
		client: wrapper,
		ctx:    ctx,
		Config: gcpConfig,
	}, nil
}

func (p *Provider) ListAllAssetTypesByName(assets []string) ([]*assetpb.Asset, error) {
	p.log.Infof("Listing GCP assets: %v", assets)

	wg := sync.WaitGroup{}
	scope := fmt.Sprintf("projects/%s", p.Config.ProjectId)
	var resourceAssets []*assetpb.Asset
	var policyAssets []*assetpb.Asset

	wg.Add(1)
	go func() {
		request := &assetpb.ListAssetsRequest{
			Parent:      scope,
			AssetTypes:  assets,
			ContentType: assetpb.ContentType_RESOURCE,
		}
		resourceAssets = getAllAssets(p.log, p.client.ListAssets(p.ctx, request))
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		request := &assetpb.ListAssetsRequest{
			Parent:      scope,
			AssetTypes:  assets,
			ContentType: assetpb.ContentType_IAM_POLICY,
		}
		policyAssets = getAllAssets(p.log, p.client.ListAssets(p.ctx, request))
		wg.Done()
	}()

	wg.Wait()

	results := append(resourceAssets, policyAssets...)
	return mergeAssetContentType(results), nil
}

func (p *Provider) Close() error {
	return p.client.Close()
}

func getAllAssets(log *logp.Logger, it Iterator) []*assetpb.Asset {
	results := make([]*assetpb.Asset, 0)

	for {
		response, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Errorf("Error fetching GCP Asset: %s", err)
			return nil
		}

		log.Debugf("Fetched GCP Asset: %+v", response.Name)
		results = append(results, response)
	}

	return results
}

func mergeAssetContentType(assets []*assetpb.Asset) []*assetpb.Asset {
	var resultsMap = make(map[string]*assetpb.Asset)

	for _, asset := range assets {
		assetKey := asset.Name
		if _, ok := resultsMap[assetKey]; !ok {
			resultsMap[assetKey] = asset
			continue
		}

		item := resultsMap[assetKey]
		if asset.Resource != nil {
			item.Resource = asset.Resource
		}
		if asset.IamPolicy != nil {
			item.IamPolicy = asset.IamPolicy
		}
	}

	var results []*assetpb.Asset
	for _, asset := range resultsMap {
		results = append(results, asset)
	}
	return results
}
