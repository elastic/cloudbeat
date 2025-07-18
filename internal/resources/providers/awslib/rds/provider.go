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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	ec2Provider "github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
)

func NewProvider(ctx context.Context, log *clog.Logger, cfg aws.Config, factory awslib.CrossRegionFactory[Client], ec2Compute ec2Provider.ElasticCompute) *Provider {
	f := func(cfg aws.Config) Client {
		return rds.NewFromConfig(cfg)
	}
	m := factory.NewMultiRegionClients(ctx, awslib.AllRegionSelector(), cfg, f, log)
	return &Provider{
		log:     log,
		clients: m.GetMultiRegionsClientMap(),
		ec2:     ec2Compute,
	}
}

func (p Provider) DescribeDBInstances(ctx context.Context) ([]awslib.AwsResource, error) {
	rdss, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var result []awslib.AwsResource
		var dbInstances []types.DBInstance
		dbInstancesInput := &rds.DescribeDBInstancesInput{}

		for {
			output, err := c.DescribeDBInstances(ctx, dbInstancesInput)
			if err != nil {
				p.log.Errorf("Could not describe DB instances. Error: %v", err)
				return result, err
			}

			dbInstances = append(dbInstances, output.DBInstances...)
			if output.Marker == nil {
				break
			}

			dbInstancesInput.Marker = output.Marker
		}

		for _, dbInstance := range dbInstances {
			subnets := p.getDBInstanceSubnets(ctx, region, dbInstance)
			result = append(result, DBInstance{
				Identifier:              *dbInstance.DBInstanceIdentifier,
				Arn:                     *dbInstance.DBInstanceArn,
				StorageEncrypted:        aws.ToBool(dbInstance.StorageEncrypted),
				AutoMinorVersionUpgrade: aws.ToBool(dbInstance.AutoMinorVersionUpgrade),
				PubliclyAccessible:      aws.ToBool(dbInstance.PubliclyAccessible),
				Subnets:                 subnets,
				region:                  region,
			})
		}

		return result, nil
	})

	return lo.Flatten(rdss), err
}

func (p Provider) getDBInstanceSubnets(ctx context.Context, region string, dbInstance types.DBInstance) []Subnet {
	results := make([]Subnet, 0, len(dbInstance.DBSubnetGroup.Subnets))
	for _, subnet := range dbInstance.DBSubnetGroup.Subnets {
		resultSubnet := Subnet{ID: *subnet.SubnetIdentifier, RouteTable: nil}
		routeTableForSubnet, err := p.ec2.GetRouteTableForSubnet(ctx, region, *subnet.SubnetIdentifier, *dbInstance.DBSubnetGroup.VpcId)
		if err != nil {
			p.log.Errorf("Could not get route table for subnet %s of DB %s. Error: %v", *subnet.SubnetIdentifier, *dbInstance.DBInstanceIdentifier, err)
		} else {
			var routes []Route
			for _, route := range routeTableForSubnet.Routes {
				routes = append(routes, Route{DestinationCidrBlock: route.DestinationCidrBlock, GatewayId: route.GatewayId})
			}

			resultSubnet.RouteTable = &RouteTable{ID: *routeTableForSubnet.RouteTableId, Routes: routes}
		}

		results = append(results, resultSubnet)
	}

	return results
}

func (d DBInstance) GetResourceArn() string {
	return d.Arn
}

func (d DBInstance) GetResourceName() string {
	return d.Identifier
}

func (d DBInstance) GetResourceType() string {
	return fetching.RdsType
}

func (d DBInstance) GetRegion() string {
	return d.region
}
