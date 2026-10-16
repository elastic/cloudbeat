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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

const (
	// DefaultPartition is the AWS commercial partition.
	DefaultPartition = "aws"
	// GovCloudPartition is the AWS GovCloud (US) partition.
	GovCloudPartition = "aws-us-gov"
)

// PartitionFromARN returns the partition of an ARN, for example "aws" or "aws-us-gov".
// It returns DefaultPartition if the ARN is empty or not valid.
func PartitionFromARN(resourceARN string) string {
	parsed, err := arn.Parse(resourceARN)
	if err != nil || parsed.Partition == "" {
		return DefaultPartition
	}
	return parsed.Partition
}

// IAMRoleARN returns the ARN of an IAM role in the given partition.
// An empty partition means DefaultPartition.
func IAMRoleARN(partition string, accountID string, roleName string) string {
	if partition == "" {
		partition = DefaultPartition
	}
	return fmt.Sprintf("arn:%s:iam::%s:role/%s", partition, accountID, roleName)
}
