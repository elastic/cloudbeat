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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

// ListRecords returns every resource record set across all hosted zones in the account as
// Asset Discovery resources. Each record set becomes a single asset.
func (p *Provider) ListRecords(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Route53 records")

	var results []awslib.AwsResource
	var lastErr error

	zonesInput := &route53.ListHostedZonesInput{}
	for {
		zones, err := p.client.ListHostedZones(ctx, zonesInput)
		if err != nil {
			return results, err
		}

		for _, zone := range zones.HostedZones {
			hostedZoneID := pointers.Deref(zone.Id)
			records, err := p.listRecordsForZone(ctx, hostedZoneID, cleanZoneID(hostedZoneID), pointers.Deref(zone.Name))
			if err != nil {
				p.log.Errorf("Could not list records for hosted zone %s: %v", hostedZoneID, err)
				lastErr = err
				continue
			}
			results = append(results, records...)
		}

		if !zones.IsTruncated || zones.NextMarker == nil {
			break
		}
		zonesInput.Marker = zones.NextMarker
	}

	p.log.Debugf("Fetched %d Route53 records", len(results))
	return results, lastErr
}

func (p *Provider) listRecordsForZone(ctx context.Context, hostedZoneID, zoneID, zoneName string) ([]awslib.AwsResource, error) {
	var results []awslib.AwsResource
	input := &route53.ListResourceRecordSetsInput{HostedZoneId: pointers.Ref(hostedZoneID)}
	for {
		output, err := p.client.ListResourceRecordSets(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, rrs := range output.ResourceRecordSets {
			results = append(results, newRecord(rrs, zoneID, zoneName))
		}
		if !output.IsTruncated {
			break
		}
		input.StartRecordName = output.NextRecordName
		input.StartRecordType = output.NextRecordType
		input.StartRecordIdentifier = output.NextRecordIdentifier
	}
	return results, nil
}
