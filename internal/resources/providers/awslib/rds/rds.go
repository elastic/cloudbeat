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

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
)

type DBInstance struct {
	Identifier              string   `json:"identifier"`
	Arn                     string   `json:"arn"`
	StorageEncrypted        bool     `json:"storage_encrypted"`
	AutoMinorVersionUpgrade bool     `json:"auto_minor_version_upgrade"`
	PubliclyAccessible      bool     `json:"publicly_accessible"`
	Subnets                 []Subnet `json:"subnets"`
	region                  string
}

type Subnet struct {
	ID         string
	RouteTable *RouteTable
}

type RouteTable struct {
	ID     string
	Routes []Route
}

type Route struct {
	DestinationCidrBlock *string
	GatewayId            *string
}

type Rds interface {
	DescribeDBInstances(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log     *clog.Logger
	clients map[string]Client
	ec2     ec2.ElasticCompute
}

type Client interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}
