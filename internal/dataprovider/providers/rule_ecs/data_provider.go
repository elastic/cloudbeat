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

package rule_ecs

import (
	"errors"

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

const (
	ruleKey = "rule"
)

type DataProvider struct{}

func NewDataProvider() DataProvider {
	return DataProvider{}
}

func (dp DataProvider) EnrichEvent(event *beat.Event, _ fetching.ResourceMetadata) error {
	ruleRaw, ok := event.Fields[ruleKey]
	if !ok {
		return nil
	}
	rule, ok := ruleRaw.(evaluator.Rule)
	if !ok {
		return errors.New("could not cast rule to 'evaluator.Rule")
	}
	rule.UUID = rule.Id
	rule.Reference = rule.References
	event.Fields[ruleKey] = rule
	return nil
}
