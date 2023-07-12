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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

func ParseConfig[T any]() (*T, error) {
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	var cfg T
	err = bindEnvs(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %v", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}

	return &cfg, nil
}

func bindEnvs(iface interface{}, parts ...string) error {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		var err error
		switch v.Kind() {
		case reflect.Struct:
			err = bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			err = viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
