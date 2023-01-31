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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	rdsClient "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

func NewProvider(cfg aws.Config, log *logp.Logger, factory awslib.CrossRegionFactory[Client]) *Provider {
	f := func(cfg aws.Config) Client {
		return rdsClient.NewFromConfig(cfg)
	}

	fetcher := factory.NewMultiRegionClients(ec2.NewFromConfig(cfg), cfg, f, log)

	return &Provider{
		log:     log,
		fetcher: fetcher,
	}
}

func (p Provider) DescribeDBInstances(ctx context.Context) ([]awslib.AwsResource, error) {
	return p.fetcher.Fetch(func(c Client) ([]awslib.AwsResource, error) {
		var result []awslib.AwsResource
		dbInstances, err := c.DescribeDBInstances(ctx, &rdsClient.DescribeDBInstancesInput{})
		if err != nil {
			p.log.Errorf("Could not describe DB instances. Error: %v", err)
			return result, err
		}

		for _, dbInstance := range dbInstances.DBInstances {
			result = append(result, DBInstance{*dbInstance.DBInstanceIdentifier, *dbInstance.DBInstanceArn, dbInstance.StorageEncrypted, dbInstance.AutoMinorVersionUpgrade})
		}

		return result, nil
	})
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
