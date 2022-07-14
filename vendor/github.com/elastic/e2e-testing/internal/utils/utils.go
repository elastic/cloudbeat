// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package utils

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	internalio "github.com/elastic/e2e-testing/internal/io"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

//nolint:unused
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

//nolint:unused
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// DownloadRequest struct contains download details ad path and URL
type DownloadRequest struct {
	URL                 string
	DownloadPath        string
	UnsanitizedFilePath string
}

// GetArchitecture retrieves if the underlying system platform is arm64 or amd64
func GetArchitecture() string {
	arch, present := os.LookupEnv("GOARCH")
	if !present {
		arch = runtime.GOARCH
	}

	log.Debugf("Go's architecture is (%s)", arch)
	return arch
}

// DownloadFile will download a url and store it in a temporary path.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
func DownloadFile(downloadRequest *DownloadRequest) error {
	var filePath string
	if downloadRequest.DownloadPath == "" {
		tempParentDir := filepath.Join(os.TempDir(), uuid.NewString())
		internalio.MkdirAll(tempParentDir)
		filePath = filepath.Join(tempParentDir, uuid.NewString())
		downloadRequest.DownloadPath = filePath
	} else {
		filePath = filepath.Join(downloadRequest.DownloadPath, uuid.NewString())
	}

	tempFile, err := os.Create(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"url":   downloadRequest.URL,
		}).Error("Error creating file")
		return err
	}
	defer tempFile.Close()

	downloadRequest.UnsanitizedFilePath = tempFile.Name()
	exp := GetExponentialBackOff(3)

	retryCount := 1
	var fileReader io.ReadCloser
	download := func() error {
		resp, err := http.Get(downloadRequest.URL)
		if err != nil {
			log.WithFields(log.Fields{
				"elapsedTime": exp.GetElapsedTime(),
				"error":       err,
				"path":        downloadRequest.UnsanitizedFilePath,
				"retry":       retryCount,
				"url":         downloadRequest.URL,
			}).Warn("Could not download the file")

			retryCount++

			return err
		}

		log.WithFields(log.Fields{
			"elapsedTime": exp.GetElapsedTime(),
			"retries":     retryCount,
			"path":        downloadRequest.UnsanitizedFilePath,
			"url":         downloadRequest.URL,
		}).Trace("File downloaded")

		fileReader = resp.Body

		return nil
	}

	log.WithFields(log.Fields{
		"url":  downloadRequest.URL,
		"path": downloadRequest.UnsanitizedFilePath,
	}).Trace("Downloading file")

	err = backoff.Retry(download, exp)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	_, err = io.Copy(tempFile, fileReader)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"url":   downloadRequest.URL,
			"path":  downloadRequest.UnsanitizedFilePath,
		}).Error("Could not write file")

		return err
	}

	_ = os.Chmod(tempFile.Name(), 0666)

	return nil
}

// IsCommit returns true if the string matches commit format
func IsCommit(s string) bool {
	re := regexp.MustCompile(`^\b[0-9a-f]{5,40}\b`)

	return re.MatchString(s)
}

//nolint:unused
func randomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString generates a random string with certain length
func RandomString(length int) string {
	return randomStringWithCharset(length, charset)
}

// RemoveQuotes removes leading and trailing quotation marks
func RemoveQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

// Sleep sleeps a duration, including logs
func Sleep(duration time.Duration) error {
	fields := log.Fields{
		"duration": duration,
	}

	log.WithFields(fields).Tracef("Waiting %v", duration)
	time.Sleep(duration)

	return nil
}
