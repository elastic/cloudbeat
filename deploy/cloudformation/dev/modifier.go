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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package dev

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
)

type ArtifactURLType string

var artifactUrlBase = map[ArtifactURLType]string{
	SnapshotArtifact: "https://snapshots.elastic.co/%s-%s/downloads/beats/elastic-agent/",
	StagingArtifact:  "https://staging.elastic.co/%s-%s/downloads/beats/elastic-agent/",
}

const (
	SnapshotArtifact ArtifactURLType = "snapshot"
	StagingArtifact  ArtifactURLType = "staging"

	prodUrlBase string = "https://artifacts.elastic.co/downloads/beats/elastic-agent/"

	latestSnapshotApi string = "https://artifacts-api.elastic.co/v1/search/%s-SNAPSHOT/elastic-agent/"
	latestSnapshotKey string = "elastic-agent-%s-SNAPSHOT-linux-arm64.tar.gz"
)

type ArtifactUrlDevMod struct {
	UrlType ArtifactURLType
	Latest  bool
	Version string
	Sha     string
}

func (m *ArtifactUrlDevMod) Modify(template *cloudformation.Template) error {
	err := m.validate()
	if err != nil {
		return err
	}

	ec2Instance, err := template.GetEC2InstanceWithName("ElasticAgentEc2Instance")
	if err != nil {
		return err
	}

	replaceUrl, err := m.resolveArtifactUrl()
	if err != nil {
		return err
	}

	err = m.recursiveReplaceArtifactUrl(ec2Instance.UserData, replaceUrl)
	if err != nil {
		return err
	}

	return nil
}

func (m *ArtifactUrlDevMod) validate() error {
	if m.Latest && m.Sha != "" {
		return fmt.Errorf("Cannot specify both latest and sha")
	}

	if !m.Latest && m.Sha == "" {
		return fmt.Errorf("Must specify either latest or sha")
	}

	if m.Sha == "" && m.UrlType == StagingArtifact {
		return fmt.Errorf("Must specify sha when using staging artifact")
	}

	return nil
}
func (m *ArtifactUrlDevMod) resolveArtifactUrl() (string, error) {
	if m.Sha != "" {
		baseUrl, ok := artifactUrlBase[m.UrlType]
		if !ok {
			return "", fmt.Errorf("Could not recognize base Url for artifact: %s", m.UrlType)
		}

		return fmt.Sprintf(baseUrl, m.Version, m.Sha), nil
	}

	return m.getLatestSnapshotArtifact()
}

func (m *ArtifactUrlDevMod) getLatestSnapshotArtifact() (string, error) {
	url := fmt.Sprintf(latestSnapshotApi, m.Version)
	key := fmt.Sprintf(latestSnapshotKey, m.Version)
	artifacts := map[string]interface{}{}
	err := getJson(url, &artifacts)
	if err != nil {
		return "", err
	}

	packages, ok := artifacts["packages"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not find packages field")
	}

	arm64Section, ok := packages[key].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not find arm64 section")
	}

	arm64Url, ok := arm64Section["url"].(string)
	if !ok {
		return "", fmt.Errorf("could not find arm64 URL")
	}
	return strings.TrimSuffix(arm64Url, key), nil
}

func getJson(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func (m *ArtifactUrlDevMod) recursiveReplaceArtifactUrl(encoded *string, replaceUrl string) error {
	if strings.Contains(*encoded, prodUrlBase) {
		*encoded = strings.ReplaceAll(*encoded, prodUrlBase, replaceUrl)
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*encoded)
	if err != nil {
		return fmt.Errorf("could not decode user data: %v", err)
	}

	decodedObj := map[string]string{}
	err = json.Unmarshal(decoded, &decodedObj)
	if err != nil {
		return fmt.Errorf("could not unmarshal user data: %v", err)
	}

	for k, v := range decodedObj {
		err = m.recursiveReplaceArtifactUrl(&v, replaceUrl)
		decodedObj[k] = v
		if err != nil {
			return err
		}
	}

	decoded, err = json.Marshal(decodedObj)
	if err != nil {
		return err
	}

	*encoded = base64.StdEncoding.EncodeToString(decoded)
	return nil
}

type SecurityGroupDevMod struct{}

func (m *SecurityGroupDevMod) Modify(template *cloudformation.Template) error {
	securityGroups, err := template.GetEC2SecurityGroupWithName("ElasticAgentSecurityGroup")
	if err != nil {
		return err
	}
	securityGroups.GroupDescription = "Allow SSH from anywhere"
	securityGroups.SecurityGroupIngress = []ec2.SecurityGroup_Ingress{
		{
			IpProtocol: "tcp",
			FromPort:   cloudformation.Int(22),
			ToPort:     cloudformation.Int(22),
			CidrIp:     cloudformation.String("0.0.0.0/0"),
		},
	}
	return nil
}

type Ec2KeyDevMod struct{}

func (m Ec2KeyDevMod) Modify(template *cloudformation.Template) error {
	template.Parameters["KeyName"] = cloudformation.Parameter{
		Type:        "AWS::EC2::KeyPair::KeyName",
		Description: cloudformation.String("SSH Keypair to login to the instance"),
	}

	ec2Instance, err := template.GetEC2InstanceWithName("ElasticAgentEc2Instance")
	if err != nil {
		return err
	}

	ec2Instance.KeyName = cloudformation.RefPtr("KeyName")
	return nil
}
