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
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Identity struct {
	Account *string
	Arn     *string
	UserId  *string
}

type IdentityProvider struct {
	client *sts.Client
}

type IdentityProviderGetter interface {
	GetIdentity(ctx context.Context) (*Identity, error)
}

func GetIdentityClient(cfg aws.Config) IdentityProviderGetter {
	svc := sts.New(cfg)

	return &IdentityProvider{
		client: svc,
	}
}

// GetIdentity This method will return your identity (Arn, user-id...)
func (provider IdentityProvider) GetIdentity(ctx context.Context) (*Identity, error) {
	input := &sts.GetCallerIdentityInput{}
	request := provider.client.GetCallerIdentityRequest(input)
	response, err := request.Send(ctx)
	if err != nil {
		return nil, err
	}

	identity := &Identity{
		Account: response.Account,
		UserId:  response.UserId,
		Arn:     response.Arn,
	}
	return identity, nil
}
