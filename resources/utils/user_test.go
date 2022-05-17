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

package utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type testAttr struct {
	name     string
	id       uint32
	expected string
}

const (
	UserFile  = "./mocked_psswd_file.txt"
	GroupFile = "./mocked_group_file.txt"
)

type UserTestSuite struct {
	suite.Suite
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (s UserTestSuite) SetupTest() {}

func (s UserTestSuite) TearDownTest() {}

func (s UserTestSuite) TestGetUserNameFromID() {
	var userTests = []testAttr{
		{
			name:     "Should return root as a username",
			id:       0,
			expected: "root",
		},
		{
			name:     "Should return daemon as a username",
			id:       1,
			expected: "daemon",
		},
		{
			name:     "Should return Proxy as a username - no friendly name exists",
			id:       13,
			expected: "proxy",
		},
		{
			name:     "Should not return a username",
			id:       6666,
			expected: "",
		},
	}

	for _, tt := range userTests {
		s.SetupTest()
		s.Run(tt.name, func() {
			username := GetUserNameFromID(tt.id, UserFile)
			s.Equal(tt.expected, username)
		})
	}
}

func (s UserTestSuite) TestGetGroupNameFromID() {
	var userTests = []testAttr{
		{
			name:     "Should return wheel as group name",
			id:       0,
			expected: "wheel",
		},
		{
			name:     "Should return daemon as group name",
			id:       1,
			expected: "daemon",
		},
		{
			name:     "Should not return group name",
			id:       1000,
			expected: "",
		},
	}

	for _, tt := range userTests {
		s.SetupTest()
		s.Run(tt.name, func() {
			username := GetGroupNameFromID(tt.id, GroupFile)
			s.Equal(tt.expected, username)
		})
	}
}
