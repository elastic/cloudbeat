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

package kms

import (
	"context"

	kmsClient "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Provider struct {
	log     *clog.Logger
	clients map[string]Client
}

type Client interface {
	ListKeys(ctx context.Context, params *kmsClient.ListKeysInput, optFns ...func(*kmsClient.Options)) (*kmsClient.ListKeysOutput, error)
	DescribeKey(ctx context.Context, params *kmsClient.DescribeKeyInput, optFns ...func(*kmsClient.Options)) (*kmsClient.DescribeKeyOutput, error)
	GetKeyRotationStatus(ctx context.Context, params *kmsClient.GetKeyRotationStatusInput, optFns ...func(*kmsClient.Options)) (*kmsClient.GetKeyRotationStatusOutput, error)
}

func (p *Provider) DescribeSymmetricKeys(ctx context.Context) ([]awslib.AwsResource, error) {
	symmetricKeys, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var kmsKeys []types.KeyListEntry
		input := &kmsClient.ListKeysInput{}
		for {
			output, err := c.ListKeys(ctx, input)
			if err != nil {
				return nil, err
			}
			kmsKeys = append(kmsKeys, output.Keys...)
			if !output.Truncated {
				break
			}
			input.Marker = output.NextMarker
		}

		var result []awslib.AwsResource
		for _, keyEntry := range kmsKeys {
			keyInfo, err := c.DescribeKey(ctx, &kmsClient.DescribeKeyInput{
				KeyId: keyEntry.KeyId,
			})
			if err != nil {
				p.log.Error(err.Error())
				continue
			}

			if keyInfo.KeyMetadata.KeySpec != types.KeySpecSymmetricDefault {
				continue
			}

			if keyInfo.KeyMetadata.KeyManager != types.KeyManagerTypeCustomer {
				continue
			}

			rotationStatus, err := c.GetKeyRotationStatus(ctx, &kmsClient.GetKeyRotationStatusInput{
				KeyId: keyEntry.KeyId,
			})
			if err != nil {
				p.log.Error(err.Error())
				continue
			}

			result = append(result, KmsInfo{
				KeyMetadata:        *keyInfo.KeyMetadata,
				KeyRotationEnabled: rotationStatus.KeyRotationEnabled,
				region:             region,
			})
		}
		return result, nil
	})

	return lo.Flatten(symmetricKeys), err
}

func (k KmsInfo) GetResourceArn() string {
	if k.KeyMetadata.Arn == nil {
		return ""
	}
	return *k.KeyMetadata.Arn
}

func (k KmsInfo) GetResourceName() string {
	if k.KeyMetadata.KeyId == nil {
		return ""
	}
	return *k.KeyMetadata.KeyId
}

func (k KmsInfo) GetResourceType() string {
	return fetching.KmsType
}

func (k KmsInfo) GetRegion() string {
	return k.region
}
