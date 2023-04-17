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

package aws

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/dataprovider/types"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	cloudAccountIdField   = "cloud.account.id"
	cloudAccountNameField = "cloud.account.name"
	cloudProviderField    = "cloud.provider"
	cloudRegionField      = "cloud.region"
	cloudProviderValue    = "aws"
)

type DataProvider struct {
	log         *logp.Logger
	accountId   string
	accountName string
}

func New(options ...Option) DataProvider {
	adp := DataProvider{}
	for _, opt := range options {
		opt(&adp)
	}
	return adp
}

func (a DataProvider) FetchData(_ string, id string) (types.Data, error) {
	return types.Data{
		ResourceID: id,
		VersionInfo: version.CloudbeatVersionInfo{
			Version: version.CloudbeatVersion(),
			Policy:  version.PolicyVersion(),
		},
	}, nil
}

func (a DataProvider) EnrichEvent(event *beat.Event, resMetadata fetching.ResourceMetadata) error {
	_, err := event.Fields.Put(cloudAccountIdField, a.accountId)
	if err != nil {
		return err
	}

	_, err = event.Fields.Put(cloudAccountNameField, a.accountName)
	if err != nil {
		return err
	}

	_, err = event.Fields.Put(cloudProviderField, cloudProviderValue)
	if err != nil {
		return err
	}

	_, err = event.Fields.Put(cloudRegionField, resMetadata.Region)
	if err != nil {
		return err
	}

	return nil
}
