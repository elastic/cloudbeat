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

package dataprovider

import "github.com/elastic/beats/v7/libbeat/beat"

type ElasticCommonDataProvider interface {
	GetElasticCommonData() (map[string]any, error)
}

type Enricher struct {
	dataprovider ElasticCommonDataProvider
}

func NewEnricher(dataprovider ElasticCommonDataProvider) *Enricher {
	return &Enricher{
		dataprovider: dataprovider,
	}
}

func (e *Enricher) EnrichEvent(event *beat.Event) error {
	ecsData, err := e.dataprovider.GetElasticCommonData()
	if err != nil {
		return err
	}

	for k, v := range ecsData {
		_, err := event.PutValue(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
