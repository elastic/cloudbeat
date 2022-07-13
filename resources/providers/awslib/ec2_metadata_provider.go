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
	"github.com/aws/aws-sdk-go-v2/aws/ec2metadata"
)

type Ec2MetadataProvider struct {
}

type Metadata = ec2metadata.EC2InstanceIdentityDocument

type MetadataGetter interface {
	GetMetadata(ctx context.Context, cfg aws.Config) (Metadata, error)
}

func (provider Ec2MetadataProvider) GetMetadata(ctx context.Context, cfg aws.Config) (Metadata, error) {
	svc := ec2metadata.New(cfg)
	identityDocument, err := svc.GetInstanceIdentityDocument(ctx)
	if err != nil {
		return ec2metadata.EC2InstanceIdentityDocument{}, err
	}

	return identityDocument, err
}
