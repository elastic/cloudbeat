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

package gcplib

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

func GetGcpClientConfig(cfg *config.Config, log *logp.Logger) ([]option.ClientOption, error) {
	log.Info("GetGCPClientConfig create credentials options")
	gcpCred := cfg.CloudConfig.GcpCfg
	if gcpCred.CredentialsJSON == "" && gcpCred.CredentialsFilePath == "" {
		return nil, errors.New("no credentials_file_path or credentials_json specified")
	}

	var opts []option.ClientOption
	if gcpCred.CredentialsFilePath != "" {
		if err := validateJSONFromFile(log, gcpCred.CredentialsFilePath); err == nil {
			log.Infof("Appending credentials file path to gcp client options: %s", gcpCred.CredentialsFilePath)
			opts = append(opts, option.WithCredentialsFile(gcpCred.CredentialsFilePath))
		} else {
			return nil, err
		}
	}

	if gcpCred.CredentialsJSON != "" {
		if json.Valid([]byte(gcpCred.CredentialsJSON)) {
			log.Info("Appending credentials JSON to client options")
			opts = append(opts, option.WithCredentialsJSON([]byte(gcpCred.CredentialsJSON)))
		} else {
			return nil, errors.New("invalid credentials JSON")
		}
	}

	return opts, nil
}

func validateJSONFromFile(log *logp.Logger, filePath string) error {
	if _, err := os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the file %q cannot be found", filePath)

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
