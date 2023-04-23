package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	prodTemplatePath = "elastic-agent-ec2.yml"
	devTemplatePath  = "elastic-agent-ec2-dev.yml"
)

func editArtifactURL(content string) string {
	prodURL := "https://artifacts.elastic.co/downloads/beats/elastic-agent/"

	// TODO: Dynamically get the latest snapshot URL
	devURL := "https://snapshots.elastic.co/8.8.0-3f572553/downloads/beats/elastic-agent/"
	return strings.ReplaceAll(content, prodURL, devURL)
}

func allowSSHFromAnywhere(content string) string {
	blockSecurityGroup := `
      GroupDescription: Block incoming traffic
      SecurityGroupIngress: []`

	allowSSHSecurityGroup := `
      GroupDescription: Allow SSH from anywhere
      SecurityGroupIngress:
      - IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: 0.0.0.0/0`
	return strings.ReplaceAll(content, blockSecurityGroup, allowSSHSecurityGroup)
}

func acceptEC2Key(content string) string {
	parametersSection := `
Parameters:`

	keyParameter := `
Parameters:
  KeyName:
    Type: AWS::EC2::KeyPair::KeyName
    Description: SSH Keypair to login to the instance`
	return strings.ReplaceAll(content, parametersSection, keyParameter)
}

func assignEC2Key(content string) string {
	ec2Props := `
      ImageId: !Ref LatestAmiId`
	keyAssignment := `
      ImageId: !Ref LatestAmiId
      KeyName: !Ref KeyName`
	return strings.ReplaceAll(content, ec2Props, keyAssignment)
}

func generateDevTemplate() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Could not get exexutable: %v", err)
	}

	inputPath := filepath.Join(currentDir, prodTemplatePath)
	outputPath := filepath.Join(currentDir, devTemplatePath)

	fileContents, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("Could not read input: %v", err)
	}

	modifiedContents := editArtifactURL(string(fileContents))
	modifiedContents = allowSSHFromAnywhere(modifiedContents)
	modifiedContents = acceptEC2Key(modifiedContents)
	modifiedContents = assignEC2Key(modifiedContents)

	if err := ioutil.WriteFile(outputPath, []byte(modifiedContents), 0644); err != nil {
		return fmt.Errorf("Could not write output: %v", err)
	}

	log.Printf("Created dev template %s", outputPath)
	return nil
}
