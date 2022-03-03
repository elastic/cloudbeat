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
	"github.com/pkg/errors"
)

// CustomizePackaging modifies the device in the configuration files based on
// the target OS.
func CustomizePackaging() {
	var (
		configYml = devtools.PackageFile{
			Mode:          0o600,
			Source:        "{{.PackageDir}}/cloudbeat.yml",
			Config:        true,
			SkipOnMissing: true,
			// todo: add deps to generate this files each build
		}
		//referenceConfigYml = devtools.PackageFile{
		//	Mode:          0o644,
		//	Source:        "{{.PackageDir}}/cloudbeat.reference.yml",
		//	SkipOnMissing: true,
		//	// todo: add deps to generate this files each build
		//}
	)

	for _, args := range devtools.Packages {
		if len(args.Types) == 0 {
			continue
		}
		// Replace the generic Beats README.md with an cloudbeat specific one, and remove files unused by apm-server.
		for filename, filespec := range args.Spec.Files {
			switch filespec.Source {
			case "_meta/kibana.generated", "fields.yml", "{{.BeatName}}.reference.yml":
				delete(args.Spec.Files, filename)
			}
		}

		switch pkgType := args.Types[0]; pkgType {
		case devtools.TarGz, devtools.Zip:
			args.Spec.ReplaceFile("{{.BeatName}}.yml", configYml)
			//args.Spec.ReplaceFile("{{.BeatName}}.reference.yml", referenceConfigYml)
		case devtools.Deb, devtools.RPM:
			args.Spec.ReplaceFile("/etc/{{.BeatName}}/{{.BeatName}}.yml", configYml)
			//args.Spec.ReplaceFile("/etc/{{.BeatName}}/{{.BeatName}}.reference.yml", referenceConfigYml)
		case devtools.Docker:
			args.Spec.ExtraVar("linux_capabilities", "cap_net_raw,cap_net_admin+eip")
		case devtools.DMG:
		default:
			panic(errors.Errorf("unhandled package type: %v, name: %v", pkgType, args.Spec.Name))
		}
	}
}
