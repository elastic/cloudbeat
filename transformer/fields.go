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

package transformer

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/ecs"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type Fields struct {
	Process    ecs.Process             `ecs:"process"`
	File       ecs.File                `ecs:"file"`
	Event      ecs.Event               `ecs:"event"`
	Resource   fetching.ResourceFields `ecs:"resource"`
	ResourceID string                  `ecs:"resource_id"` // Deprecated
	Type       string                  `ecs:"type"`        // Deprecated
	Result     evaluator.Result        `ecs:"result"`
	Rule       evaluator.Rule          `ecs:"rule"`
	Message    string                  `ecs:"message"`
}

// MarshalMapStr marshals the fields into MapStr. It returns an error if there
// is a problem writing the keys to the given map (like if an intermediate key
// exists and is not a map).
func (f *Fields) MarshalMapStr(m mapstr.M) error {
	typ := reflect.TypeOf(*f)
	val := reflect.ValueOf(*f)

	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		tag := structField.Tag.Get("ecs")
		if tag == "" {
			continue
		}

		fieldValue := val.Field(i)
		if !fieldValue.IsValid() || isEmptyValue(fieldValue) {
			continue
		}

		if err := marshalStruct(m, tag, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

func getTag(f reflect.StructField) string {
	if tag := f.Tag.Get("ecs"); tag != "" {
		return tag
	}
	return ""
}

func marshalStruct(m mapstr.M, key string, val reflect.Value) error {
	// Dereference pointers.
	if val.Type().Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}

		val = val.Elem()
	}

	// Ignore zero values.
	if !val.IsValid() {
		return nil
	}

	typ := val.Type()
	if typ.Kind() == reflect.String {
		v := val.Interface()
		_, err := m.Put(key, v)
		return err
	}
	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		tag := getTag(structField)
		if tag == "" {
			continue
		}

		inline := false
		tags := strings.Split(tag, ",")
		if len(tags) > 1 {
			for _, flag := range tags[1:] {
				switch flag {
				case "inline":
					inline = true
				default:
					return fmt.Errorf("unsupported flag %q in tag %q of type %s", flag, tag, typ)
				}
			}
			tag = tags[0]
		}

		fieldValue := val.Field(i)
		if !fieldValue.IsValid() || isEmptyValue(fieldValue) {
			continue
		}

		if inline {
			if err := marshalStruct(m, key, fieldValue); err != nil {
				return err
			}
		} else {
			if _, err := m.Put(key+"."+tag, fieldValue.Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

// isEmptyValue returns true if the given value is empty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int64:
		if duration, ok := v.Interface().(time.Duration); ok {
			return duration <= 0
		}
		return v.Int() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	switch t := v.Interface().(type) {
	case time.Time:
		return t.IsZero()
	}
	return false
}
