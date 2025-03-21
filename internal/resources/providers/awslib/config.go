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
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	libbeataws "github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
)

func RetryableCodesOption(o *retry.StandardOptions) {
	o.Retryables = append(o.Retryables, retry.RetryableHTTPStatusCode{
		Codes: map[int]struct{}{
			http.StatusTooManyRequests: {},
		},
	})
}

func awsConfigRetrier() aws.Retryer {
	return retry.NewStandard(RetryableCodesOption)
}

func InitializeAWSConfig(cfg libbeataws.ConfigAWS) (*aws.Config, error) {
	awsConfig, err := libbeataws.InitializeAWSConfig(cfg)
	if err != nil {
		return nil, err
	}

	awsConfig.Retryer = awsConfigRetrier

	return &awsConfig, nil
}
