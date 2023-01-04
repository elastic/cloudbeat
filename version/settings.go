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

package version

import (
	"runtime/debug"
	"time"
)

var (
	vcsRevision string
	vcsTime     time.Time
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.time":
				vcsTime, _ = time.Parse(time.RFC3339, setting.Value)
			case "vcs.revision":
				vcsRevision = setting.Value
			}
		}
	}
}

// cloudbeatCommitHash returns the hash of the git commit used for the build.
func cloudbeatCommitHash() string {
	return vcsRevision
}

// cloudbeatCommitTime returns the timestamp of the commit used for the build.
func cloudbeatCommitTime() string {
	return vcsTime.Format(time.RFC3339Nano)
}

// CloudbeatSemanticVersion returns the current cloudbeat version.
func CloudbeatSemanticVersion() string {
	return defaultBeatVersion
}

// CloudbeatVersion returns cloudbeat version info used for the build.
func CloudbeatVersion() Version {
	return Version{
		Version:    CloudbeatSemanticVersion(),
		CommitHash: cloudbeatCommitHash(),
		CommitTime: cloudbeatCommitTime(),
	}
}
