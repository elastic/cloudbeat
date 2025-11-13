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
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

// FileTokenRetriever implements stscreds.IdentityTokenRetriever by reading a JWT token from a file.
// This is used for AssumeRoleWithWebIdentity authentication.
type FileTokenRetriever struct {
	filePath string
}

// NewFileTokenRetriever creates a new FileTokenRetriever that reads tokens from the specified file path.
func NewFileTokenRetriever(filePath string) *FileTokenRetriever {
	return &FileTokenRetriever{filePath: filePath}
}

// GetIdentityToken reads and returns the JWT token from the configured file.
func (f *FileTokenRetriever) GetIdentityToken() ([]byte, error) {
	tokenBytes, err := os.ReadFile(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT token from %s: %w", f.filePath, err)
	}
	return tokenBytes, nil
}

// Compile-time check to ensure FileTokenRetriever implements stscreds.IdentityTokenRetriever
var _ stscreds.IdentityTokenRetriever = (*FileTokenRetriever)(nil)
