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

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	StackName             string     `mapstructure:"STACK_NAME"`
	FleetURL              string     `mapstructure:"FLEET_URL"`
	EnrollmentToken       string     `mapstructure:"ENROLLMENT_TOKEN"`
	ElasticArtifactServer *string    `mapstructure:"ELASTIC_ARTIFACT_SERVER"`
	ElasticAgentVersion   string     `mapstructure:"ELASTIC_AGENT_VERSION"`
	Dev                   *devConfig `mapstructure:"DEV"`
}

type devConfig struct {
	AllowSSH bool   `mapstructure:"ALLOW_SSH"`
	KeyName  string `mapstructure:"KEY_NAME"`
}

func parseConfig() (*config, error) {
	// Read in configuration from on of the files: config.json, config.yml or config.env
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %v", err)
	}

	var cfg config
	err = bindEnvs(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %v", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}

	err = validateConfig(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func bindEnvs(iface any, parts ...string) error {
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

func validateConfig(cfg *config) error {
	if cfg.StackName == "" {
		return fmt.Errorf("missing required flag: STACK_NAME")
	}

	if cfg.FleetURL == "" {
		return fmt.Errorf("missing required flag: FLEET_URL")
	}

	if cfg.EnrollmentToken == "" {
		return fmt.Errorf("missing required flag: ENROLLMENT_TOKEN")
	}

	if cfg.ElasticAgentVersion == "" {
		return fmt.Errorf("missing required flag: ELASTIC_AGENT_VERSION")
	}

	if cfg.Dev != nil {
		return validateDevConfig(cfg.Dev)
	}

	return nil
}

func validateDevConfig(cfg *devConfig) error {
	if cfg.AllowSSH && cfg.KeyName == "" {
		return fmt.Errorf("missing required flag for SSH enablement mode: DEV.KEY_NAME")
	}

	return nil
}
