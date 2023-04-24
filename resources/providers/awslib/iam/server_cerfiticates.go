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

	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

func (p Provider) ListServerCertificates(ctx context.Context) (awslib.AwsResource, error) {

	p.log.Debug("IAMProvider.ListServerCertificates")

	var certificates []types.ServerCertificateMetadata
	input := &iamsdk.ListServerCertificatesInput{}

	for {
		certs, err := p.client.ListServerCertificates(ctx, input)
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, certs.ServerCertificateMetadataList...)
		if !certs.IsTruncated {
			break
		}

		input.Marker = certs.Marker
	}

	return &ServerCertificatesInfo{
		Certificates: certificates,
	}, nil
}

func (c ServerCertificatesInfo) GetResourceArn() string {
	return ""
}

func (c ServerCertificatesInfo) GetResourceName() string {
	return "account-iam-server-certificates"
}

func (c ServerCertificatesInfo) GetResourceType() string {
	return fetching.IAMServerCertificateType
}

func (p ServerCertificatesInfo) GetRegion() string {
	return awslib.GlobalRegion
}
