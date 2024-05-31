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

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type TransitGatewayAttachmentInfo struct {
	TransitGatewayAttachment types.TransitGatewayAttachment `json:"transit_gateway_attachment"`
	awsAccount               string
	region                   string
}

func (v TransitGatewayAttachmentInfo) GetResourceArn() string {
	id := pointers.Deref(v.TransitGatewayAttachment.TransitGatewayAttachmentId)
	if id == "" {
		return ""
	}
	return fmt.Sprintf("arn:aws:ec2:%s:%s:transit-gateway-attachment/%s", v.region, v.awsAccount, id)
}

func (v TransitGatewayAttachmentInfo) GetResourceName() string {
	return pointers.Deref(v.TransitGatewayAttachment.TransitGatewayAttachmentId)
}

func (v TransitGatewayAttachmentInfo) GetResourceType() string {
	return fetching.TransitGatewayAttachment
}

func (v TransitGatewayAttachmentInfo) GetRegion() string {
	return v.region
}
