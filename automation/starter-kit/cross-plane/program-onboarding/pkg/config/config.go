// Parse config.yaml
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

type Config struct {
	ClusterName string `yaml:"clusterName"`
	Helm        struct {
		CrossPlane struct {
			Chart    string `json:"chart"`
			RepoName string `json:"repoName"`
			Context  string `json:"context"`
		} `json:"crossPlane"`
	} `json:"helm"`
	KubeConfig struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
	} `json:"kubeconfig"`
	GCP struct {
		ProjectID  string `json:"projectId"`
		Region     string `json:"region"`
		Zone       string `json:"zone"`
		SecretName string `json:"secretName"`
	} `json:"gcp"`
}

func LoadConfig(path string) (*Config, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}
	defer jsonFile.Close()

	var config Config

	// 2. Read the file into a byte array
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, fmt.Errorf("Failed to parse %s ", path)
	}

	// 3. Unmarshal (parse) the JSON into our struct
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, fmt.Errorf("Failed to parse %s ", path)
	}

	return &config, nil
}
