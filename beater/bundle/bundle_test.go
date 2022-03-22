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

package bundle

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateServer(t *testing.T) {
	assert := assert.New(t)

	_, err := StartServer()
	assert.NoError(err)

	var tests = []struct {
		path               string
		expectedStatusCode string
	}{
		{
			"/bundles/bundle.tar.gz", "200 OK",
		},
		{
			"/bundles/notExistBundle.tar.gz", "404 Not Found",
		},
		{
			"/bundles/notExistBundle", "404 Not Found",
		},
	}

	time.Sleep(time.Second * 2)
	for _, test := range tests {
		target := ServerAddress + test.path
		client := &http.Client{}
		res, err := client.Get(target)

		assert.NoError(err)
		assert.Equal(test.expectedStatusCode, res.Status)
	}
}
