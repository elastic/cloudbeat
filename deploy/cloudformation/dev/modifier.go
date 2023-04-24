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

type ArtifactUrlDevMod struct {
	artifactUrl string
}

func (m *ArtifactUrlDevMod) Modify(template *cloudformation.Template) error {
	ec2Instance, err := template.GetEC2InstanceWithName("ElasticAgentEc2Instance")
	if err != nil {
		return err
	}

	m.artifactUrl, err = elasticAgentSnapshotArtifact()
	if err != nil {
		return err
	}

	err = m.recursiveReplaceArtifactUrl(ec2Instance.UserData)
	if err != nil {
		return err
	}

	return nil
}

func elasticAgentSnapshotArtifact() (string, error) {
	url := "https://artifacts-api.elastic.co/v1/search/8.8-SNAPSHOT/elastic-agent/"
	key := "elastic-agent-8.8.0-SNAPSHOT-linux-arm64.tar.gz"

	artifacts := map[string]interface{}{}
	err := getJson(url, &artifacts)
	if err != nil {
		return "", err
	}

	packages, ok := artifacts["packages"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Could not find packages field")
	}

	arm64Section, ok := packages[key].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Could not find arm64 section")
	}

	arm64Url, ok := arm64Section["url"].(string)
	if !ok {
		return "", fmt.Errorf("Could not find arm64 link")
	}
	return arm64Url, nil
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

func (m *ArtifactUrlDevMod) recursiveReplaceArtifactUrl(encoded *string) error {
	prodURL := "https://artifacts.elastic.co/downloads/beats/elastic-agent/"

	if strings.Index(*encoded, prodURL) > -1 {
		*encoded = strings.ReplaceAll(*encoded, prodURL, m.artifactUrl)
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*encoded)
	if err != nil {
		return fmt.Errorf("Could not decode user data: %v", err)
	}

	decodedObj := map[string]string{}
	err = json.Unmarshal(decoded, &decodedObj)
	if err != nil {
		return fmt.Errorf("Could not unmarshal user data: %v", err)
	}

	for k, v := range decodedObj {
		err = m.recursiveReplaceArtifactUrl(&v)
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
