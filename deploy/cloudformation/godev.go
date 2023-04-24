package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/awslabs/goformation/v7"
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/intrinsics"
)

const (
	prodTemplatePath = "elastic-agent-ec2.yml"
	devTemplatePath  = "elastic-agent-ec2-dev.yml"
)

type devModifier interface {
	Modify(template *cloudformation.Template) error
}

type artifactUrlDevMod struct{}

type securityGroupDevMod struct{}

type ec2KeyDevMod struct{}

var devModifiers = []devModifier{
	&securityGroupDevMod{}, &ec2KeyDevMod{}, &artifactUrlDevMod{},
}

func (m *artifactUrlDevMod) Modify(template *cloudformation.Template) error {
	ec2Instance, err := template.GetEC2InstanceWithName("ElasticAgentEc2Instance")
	if err != nil {
		return err
	}

	err = recursiveReplaceArtifactUrl(ec2Instance.UserData)
	if err != nil {
		return err
	}

	return nil
}

func recursiveReplaceArtifactUrl(encoded *string) error {
	// TODO: Dynamically get the latest snapshot URL
	devURL := "https://snapshots.elastic.co/8.8.0-3f572553/downloads/beats/elastic-agent/"
	prodURL := "https://artifacts.elastic.co/downloads/beats/elastic-agent/"

	if strings.Index(*encoded, prodURL) > -1 {
		*encoded = strings.ReplaceAll(*encoded, prodURL, devURL)
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
		err = recursiveReplaceArtifactUrl(&v)
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

func (m *securityGroupDevMod) Modify(template *cloudformation.Template) error {
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

func (m ec2KeyDevMod) Modify(template *cloudformation.Template) error {
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

func generateDevTemplate() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get exexutable: %v", err)
	}

	inputPath := filepath.Join(currentDir, prodTemplatePath)
	outputPath := filepath.Join(currentDir, devTemplatePath)

	template, err := goformation.OpenWithOptions(inputPath, &intrinsics.ProcessorOptions{
		IntrinsicHandlerOverrides: cloudformation.EncoderIntrinsics,
	})

	if err != nil {
		return fmt.Errorf("Could not read CloudFormation input: %v", err)
	}

	for _, m := range devModifiers {
		err := m.Modify(template)
		if err != nil {
			name := reflect.TypeOf(m)
			return fmt.Errorf("Modifier %s could not modify template: %v", name, err)
		}
	}

	yaml, err := template.YAML()
	if err != nil {
		return fmt.Errorf("Could not generate output yaml: %v", err)
	}

	if err := os.WriteFile(outputPath, yaml, 0644); err != nil {
		return fmt.Errorf("Could not write output: %v", err)
	}

	log.Printf("Created dev template %s", outputPath)
	return nil
}
