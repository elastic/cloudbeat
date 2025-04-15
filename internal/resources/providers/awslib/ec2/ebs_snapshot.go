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

package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

type EBSSnapshot struct {
	Instance    Ec2Instance
	SnapshotId  string
	State       types.SnapshotState
	Region      string
	awsAccount  string
	VolumeSize  int
	IsEncrypted bool
}

func (e EBSSnapshot) GetResourceArn() string {
	// TODO: check if this is the correct ARN
	return fmt.Sprintf("arn:aws:ec2:%s:%s:ec2/%s", e.Region, e.awsAccount, e.SnapshotId)
}

func (e EBSSnapshot) GetResourceName() string {
	// TODO: From tags?
	return fmt.Sprintf("ebs-snapshot-by-default-%s-%s", e.awsAccount, e.Region)
}

func (e EBSSnapshot) GetResourceType() string {
	return fetching.EBSSnapshotType
}

func FromSnapshotInfo(snapshot types.SnapshotInfo, region string, awsAccount string, ins Ec2Instance) EBSSnapshot {
	return EBSSnapshot{
		Instance:    ins,
		SnapshotId:  *snapshot.SnapshotId,
		State:       snapshot.State,
		Region:      region,
		awsAccount:  awsAccount,
		VolumeSize:  int(aws.ToInt32(snapshot.VolumeSize)),
		IsEncrypted: aws.ToBool(snapshot.Encrypted),
	}
}

func FromSnapshot(snapshot types.Snapshot, region string, awsAccount string, ins Ec2Instance) EBSSnapshot {
	return EBSSnapshot{
		SnapshotId:  *snapshot.SnapshotId,
		State:       snapshot.State,
		Region:      region,
		awsAccount:  awsAccount,
		VolumeSize:  int(aws.ToInt32(snapshot.VolumeSize)),
		Instance:    ins,
		IsEncrypted: aws.ToBool(snapshot.Encrypted),
	}
}
