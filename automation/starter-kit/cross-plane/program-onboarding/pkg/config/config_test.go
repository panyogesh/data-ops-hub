// Write a simple test for the config package
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("config.json")
	assert.NoError(t, err)
	assert.Equal(t, "kind-crossplane-dev", config.KubeConfig.Context)
	assert.Equal(t, "crossplane-system", config.KubeConfig.Namespace)
}
