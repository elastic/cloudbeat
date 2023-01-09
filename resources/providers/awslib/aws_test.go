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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

var successfulOutput = &ec2.DescribeRegionsOutput{
	Regions: []types.Region{
		{
			RegionName: aws.String("us-east-1"),
		},
		{
			RegionName: aws.String("eu-west-1"),
		},
	},
}

func TestGetRegions(t *testing.T) {
	type args struct {
		client func() AWSCommonUtil
		log    *logp.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Error - return no regions",
			args: args{
				client: func() AWSCommonUtil {
					m := &MockAWSCommonUtil{}
					m.On("DescribeRegions", mock.Anything, mock.Anything).Return(nil, errors.New("fail to query endpoint"))
					return m
				},
				log: logp.NewLogger("aws-test"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return enabled regions",
			args: args{
				client: func() AWSCommonUtil {
					m := &MockAWSCommonUtil{}
					m.On("DescribeRegions", mock.Anything, mock.Anything).Return(successfulOutput, nil)
					return m
				},
				log: logp.NewLogger("aws-test"),
			},
			want:    []string{"us-east-1", "eu-west-1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRegions(tt.args.log, tt.args.client())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
