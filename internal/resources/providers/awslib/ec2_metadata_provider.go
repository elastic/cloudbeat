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
	ec2imds "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

type Ec2Metadata = ec2imds.InstanceIdentityDocument

type Ec2MetadataProvider struct{}

type MetadataProvider interface {
	GetMetadata(ctx context.Context, cfg aws.Config) (*Ec2Metadata, error)
}

func (provider Ec2MetadataProvider) GetMetadata(ctx context.Context, cfg aws.Config) (*Ec2Metadata, error) {
	svc := ec2imds.NewFromConfig(cfg)
	input := &ec2imds.GetInstanceIdentityDocumentInput{}
	// this call will fail running from local machine
	// TODO: mock local struct
	identityDocument, err := svc.GetInstanceIdentityDocument(ctx, input)
	if err != nil {
		return nil, err
	}

	return &identityDocument.InstanceIdentityDocument, nil
}
