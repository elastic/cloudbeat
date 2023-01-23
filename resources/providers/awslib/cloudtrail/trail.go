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

package cloudtrail

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type TrailInfo struct {
	TrailARN                  string          `json:"trail_arn"`
	Name                      string          `json:"name"`
	EnableLogFileValidation   bool            `json:"enable_log_file_validation"`
	IsMultiRegion             bool            `json:"is_multi_region"`
	KMSKeyID                  string          `json:"kms_key_id"`
	CloudWatchLogsLogGroupArn string          `json:"cloud_watch_logs_log_group_arn"`
	IsLogging                 bool            `json:"is_logging"`
	BucketName                string          `json:"bucket_name"`
	SnsTopicARN               string          `json:"sns_topic_arn"`
	EventSelectors            []EventSelector `json:"event_selectors"`
}

type DataResource struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type EventSelector struct {
	DataResources []DataResource      `json:"data_resources"`
	ReadWriteType types.ReadWriteType `json:"read_write_type"`
}

func (t TrailInfo) GetResourceArn() string {
	return t.TrailARN
}

func (t TrailInfo) GetResourceName() string {
	return t.Name
}

func (t TrailInfo) GetResourceType() string {
	return fetching.TrailType
}
