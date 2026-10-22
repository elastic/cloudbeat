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
	"path/filepath"

	devtools "github.com/elastic/beats/v7/dev-tools/mage"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	// assetCmdPath is the beats generator that embeds fields.yml into a Go file.
	assetCmdPath = "dev-tools/cmd/asset/asset.go"
	// assetLicense is the identifier of Elastic License 2.0 expected by the generator.
	assetLicense = "Elastic"
)

// Update target namespace.
type Update mg.Namespace

// Aliases stores aliases for the targets.
var Aliases = map[string]any{
	"update": Update.All,
}

// All updates all generated content.
func (Update) All() {
	mg.Deps(Update.Fields, Update.IncludeFields, Update.Config, Update.FieldDocs)
}

// Config generates both the short and reference configs.
func (Update) Config() error {
	cfgArgs := XPackConfigFileParams()
	policyType := devtools.EnvOr("POLICY_TYPE", "")
	cfgArgs.ExtraVars["type"] = policyType
	return devtools.Config(devtools.ShortConfigType|devtools.ReferenceConfigType, cfgArgs, ".")
}

// Fields generates a fields.yml for the Beat.
func (Update) Fields() error {
	return devtools.GenerateFieldsYAML()
}

// FieldDocs collects all fields by provider and generates documentation for them.
func (Update) FieldDocs() error {
	mg.Deps(Update.Fields)

	return devtools.Docs.FieldDocs("fields.yml")
}

// IncludeFields generates include/fields.go by provider.
func (Update) IncludeFields() error {
	mg.Deps(Update.Fields)

	return generateFieldsGo("fields.yml", "include/fields.go")
}

// generateFieldsGo generates a Go file with the fields.yml data, like
// devtools.GenerateFieldsGo, but runs the generator with the toolchain of this
// repository. devtools.GenerateFieldsGo sets GOTOOLCHAIN to the Go version of
// the beats module, which is older than ours and makes the run fail with
// "go.mod requires go >= X (running go Y; GOTOOLCHAIN=goY)".
func generateFieldsGo(fieldsYML, out string) error {
	beatsDir, err := devtools.ElasticBeatsDir()
	if err != nil {
		return err
	}

	return sh.Run("go", "run",
		"-mod=readonly",
		filepath.Join(beatsDir, assetCmdPath),
		"-pkg", "include",
		"-in", fieldsYML,
		"-out", devtools.CreateDir(out),
		"-license", assetLicense,
		devtools.BeatName,
	)
}
