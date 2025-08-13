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

package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type KmsFetcher struct {
	log           *clog.Logger
	kms           kms.KMS
	resourceCh    chan fetching.ResourceInfo
	statusHandler statushandler.StatusHandlerAPI
}

type KmsResource struct {
	key awslib.AwsResource
}

func NewKMSFetcher(log *clog.Logger, provider kms.KMS, ch chan fetching.ResourceInfo, statusHandler statushandler.StatusHandlerAPI) *KmsFetcher {
	return &KmsFetcher{
		log:           log,
		kms:           provider,
		resourceCh:    ch,
		statusHandler: statusHandler,
	}
}

func (f *KmsFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Info("Starting KMSFetcher.Fetch")

	keys, err := f.kms.DescribeSymmetricKeys(ctx)
	if err != nil {
		f.log.Errorf("failed to describe keys from KMS: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
		return nil
	}

	for _, key := range keys {
		resource := KmsResource{key}
		f.log.Debugf("Fetched key: %s", key.GetResourceName())
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      resource,
			CycleMetadata: cycleMetadata,
		}
	}

	return nil
}

func (f *KmsFetcher) Stop() {}

func (r KmsResource) GetData() any {
	return r.key
}

func (r KmsResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.key.GetResourceArn(),
		Type:    fetching.KeyManagement,
		SubType: r.key.GetResourceType(),
		Name:    r.key.GetResourceName(),
		Region:  r.key.GetRegion(),
	}, nil
}

func (r KmsResource) GetIds() []string {
	return []string{r.key.GetResourceArn()}
}

func (r KmsResource) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{
		"cloud.service.name": "KMS",
	}

	key, ok := r.key.(kms.KmsInfo)
	if ok {
		m["x509.not_after"] = key.KeyMetadata.ValidTo
		m["x509.not_before"] = key.KeyMetadata.CreationDate
		switch key.KeyMetadata.KeyUsage {
		case types.KeyUsageTypeSignVerify:
			m["x509.signature_algorithm"] = key.KeyMetadata.KeySpec
		case types.KeyUsageTypeEncryptDecrypt:
			m["x509.public_key_algorithm"] = key.KeyMetadata.KeySpec
		default:
		}
	}

	return m, nil
}
