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
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
	csppolicies "github.com/elastic/csp-security-policies/bundle"
)

var address = "127.0.0.1:18080"

var ServerAddress = "http://" + address

var Config = `{
        "services": {
            "test": {
                "url": %q
            }
        },
        "bundles": {
            "test": {
                "resource": "/bundles/bundle.tar.gz"
            }
        },
        "decision_logs": {
            "console": true
        }
    }`

func StartServer() (*http.Server, error) {
	policies, err := csppolicies.CISKubernetes()
	if err != nil {
		return nil, err
	}

	bundleServer := csppolicies.NewServer()
	err = bundleServer.HostBundle("bundle.tar.gz", policies)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:         address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      bundleServer,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logp.L().Errorf("bundle server closed: %v", err)
		}
	}()

	return srv, nil
}
