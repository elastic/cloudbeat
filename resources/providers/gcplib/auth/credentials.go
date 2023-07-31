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

package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/elastic/elastic-agent-libs/logp"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/config"
)

type GcpFactoryConfig struct {
	ProjectId  string
	ClientOpts []option.ClientOption
}

func GetGcpClientConfig(cfg config.GcpConfig, log *logp.Logger) ([]option.ClientOption, error) {
	log.Info("GetGCPClientConfig create credentials options")
	if cfg.CredentialsJSON == "" && cfg.CredentialsFilePath == "" {
		return nil, errors.New("the credentials file path or credentials JSON have not been specified")
	}

	var opts []option.ClientOption
	if cfg.CredentialsFilePath != "" {
		if err := validateJSONFromFile(cfg.CredentialsFilePath); err == nil {
			log.Infof("Appending credentials file path to gcp client options: %s", cfg.CredentialsFilePath)
			opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFilePath))
		} else {
			return nil, err
		}
	}

	if cfg.CredentialsJSON != "" {
		if json.Valid([]byte(cfg.CredentialsJSON)) {
			log.Info("Appending credentials JSON to client options")
			opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
		} else {
			return nil, errors.New("invalid credentials JSON")
		}
	}

	return opts, nil
}

func validateJSONFromFile(filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file %q cannot be found", filePath)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("the file %q cannot be read", filePath)
	}

	if !json.Valid(b) {
		return fmt.Errorf("the file %q does not contain valid JSON", filePath)
	}

	return nil
}
