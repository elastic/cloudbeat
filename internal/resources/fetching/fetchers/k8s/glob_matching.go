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

package fetchers

import (
	"os"
	"path/filepath"
	"strings"
)

// Globs represents one filepath glob, with its elements joined by "**".
// Based on https://github.com/yargevad/filepathx/blob/master/filepathx.go
type Globs []string

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// pass to core package if no double-star
		return filepath.Glob(pattern)
	}
	return Globs(strings.Split(pattern, "**")).Expand()
}

// Expand finds matches for the provided Globs.
func (globs Globs) Expand() ([]string, error) {
	matches := []string{""} // accumulate here
	for _, glob := range globs {
		var hits []string
		hitMap := map[string]bool{}
		for _, match := range matches {
			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				err = filepath.Walk(path, func(walkPath string, _ os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					// save deduped match from current iteration
					if _, ok := hitMap[walkPath]; !ok {
						hits = append(hits, walkPath)
						hitMap[walkPath] = true
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	// fix up return value for nil input
	if globs == nil && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}
