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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartitionFromARN(t *testing.T) {
	tests := []struct {
		name string
		arn  string
		want string
	}{
		{
			name: "commercial assumed role",
			arn:  "arn:aws:sts::111122223333:assumed-role/cloudbeat-root/session",
			want: "aws",
		},
		{
			name: "govcloud assumed role",
			arn:  "arn:aws-us-gov:sts::111122223333:assumed-role/cloudbeat-asset-inventory-root/session",
			want: "aws-us-gov",
		},
		{
			name: "govcloud iam user",
			arn:  "arn:aws-us-gov:iam::111122223333:user/cloudbeat",
			want: "aws-us-gov",
		},
		{
			name: "china partition",
			arn:  "arn:aws-cn:sts::111122223333:assumed-role/cloudbeat-root/session",
			want: "aws-cn",
		},
		{
			name: "empty arn",
			arn:  "",
			want: DefaultPartition,
		},
		{
			name: "invalid arn",
			arn:  "not-an-arn",
			want: DefaultPartition,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PartitionFromARN(tt.arn))
		})
	}
}

func TestIAMRoleARN(t *testing.T) {
	tests := []struct {
		name      string
		partition string
		want      string
	}{
		{
			name:      "empty partition uses commercial",
			partition: "",
			want:      "arn:aws:iam::111122223333:role/cloudbeat-root",
		},
		{
			name:      "commercial",
			partition: DefaultPartition,
			want:      "arn:aws:iam::111122223333:role/cloudbeat-root",
		},
		{
			name:      "govcloud",
			partition: GovCloudPartition,
			want:      "arn:aws-us-gov:iam::111122223333:role/cloudbeat-root",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IAMRoleARN(tt.partition, "111122223333", "cloudbeat-root"))
		})
	}
}
