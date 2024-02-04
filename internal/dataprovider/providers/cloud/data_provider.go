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
	"errors"

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/strings"
)

const (
	cloudAccountIdField   = "cloud.account.id"
	cloudAccountNameField = "cloud.account.name"
	cloudProviderField    = "cloud.provider"
	cloudRegionField      = "cloud.region"
	// TODO: update fields names when an ECS field is decided
	cloudOrganizationIdField   = "cloud.Organization.id"
	cloudOrganizationNameField = "cloud.Organization.name"
)

type Identity struct {
	Provider         string
	Account          string
	AccountAlias     string
	OrganizationId   string
	OrganizationName string
}

type DataProvider struct {
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
	return errors.Join(
		// always insert cloud.account.id as missing fields can crash the dashboard
		insert(cloudAccountIdField, strings.FirstNonEmpty(resMetadata.AccountId, a.accountId), event),
		insertIfNotEmpty(cloudAccountNameField, strings.FirstNonEmpty(resMetadata.AccountName, a.accountName), event),
		insertIfNotEmpty(cloudProviderField, a.providerName, event),
		insertIfNotEmpty(cloudRegionField, resMetadata.Region, event),
		insertIfNotEmpty(cloudOrganizationIdField, resMetadata.OrganisationId, event),
		insertIfNotEmpty(cloudOrganizationNameField, resMetadata.OrganizationName, event),
	)
}

func insert(field string, value string, event *beat.Event) error {
	_, err := event.Fields.Put(field, value)
	return err
}

func insertIfNotEmpty(field string, value string, event *beat.Event) error {
	if value != "" {
		_, err := event.Fields.Put(field, value)
		return err
	}
	return nil
}
