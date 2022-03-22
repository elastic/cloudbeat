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
	"github.com/magefile/mage/mg"

	devtools "github.com/elastic/beats/v7/dev-tools/mage"
)

// Update target namespace.
type Update mg.Namespace

// Aliases stores aliases for the targets.
var Aliases = map[string]interface{}{
	"update": Update.All,
}

// All updates all generated content.
func (Update) All() {
	mg.Deps(Update.Fields, Update.IncludeFields, Update.Config, Update.FieldDocs)
}

// Config generates both the short and reference configs.
func (Update) Config() error {
	return devtools.Config(devtools.ShortConfigType|devtools.ReferenceConfigType, XPackConfigFileParams(), ".")
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

	return devtools.GenerateAllInOneFieldsGo()
}
