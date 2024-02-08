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

package maps

import (
	"encoding/json"
	"fmt"
)

func AsMapStringAny(item any) (map[string]any, error) {
	js, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %T to json: %w", item, err)
	}

	m := map[string]any{}
	err = json.Unmarshal(js, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %T from json: %w", item, err)
	}

	return m, nil
}

func AddIfNotNil[Value any](m map[string]any, key string, val *Value) {
	if val != nil {
		m[key] = val
	}
}

func AddIfMapNotEmpty[Value any](m map[string]any, key string, val map[string]*Value) {
	if len(val) > 0 {
		m[key] = val
	}
}

func AddIfSliceNotEmpty[Value any](m map[string]any, key string, val []Value) {
	if len(val) > 0 {
		m[key] = val
	}
}
