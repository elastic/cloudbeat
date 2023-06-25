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
	"errors"
	"github.com/elastic/cloudbeat/config"
	"google.golang.org/api/option"
)

func GetGcpClientConfig(cfg *config.Config) ([]option.ClientOption, error) {
	// Create credentials options
	gcpCred := cfg.CloudConfig.GcpCfg
	var opt []option.ClientOption
	if gcpCred.CredentialsFilePath != "" && gcpCred.CredentialsJSON != "" {
		return nil, errors.New("both credentials_file_path and credentials_json specified, you must use only one of them")
	} else if gcpCred.CredentialsFilePath != "" {
		opt = []option.ClientOption{option.WithCredentialsFile(gcpCred.CredentialsFilePath)}
	} else if gcpCred.CredentialsJSON != "" {
		opt = []option.ClientOption{option.WithCredentialsJSON([]byte(gcpCred.CredentialsJSON))}
	} else {
		return nil, errors.New("no credentials_file_path or credentials_json specified")
	}

	return opt, nil
}
