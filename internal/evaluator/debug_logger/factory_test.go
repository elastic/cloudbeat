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

package dlogger

import (
	"testing"

	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactoryNew(t *testing.T) {
	f := Factory{}
	manager, err := plugins.New([]byte{}, "test", inmem.New())
	require.NoError(t, err)

	p := f.New(manager, config{})
	assert.NotNil(t, p)
}

func TestFactoryValidate(t *testing.T) {
	f := Factory{}
	manager, err := plugins.New([]byte{}, "test", inmem.New())
	require.NoError(t, err)

	cfg, err := f.Validate(manager, []byte{})
	require.NoError(t, err)
	assert.IsType(t, config{}, cfg)
}
