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

package version

// name matches github.com/elastic/beats/v7/dev-tools/mage/settings.go parseBeatVersion
const defaultBeatVersion = "9.1.2"

var qualifier = ""

// Version represents version information for a package
type Version struct {
	Version    string `mapstructure:"version,omitempty"`     // Version is the semantic version of the package
	CommitHash string `mapstructure:"commit_sha,omitempty"`  // CommitHash is the git commit hash of the package
	CommitTime string `mapstructure:"commit_time,omitempty"` // CommitTime is the git commit time of the package
}

type CloudbeatVersionInfo struct {
	Version `mapstructure:",squash"` // Version info for cloudbeat
	Policy  Version                  `mapstructure:"policy,omitempty"` // Policy version info for the rules policy
}
