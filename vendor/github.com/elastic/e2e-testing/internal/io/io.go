// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package io

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory will be overridden if it
// exists. Symlinks are ignored and skipped.
func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.New("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// always override

	err = MkdirAll(dst)
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, 10000)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a file from a source to a destiny, always overridding
// the destination file
// Optimising the copy of files in Go:
// https://opensource.com/article/18/6/copying-files-go
func CopyFile(src string, dst string, bufferSize int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return errors.New(src + " is not a regular file")
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// always override

	err = MkdirAll(filepath.Dir(dst))
	if err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return err
}

// Exists checks if a path exists in the file system
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// MkdirAll creates all directories for a directory path
func MkdirAll(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"path":  path,
			}).Fatal("Directory cannot be created")

			return err
		}
	}

	return nil
}

// FindFiles finds files recursively using a Glob pattern for the matching
func FindFiles(pattern string) []string {
	matches, err := filepath.Glob(pattern)

	if err != nil {
		log.WithFields(log.Fields{
			"pattern": pattern,
		}).Warn("pattern is not a Glob")

		return []string{}
	}

	return matches
}

// ReadDir lists the contents of a directory
func ReadDir(path string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": path,
		}).Warn("Could not read file system")
		return []os.FileInfo{}, err
	}

	return files, nil
}

// ReadFile returns the byte array representing a file
func ReadFile(path string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": path,
		}).Warn("Could not read file")
		return []byte{}, err
	}

	return bytes, nil
}

// WriteFile writes bytes into target
func WriteFile(bytes []byte, target string) error {
	err := ioutil.WriteFile(target, bytes, 0755)
	if err != nil {
		log.WithFields(log.Fields{
			"target": target,
			"error":  err,
		}).Error("Cannot write file")

		return err
	}

	return nil
}
