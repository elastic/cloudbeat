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

package mage

import (
	devtools "github.com/elastic/beats/v7/dev-tools/mage"
)

const opaBundle = "bundle.tar.gz"

// CustomizePackaging modifies the device in the configuration files based on
// the target OS.
func CustomizePackaging() {
	bundleDir := devtools.PackageFile{
		Mode:   0o644,
		Source: opaBundle,
	}

	for _, args := range devtools.Packages {
		if len(args.Types) == 0 {
			continue
		}

		// Add csp-policies bundle archive to package
		args.Spec.Files[opaBundle] = bundleDir

		// Remove files unused by cloudbeat.
		for filename, filespec := range args.Spec.Files {
			switch filespec.Source {
			case "_meta/kibana.generated", "fields.yml", "{{.BeatName}}.reference.yml":
				delete(args.Spec.Files, filename)
			default:
			}
		}
	}
}
