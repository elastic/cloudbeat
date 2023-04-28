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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/elastic/cloudbeat/deploy/cloudformation/dev"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = createFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func createFromConfig(cfg *config) error {
	params := map[string]string{}

	params["FleetUrl"] = cfg.FleetURL
	params["EnrollmentToken"] = cfg.EnrollmentToken
	params["ElasticAgentVersion"] = cfg.ElasticAgentVersion

	templatePath := prodTemplatePath
	if cfg.Dev != nil {
		modifiers := []devModifier{}
		if cfg.Dev.AllowSSH {
			modifiers = append(modifiers, &dev.SecurityGroupDevMod{}, &dev.Ec2KeyDevMod{})
			params["KeyName"] = cfg.Dev.KeyName
		}

		if cfg.Dev.PreRelease {
			rawVersion := strings.TrimSuffix(cfg.ElasticAgentVersion, "-SNAPSHOT")
			artifactModifier := &dev.ArtifactUrlDevMod{
				Version: rawVersion,
				Sha:     cfg.Dev.Sha,
				UrlType: dev.StagingArtifact,
			}

			if strings.HasSuffix(cfg.ElasticAgentVersion, "-SNAPSHOT") {
				artifactModifier.UrlType = dev.SnapshotArtifact
			}

			if cfg.Dev.Sha == "" {
				artifactModifier.Latest = true
			}
			modifiers = append(modifiers, artifactModifier)
		}

		err := generateDevTemplate(modifiers)
		if err != nil {
			return fmt.Errorf("Could not generate dev template: %v", err)
		}
		templatePath = devTemplatePath
	}

	err := createStack(cfg.StackName, templatePath, params)
	if err != nil {
		return fmt.Errorf("Failed to create CloudFormation stack: %v", err)
	}

	return nil
}

func createStack(stackName string, templatePath string, params map[string]string) error {
	ctx := context.Background()

	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("Failed to load AWS SDK config: %v", err)
	}

	svc := cloudformation.NewFromConfig(cfg)
	var cfParams []types.Parameter
	for key, value := range params {
		p := types.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		}
		cfParams = append(cfParams, p)
	}

	file, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("Failed to open template file: %v", err)
	}
	filestring := string(file)

	createStackInput := &cloudformation.CreateStackInput{
		StackName:    &stackName,
		TemplateBody: &filestring,
		Parameters:   cfParams,
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
	}

	stackOutput, err := svc.CreateStack(ctx, createStackInput)
	if err != nil {
		return fmt.Errorf("Failed to call AWS CloudFormation CreateStack: %v", err)
	}

	log.Printf("Created stack %s", *stackOutput.StackId)
	return nil
}
