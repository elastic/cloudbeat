package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/awslabs/goformation/v7"
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/intrinsics"
	"github.com/elastic/cloudbeat/deploy/cloudformation/dev"
)

const (
	prodTemplatePath = "elastic-agent-ec2.yml"
	devTemplatePath  = "elastic-agent-ec2-dev.yml"
)

type devModifier interface {
	Modify(template *cloudformation.Template) error
}

var devModifiers = []devModifier{
	&dev.SecurityGroupDevMod{}, &dev.Ec2KeyDevMod{}, &dev.ArtifactUrlDevMod{},
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
