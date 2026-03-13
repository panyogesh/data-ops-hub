// Parse config.yaml
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

type ProviderConfig struct {
	Name    string `json:"name"`
	Compute string `json:"compute"`
	Package string `json:"package"`
	Source  string `json:"source"`
}

type Config struct {
	ClusterName string `json:"clusterName"`
	Helm        struct {
		CrossPlane struct {
			Chart     string `json:"chart"`
			RepoName  string `json:"reponame"`
			RepoURL   string `json:"repourl"`
			Version   string `json:"version"`
			Namespace string `json:"namespace"`
			Context   string `json:"context"`
		} `json:"cross_plane"`
	} `json:"helm"`
	KubeConfig struct {
		Context        string `json:"context"`
		Namespace      string `json:"namespace"`
		KubeConfigPath string `json:"kubeconfigPath"`
	} `json:"kubeconfig"`
	ProviderDetails ProviderConfig `json:"provider"`
	YAMLPaths       struct {
		Provider       string `json:"provider"`
		Secret         string `json:"secret"`
		ProviderConfig string `json:"providerConfig"`
	} `json:"yamlPaths"`
	GCP struct {
		ProjectID  string `json:"project_id"`
		Region     string `json:"region"`
		Zone       string `json:"zone"`
		SecretName string `json:"secret_name"`
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
