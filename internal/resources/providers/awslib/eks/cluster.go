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

package eks

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

// Cluster is a flattened, normalized view of an EKS cluster. One Cluster maps to one asset.
type Cluster struct {
	Name                  string            `json:"name"`
	Arn                   string            `json:"arn"`
	Status                string            `json:"status,omitempty"`
	Version               string            `json:"version,omitempty"`
	Endpoint              string            `json:"endpoint,omitempty"`
	RoleArn               string            `json:"role_arn,omitempty"`
	PlatformVersion       string            `json:"platform_version,omitempty"`
	EndpointPublicAccess  bool              `json:"endpoint_public_access"`
	EndpointPrivateAccess bool              `json:"endpoint_private_access"`
	Tags                  map[string]string `json:"tags,omitempty"`
	CreatedAt             *time.Time        `json:"created_at,omitempty"`

	region string
}

func newCluster(c types.Cluster, region string) Cluster {
	cluster := Cluster{
		Name:            pointers.Deref(c.Name),
		Arn:             pointers.Deref(c.Arn),
		Status:          string(c.Status),
		Version:         pointers.Deref(c.Version),
		Endpoint:        pointers.Deref(c.Endpoint),
		RoleArn:         pointers.Deref(c.RoleArn),
		PlatformVersion: pointers.Deref(c.PlatformVersion),
		Tags:            c.Tags,
		CreatedAt:       c.CreatedAt,
		region:          region,
	}
	if c.ResourcesVpcConfig != nil {
		cluster.EndpointPublicAccess = c.ResourcesVpcConfig.EndpointPublicAccess
		cluster.EndpointPrivateAccess = c.ResourcesVpcConfig.EndpointPrivateAccess
	}
	return cluster
}

func (c Cluster) GetResourceArn() string {
	return c.Arn
}

func (c Cluster) GetResourceName() string {
	return c.Name
}

func (c Cluster) GetResourceType() string {
	return fetching.EKSType
}

func (c Cluster) GetRegion() string {
	return c.region
}

// GetOwnerTag returns the value of the "Owner" tag (case-insensitive), if present.
func (c Cluster) GetOwnerTag() string {
	for k, v := range c.Tags {
		if strings.EqualFold(k, "owner") {
			return v
		}
	}
	return ""
}
