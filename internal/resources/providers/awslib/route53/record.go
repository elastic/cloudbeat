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

package route53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

// Record is a flattened, normalized view of a Route53 resource record set together with the
// hosted zone it belongs to. One Record maps to one Asset Discovery asset.
type Record struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Records           []string `json:"records,omitempty"`
	Weight            *int64   `json:"weight,omitempty"`
	RoutingRegion     string   `json:"routing_region,omitempty"`
	ZoneID            string   `json:"zone_id"`
	ZoneName          string   `json:"zone_name"`
	AliasTargetDNS    string   `json:"alias_target_dns,omitempty"`
	AliasTargetZoneID string   `json:"alias_target_zone_id,omitempty"`
	HealthCheckID     string   `json:"health_check_id,omitempty"`

	setIdentifier string
}

func newRecord(rrs types.ResourceRecordSet, zoneID, zoneName string) Record {
	r := Record{
		Name:          pointers.Deref(rrs.Name),
		Type:          string(rrs.Type),
		Weight:        rrs.Weight,
		RoutingRegion: string(rrs.Region),
		ZoneID:        zoneID,
		ZoneName:      zoneName,
		HealthCheckID: pointers.Deref(rrs.HealthCheckId),
		setIdentifier: pointers.Deref(rrs.SetIdentifier),
	}
	for _, rec := range rrs.ResourceRecords {
		if v := pointers.Deref(rec.Value); v != "" {
			r.Records = append(r.Records, v)
		}
	}
	if rrs.AliasTarget != nil {
		r.AliasTargetDNS = pointers.Deref(rrs.AliasTarget.DNSName)
		r.AliasTargetZoneID = pointers.Deref(rrs.AliasTarget.HostedZoneId)
	}
	return r
}

// GetResourceArn synthesizes a stable identifier for the record set. Route53 record sets have
// no ARN, so we derive one from the zone, record name, type and (for weighted/latency/geo
// records) the set identifier, which together uniquely identify a record set.
func (r Record) GetResourceArn() string {
	arn := fmt.Sprintf("arn:aws:route53:::hostedzone/%s/recordset/%s/%s", r.ZoneID, r.Name, r.Type)
	if r.setIdentifier != "" {
		arn += "/" + r.setIdentifier
	}
	return arn
}

func (r Record) GetResourceName() string {
	return r.Name
}

func (r Record) GetResourceType() string {
	return fetching.Route53Type
}

// GetRegion reports the AWS region. Route53 is a global service.
func (r Record) GetRegion() string {
	return awslib.GlobalRegion
}

// cleanZoneID strips the "/hostedzone/" prefix Route53 returns on hosted zone IDs.
func cleanZoneID(id string) string {
	return strings.TrimPrefix(id, "/hostedzone/")
}
