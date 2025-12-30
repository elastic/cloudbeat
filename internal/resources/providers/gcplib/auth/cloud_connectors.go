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

package auth

import (
	"context"

	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/providers/gcplib"
)

// initializeGCPConfigCloudConnectors is a wrapper to avoid circular import dependencies
// between auth and gcplib packages
func initializeGCPConfigCloudConnectors(ctx context.Context, cfg config.GcpConfig) ([]option.ClientOption, error) {
	return gcplib.InitializeGCPConfigCloudConnectors(ctx, cfg)
}
