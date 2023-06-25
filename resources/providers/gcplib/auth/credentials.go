package gcplib

import (
	"errors"
	"github.com/elastic/cloudbeat/config"
	"google.golang.org/api/option"
)

func GetGcpClientConfig(cfg *config.Config) ([]option.ClientOption, error) {
	// Create credentials options
	gcpCred := cfg.CloudConfig.GcpCfg
	var opt []option.ClientOption
	if gcpCred.CredentialsFilePath != "" && gcpCred.CredentialsJSON != "" {
		return nil, errors.New("both credentials_file_path and credentials_json specified, you must use only one of them")
	} else if gcpCred.CredentialsFilePath != "" {
		opt = []option.ClientOption{option.WithCredentialsFile(gcpCred.CredentialsFilePath)}
	} else if gcpCred.CredentialsJSON != "" {
		opt = []option.ClientOption{option.WithCredentialsJSON([]byte(gcpCred.CredentialsJSON))}
	} else {
		return nil, errors.New("no credentials_file_path or credentials_json specified")
	}

	return opt, nil
}
