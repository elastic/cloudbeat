package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/elastic/cloudbeat/deploy/util"

	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v3"
)

const (
	templateFile = "compute-engine.py"
	dmConfFile   = "dmconf.yml"
	gcpProject   = "elastic-security-test"
)

type Resource struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type DmConf struct {
	Imports   interface{} `json:"imports"`
	Resources []Resource  `json:"resources"`
}

type config struct {
	DeploymentName        string  `mapstructure:"DEPLOYMENT_NAME"`
	FleetURL              string  `mapstructure:"FLEET_URL"`
	EnrollmentToken       string  `mapstructure:"ENROLLMENT_TOKEN"`
	ElasticArtifactServer *string `mapstructure:"ELASTIC_ARTIFACT_SERVER"`
	ElasticAgentVersion   string  `mapstructure:"ELASTIC_AGENT_VERSION"`
	Zone                  string  `mapstructure:"ZONE"`
	AllowSSH              bool    `mapstructure:"ALLOW_SSH"`
}

func main() {
	cfg, err := util.ParseConfig[config]()
	if err != nil {
		log.Fatal(err)
	}

	err = validateInput(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = createDeployment(cfg.DeploymentName, templateFile, cfg)
	if err != nil {
		log.Fatalf("failed to create CloudFormation stack: %v", err)
	}
}

func createDeployment(deploymentName string, templatePath string, cfg *config) error {
	templateData, confByte, err := createDeploymentCfg(cfg, templatePath)
	if err != nil {
		return err
	}

	ctx := context.Background()
	dmService, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return err
	}

	d := &deploymentmanager.Deployment{
		Name: deploymentName,
		Target: &deploymentmanager.TargetConfiguration{
			Config: &deploymentmanager.ConfigFile{
				Content: string(confByte),
			},
			Imports: []*deploymentmanager.ImportFile{
				{
					Content: string(templateData),
					Name:    templateFile,
				},
			},
		},
	}

	_, err = dmService.Deployments.Insert(gcpProject, d).Context(ctx).Do()
	if err != nil {
		return err
	}
	return nil
}

func createDeploymentCfg(cfg *config, templatePath string) ([]byte, []byte, error) {
	var dmconf DmConf
	err := loadConfig(dmConfFile, &dmconf)
	if err != nil {
		return nil, nil, err
	}

	dmconf.Resources[0].Name = cfg.DeploymentName
	dmconf.Resources[0].Properties["zone"] = cfg.Zone
	dmconf.Resources[0].Properties["fleetUrl"] = cfg.FleetURL
	dmconf.Resources[0].Properties["enrollmentToken"] = cfg.EnrollmentToken
	dmconf.Resources[0].Properties["elasticAgentVersion"] = cfg.ElasticAgentVersion
	dmconf.Resources[0].Properties["allowSSH"] = cfg.AllowSSH

	if cfg.ElasticArtifactServer != nil {
		dmconf.Resources[0].Properties["ElasticArtifactServer"] = *cfg.ElasticArtifactServer
	}

	confByte, err := yaml.Marshal(dmconf)
	if err != nil {
		return nil, nil, err
	}

	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, nil, err
	}

	return templateData, confByte, nil
}

// loadConfig Load yaml config
func loadConfig(path string, o interface{}) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, o); err != nil {
		return err
	}
	return nil
}

func validateInput(cfg *config) error {
	if cfg.DeploymentName == "" {
		return fmt.Errorf("missing required flag: DEPLOYMENT_NAME")
	}

	if cfg.FleetURL == "" {
		return fmt.Errorf("missing required flag: FLEET_URL")
	}

	if cfg.EnrollmentToken == "" {
		return fmt.Errorf("missing required flag: ENROLLMENT_TOKEN")
	}

	if cfg.ElasticAgentVersion == "" {
		return fmt.Errorf("missing required flag: ELASTIC_AGENT_VERSION")
	}

	if cfg.Zone == "" {
		return fmt.Errorf("missing required flag: ZONE")
	}

	return nil
}
