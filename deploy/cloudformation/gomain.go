package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/spf13/viper"
)

type config struct {
	StackName           string `mapstructure:"STACK_NAME"`
	FleetURL            string `mapstructure:"FLEET_URL"`
	EnrollmentToken     string `mapstructure:"ENROLLMENT_TOKEN"`
	ElasticAgentVersion string `mapstructure:"ELASTIC_AGENT_VERSION"`
	Dev                 bool   `mapstructure:"DEV"`
	KeyName             string `mapstructure:"KEY_NAME"`
}

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

func parseConfig() (*config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	var cfg config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}

	if cfg.StackName == "" {
		return nil, fmt.Errorf("Missing required flag: -stack-name")
	}

	if cfg.FleetURL == "" {
		return nil, fmt.Errorf("Missing required flag: -fleet-url")
	}

	if cfg.EnrollmentToken == "" {
		return nil, fmt.Errorf("Missing required flag: -enrollment-token")
	}

	if cfg.Dev && cfg.KeyName == "" {
		return nil, fmt.Errorf("Missing required flag for development mode: -key-name")
	}

	if cfg.ElasticAgentVersion == "" {
		cfg.ElasticAgentVersion = "elastic-agent-8.8.0-SNAPSHOT-linux-arm64"
	}

	return &cfg, nil
}

func createFromConfig(cfg *config) error {
	params := map[string]string{}

	params["FleetUrl"] = cfg.FleetURL
	params["EnrollmentToken"] = cfg.EnrollmentToken
	params["ElasticAgentVersion"] = cfg.ElasticAgentVersion

	templatePath := prodTemplatePath
	if cfg.Dev {
		err := generateDevTemplate()
		if err != nil {
			return fmt.Errorf("Could not generate dev template: %v", err)
		}
		templatePath = devTemplatePath
		params["KeyName"] = cfg.KeyName
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

	file, err := ioutil.ReadFile(templatePath)
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
