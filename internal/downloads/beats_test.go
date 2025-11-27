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

package downloads

import (
	"context"
	"testing"
	"os"
)

func TestFetchBeatsBinary(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "downloads_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with a small, commonly available artifact (this is just a smoke test)
	// We're testing that the function doesn't panic and creates the proper file structure
	ctx := context.Background()
	
	// Test that the function doesn't crash with valid parameters
	// We don't actually want to download a large file in tests, so we'll just check
	// that the function signature works and doesn't cause import errors
	artifactName := "test-artifact.tar.gz"
	artifact := "filebeat"
	version := "8.0.0"
	
	// This should fail to find the artifact (which is expected for this test name)
	// but should not cause compilation or import errors
	_, err = FetchBeatsBinary(ctx, artifactName, artifact, version, 1, false, tmpDir, false)
	
	// We expect this to fail since "test-artifact.tar.gz" doesn't exist
	// The important thing is that it compiles and the function is callable
	if err == nil {
		t.Error("Expected error for non-existent artifact, but got none")
	}
	
	// Check that error message is reasonable (contains some indication of failure)
	if err != nil && err.Error() == "" {
		t.Error("Error should have a message")
	}
	
	t.Logf("Function executed successfully with expected error: %v", err)
}

func TestDownloadFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "downloads_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	
	// Test with an invalid URL to ensure error handling works
	url := "https://invalid-url-that-should-not-exist.example.com/file.txt"
	fileName := "test-file.txt"
	
	_, err = downloadFile(ctx, url, tmpDir, fileName, 5) // 5 seconds timeout
	
	// We expect this to fail since the URL doesn't exist
	if err == nil {
		t.Error("Expected error for invalid URL, but got none")
	}
	
	// Check that the error is meaningful
	if err != nil && err.Error() == "" {
		t.Error("Error should have a message")
	}
	
	t.Logf("Download function executed successfully with expected error: %v", err)
}