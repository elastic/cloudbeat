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

package awslib

import (
	"errors"
	"fmt"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

// WithGovCloudFallback calls getIdentity. If getIdentity fails and defaultRegion is empty,
// it sets defaultRegion to DefaultGovRegion and calls getIdentity again.
// The default region of the AWS SDK is a commercial region (us-east-1), where GovCloud credentials fail.
// If the second call also fails, the returned error contains the errors of the two calls.
func WithGovCloudFallback(log *clog.Logger, defaultRegion *string, getIdentity func() error) error {
	err := getIdentity()
	if err == nil || *defaultRegion != "" {
		return err
	}

	log.Warnf("failed to initialize identity (%v); retrying to check AWS Gov Cloud regions", err)
	*defaultRegion = DefaultGovRegion
	if retryErr := getIdentity(); retryErr != nil {
		// Keep the first error. It can show the real cause, for example credentials that are not valid.
		return errors.Join(err, fmt.Errorf("retry in region %s: %w", DefaultGovRegion, retryErr))
	}
	return nil
}
