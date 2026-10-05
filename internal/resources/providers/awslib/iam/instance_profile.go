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
)

// InstanceProfileGetter resolves an IAM instance profile by name.
type InstanceProfileGetter interface {
	GetInstanceProfile(ctx context.Context, instanceProfileName string) (*types.InstanceProfile, error)
}

// GetInstanceProfile fetches the IAM instance profile with the given name and
// returns the profile object, which includes the list of attached roles.
func (p Provider) GetInstanceProfile(ctx context.Context, instanceProfileName string) (*types.InstanceProfile, error) {
	input := &iam.GetInstanceProfileInput{
		InstanceProfileName: &instanceProfileName,
	}

	response, err := p.client.GetInstanceProfile(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance profile %s - %w", instanceProfileName, err)
	}

	if response.InstanceProfile == nil {
		return nil, fmt.Errorf("GetInstanceProfile returned nil for %s", instanceProfileName)
	}

	return response.InstanceProfile, nil
}
