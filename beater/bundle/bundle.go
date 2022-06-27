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
	"fmt"
	"net/http"
	"time"

	csppolicies "github.com/elastic/csp-security-policies/bundle"
	"github.com/elastic/elastic-agent-libs/logp"
)

var (
	address = "127.0.0.1:18080"

	ServerAddress = fmt.Sprintf("http://%s", address)
	Config        = `{
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
            "console": false
        }
    }`
)

func StartServer() (*http.Server, error) {
	policies, err := csppolicies.CISKubernetes()
	if err != nil {
		return nil, err
	}

	h := csppolicies.NewServer()
	if err := csppolicies.HostBundle("bundle.tar.gz", policies); err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:         address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h,
	}

	log := logp.NewLogger("cloudbeat_bundle_server")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Errorf("Bundle server closed: %v", err)
		}
	}()

	return srv, nil
}
