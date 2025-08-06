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

package configservice

import (
	"context"

	configSDK "github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

func (p *Provider) DescribeConfigRecorders(ctx context.Context) ([]awslib.AwsResource, error) {
	configs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) (awslib.AwsResource, error) {
		recorderList, err := c.DescribeConfigurationRecorders(ctx, nil)
		if err != nil {
			p.log.Errorf(ctx, "Error fetching AWS Config recorders: %v", err)
			return nil, err
		}

		var result []Recorder
		for _, recorder := range recorderList.ConfigurationRecorders {
			recorderStatus, err := c.DescribeConfigurationRecorderStatus(ctx, &configSDK.DescribeConfigurationRecorderStatusInput{
				ConfigurationRecorderNames: []string{*recorder.Name},
			})

			if err != nil {
				p.log.Error("Error fetching recorder status, recorder: %v , Error: %v:", recorder, err)
				return nil, err
			}
			result = append(result, Recorder{
				ConfigurationRecorder: recorder,
				Status:                recorderStatus.ConfigurationRecordersStatus,
			})
		}

		return Config{Recorders: result, region: region, accountId: p.awsAccountId}, nil
	})

	return configs, err
}
