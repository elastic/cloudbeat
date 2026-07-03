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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

// DescribeClusters returns every EKS cluster across all regions as Asset Discovery resources.
// EKS only exposes cluster names via ListClusters, so each is then described individually.
func (p *Provider) DescribeClusters(ctx context.Context) ([]awslib.AwsResource, error) {
	clusters, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		var names []string
		input := &eks.ListClustersInput{}
		for {
			output, err := c.ListClusters(ctx, input)
			if err != nil {
				return nil, err
			}
			names = append(names, output.Clusters...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}

		result := make([]awslib.AwsResource, 0, len(names))
		for _, name := range names {
			described, err := c.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: pointers.Ref(name)})
			if err != nil {
				p.log.Errorf("Could not describe EKS cluster %s: %v", name, err)
				continue
			}
			if described.Cluster == nil {
				continue
			}
			result = append(result, newCluster(*described.Cluster, region))
		}
		return result, nil
	})
	return lo.Flatten(clusters), err
}
