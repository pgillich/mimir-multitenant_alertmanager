package configs

import (
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
)

func SaveServerConfig(serverConfig ServerConfig, dirName string, fileName string) (string, error) {
	configFile := filepath.Join(dirName, "multitenant_alertmanager.yaml")
	file, err := os.Create(configFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(serverConfig); err != nil {
		return "", err
	}

	return configFile, nil
}

func RenderServerConfig(serverConfig ServerConfig) (string, error) {
	yamlData, err := yaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}
