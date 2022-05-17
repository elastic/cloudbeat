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

package conditions

import (
	"fmt"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type ErrorCondition struct {
	err error
	factoryName string
}

func NewErrorCondition(factoryName string, err error) fetching.Condition {
	return &ErrorCondition{
		err: err,
		factoryName: factoryName,
	}
}

func (c *ErrorCondition) Condition() bool {
	logp.L().Error(fmt.Errorf("ErrorCondition of fetcher %s failed due to: %v", c.factoryName, c.err))
	return false
}

func (c *ErrorCondition) Name() string {
	return "error_condition"
}
