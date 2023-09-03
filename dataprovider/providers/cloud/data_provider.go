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

package cloud

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/strings"
)

const (
	cloudAccountIdField   = "cloud.account.id"
	cloudAccountNameField = "cloud.account.name"
	cloudProviderField    = "cloud.provider"
	cloudRegionField      = "cloud.region"
)

type Identity struct {
	Provider     string
	Account      string
	AccountAlias string
}

type DataProvider struct {
	log          *logp.Logger
	accountId    string
	accountName  string
	providerName string
}

func NewDataProvider(options ...Option) DataProvider {
	adp := DataProvider{}
	for _, opt := range options {
		opt(&adp)
	}
	return adp
}

func (a DataProvider) EnrichEvent(event *beat.Event, resMetadata fetching.ResourceMetadata) error {
	err := insertIfNotEmpty(cloudAccountIdField, strings.FirstNonEmpty(resMetadata.AwsAccountId, a.accountId), event)
	if err != nil {
		return err
	}

	err = insertIfNotEmpty(cloudAccountNameField, strings.FirstNonEmpty(resMetadata.AwsAccountAlias, a.accountName), event)
	if err != nil {
		return err
	}

	err = insertIfNotEmpty(cloudProviderField, a.providerName, event)
	if err != nil {
		return err
	}

	err = insertIfNotEmpty(cloudRegionField, resMetadata.Region, event)
	if err != nil {
		return err
	}

	return nil
}

func insertIfNotEmpty(field string, value string, event *beat.Event) error {
	if value != "" {
		_, err := event.Fields.Put(field, value)
		return err
	}
	return nil
}
