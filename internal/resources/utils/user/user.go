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

package user

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

// The util implementation is based on Go "os/user" native pkg

// User represents a user account.
type User struct {
	// Uid is the user ID.
	// On POSIX systems, this is a decimal number representing the uid.
	// On Windows, this is a security identifier (SID) in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Uid string
	// Gid is the primary group ID.
	// On POSIX systems, this is a decimal number representing the gid.
	// On Windows, this is a SID in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Gid string
	// Username is the login name.
	Username string
	// Name is the user's real or display name.
	// It might be blank.
	// On POSIX systems, this is the first (or only) entry in the GECOS field
	// list.
	// On Windows, this is the user's display name.
	// On Plan 9, this is the contents of /dev/user.
	Name string
	// HomeDir is the path to the user's home directory (if they have one).
	HomeDir string
}

// Group represents a grouping of users.
//
// On POSIX systems Gid contains a decimal number representing the group ID.
type Group struct {
	Gid  string // group ID
	Name string // group name
}

// lineFunc returns a value, an error, or (nil, nil) to skip the row.
type lineFunc func(line []byte) (v any, err error)

type OSUser interface {
	GetUserNameFromID(uid string, userFilePath string) (string, error)
	GetGroupNameFromID(gid string, groupFilePath string) (string, error)
}

type OSUserUtil struct{}

func NewOSUserUtil() OSUser {
	return &OSUserUtil{}
}

func (*OSUserUtil) GetUserNameFromID(uid string, userFilePath string) (string, error) {
	usr, err := lookupUserId(uid, userFilePath)
	if err != nil || usr == nil {
		return "", err
	}

	return usr.Username, nil
}

func (*OSUserUtil) GetGroupNameFromID(gid string, groupFilePath string) (string, error) {
	group, err := lookupGroupId(gid, groupFilePath)
	if err != nil || group == nil {
		return "", err
	}

	return group.Name, nil
}

func lookupGroupId(id string, filepath string) (*Group, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findGroupId(id, f)
}

func lookupUserId(uid string, filepath string) (*User, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findUserId(uid, f)
}

// readColonFile parses r as an /etc/group or /etc/passwd style file, running
// fn for each row. readColonFile returns a value, an error, or (nil, nil) if
// the end of the file is reached without a match.
//
// readCols is the minimum number of colon-separated fields that will be passed
// to fn; in a long line additional fields may be silently discarded.
//
//revive:disable-next-line:cognitive-complexity,cyclomatic
func readColonFile(r io.Reader, fn lineFunc, readCols int) (v any, err error) {
	rd := bufio.NewReader(r)

	// Read the file line-by-line.
	for {
		var isPrefix bool
		var wholeLine []byte

		// Read the next line. We do so in chunks (as much as reader's
		// buffer is able to keep), check if we read enough columns
		// already on each step and store final result in wholeLine.
		for {
			var line []byte
			line, isPrefix, err = rd.ReadLine()

			if err != nil {
				// We should return (nil, nil) if EOF is reached
				// without a match.
				if err == io.EOF {
					err = nil
				}
				return nil, err
			}

			// Simple common case: line is short enough to fit in a
			// single reader's buffer.
			if !isPrefix && len(wholeLine) == 0 {
				wholeLine = line
				break
			}

			wholeLine = append(wholeLine, line...)

			// Check if we read the whole line (or enough columns)
			// already.
			if !isPrefix || bytes.Count(wholeLine, []byte{':'}) >= readCols {
				break
			}
		}

		// There's no spec for /etc/passwd or /etc/group, but we try to follow
		// the same rules as the glibc parser, which allows comments and blank
		// space at the beginning of a line.
		wholeLine = bytes.TrimSpace(wholeLine)
		if len(wholeLine) == 0 || wholeLine[0] == '#' {
			continue
		}
		v, err = fn(wholeLine)
		if v != nil || err != nil {
			return
		}

		// If necessary, skip the rest of the line
		for ; isPrefix; _, isPrefix, err = rd.ReadLine() {
			if err != nil {
				// We should return (nil, nil) if EOF is reached without a match.
				if err == io.EOF {
					err = nil
				}
				return nil, err
			}
		}
	}
}

func matchGroupIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v any, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, []byte(":")) < 3 {
			return
		}
		// wheel:*:0:root
		parts := strings.SplitN(string(line), ":", 4)
		if len(parts) < 4 || parts[0] == "" || parts[idx] != value ||
			// If the file contains +foo and you search for "foo", glibc
			// returns an "invalid argument" error. Similarly, if you search
			// for a gid for a row where the group name starts with "+" or "-",
			// glibc fails to find the record.
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			//nolint:nilerr
			return nil, nil
		}
		return &Group{Name: parts[0], Gid: parts[2]}, nil
	}
}

func findGroupId(id string, r io.Reader) (*Group, error) {
	if v, err := readColonFile(r, matchGroupIndexValue(id, 2), 3); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*Group), nil
	}
	return nil, errors.New("Unknown groupId Error, GID: " + id)
}

// returns a *User for a row if that row's has the given value at the
// given index.
func matchUserIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v any, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, []byte(":")) < 6 {
			return
		}
		// kevin:x:1005:1006::/home/kevin:/usr/bin/zsh
		parts := strings.SplitN(string(line), ":", 7)
		if len(parts) < 6 || parts[idx] != value || parts[0] == "" ||
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			//nolint:nilerr
			return nil, nil
		}
		if _, err := strconv.Atoi(parts[3]); err != nil {
			//nolint:nilerr
			return nil, nil
		}
		u := &User{
			Username: parts[0],
			Uid:      parts[2],
			Gid:      parts[3],
			Name:     parts[4],
			HomeDir:  parts[5],
		}
		// The pw_gecos field isn't quite standardized. Some docs
		// say: "It is expected to be a comma separated list of
		// personal data where the first item is the full name of the
		// user."
		if i := strings.Index(u.Name, ","); i >= 0 {
			u.Name = u.Name[:i]
		}

		return u, nil
	}
}

func findUserId(uid string, r io.Reader) (*User, error) {
	_, e := strconv.Atoi(uid)
	if e != nil {
		return nil, errors.New("user: invalid userid " + uid)
	}
	if v, err := readColonFile(r, matchUserIndexValue(uid, 2), 6); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*User), nil
	}
	return nil, errors.New("Unknown UserId Error, UID: " + uid)
}
