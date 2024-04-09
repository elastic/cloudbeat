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

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/samber/lo"
)

type RoleGetter interface {
	GetRole(ctx context.Context, roleName string) (*Role, error)
}

func (p Provider) GetRole(ctx context.Context, roleName string) (*Role, error) {
	input := &iam.GetRoleInput{
		RoleName: &roleName,
	}

	response, err := p.client.GetRole(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get role %s - %w", roleName, err)
	}

	r := &Role{
		Role: *response.Role,
	}

	return r, nil
}

func (p Provider) ListRoles(ctx context.Context) ([]*Role, error) {
	input := &iam.ListRolesInput{}

	roles := make([]types.Role, 0)
	for {
		nativeRoles, err := p.client.ListRoles(ctx, input)
		if err != nil {
			return nil, err
		}

		roles = append(roles, nativeRoles.Roles...)

		if !nativeRoles.IsTruncated {
			break
		}

		input.Marker = nativeRoles.Marker
	}

	return lo.Map(roles, func(role types.Role, _ int) *Role {
		return &Role{
			Role: role,
		}
	}), nil
}
