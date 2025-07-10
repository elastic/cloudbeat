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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FetchBeatsBinary downloads a beats binary from the Elastic artifacts repository
// This is a simplified implementation to replace the github.com/elastic/e2e-testing dependency
func FetchBeatsBinary(ctx context.Context, artifactName, artifact, version string, timeoutFactor int, xpack bool, downloadPath string, downloadSHAFile bool) (string, error) {
	// First try to download from releases API
	downloadURL, err := getBeatsReleaseURL(artifact, version, artifactName)
	if err != nil {
		// Fallback to artifacts API
		downloadURL, err = getBeatsArtifactURL(artifact, version, artifactName)
		if err != nil {
			return "", fmt.Errorf("failed to find download URL for %s %s: %w", artifact, version, err)
		}
	}

	fmt.Printf("Downloading %s from %s\n", artifactName, downloadURL)
	
	// Download the binary
	filePath, err := downloadFile(ctx, downloadURL, downloadPath, artifactName, time.Duration(timeoutFactor)*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", artifactName, err)
	}

	// Download SHA512 file if requested
	if downloadSHAFile {
		shaURL := downloadURL + ".sha512"
		shaFileName := artifactName + ".sha512"
		_, err = downloadFile(ctx, shaURL, downloadPath, shaFileName, time.Duration(timeoutFactor)*time.Minute)
		if err != nil {
			// SHA file download is not critical, just log the error
			fmt.Printf("Warning: failed to download SHA file for %s: %v\n", artifactName, err)
		}
	}

	return filePath, nil
}

// getBeatsReleaseURL constructs the URL for downloading from the releases API
func getBeatsReleaseURL(artifact, version, artifactName string) (string, error) {
	// Construct release URL
	baseURL := "https://artifacts.elastic.co/downloads"
	return fmt.Sprintf("%s/%s/%s", baseURL, artifact, artifactName), nil
}

// getBeatsArtifactURL constructs the URL for downloading from the artifacts API
func getBeatsArtifactURL(artifact, version, artifactName string) (string, error) {
	// For snapshot versions, use the artifacts API to get the exact URL
	artifactAPIURL := fmt.Sprintf("https://artifacts-api.elastic.co/v1/search/%s/%s", version, artifactName)
	
	resp, err := http.Get(artifactAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to query artifacts API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("artifacts API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read artifacts API response: %w", err)
	}

	var apiResponse struct {
		Packages map[string]struct {
			URL string `json:"url"`
		} `json:"packages"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("failed to parse artifacts API response: %w", err)
	}

	for _, pkg := range apiResponse.Packages {
		if pkg.URL != "" {
			return pkg.URL, nil
		}
	}

	return "", fmt.Errorf("no download URL found in artifacts API response")
}

// downloadFile downloads a file from the given URL to the specified directory
func downloadFile(ctx context.Context, url, downloadDir, fileName string, timeout time.Duration) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Ensure download directory exists
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Create the output file
	filePath := filepath.Join(downloadDir, fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy the response body to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}