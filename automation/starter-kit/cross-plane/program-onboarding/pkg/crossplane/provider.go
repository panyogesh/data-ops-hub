// Check whether Provider exists if not create it
// Also Add cleanup function to delete the provider after the test is done
package crossplane

import (
	"fmt"
	"strings"
	"text/template"

	"example.com/gridnode/pkg/config"
)

var ProviderTemplate = `
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: {{.ProviderName}}
spec:
  package: {{.Package}}
`

type Provider struct {
	Name   string
	yaml   string
	config config.ProviderConfig
}

func NewProvider(cfg config.Config) (*Provider, error) {
	name := cfg.ProviderDetails.Name
	if name == "" {
		return nil, fmt.Errorf("provider name is not set in config")
	}

	provider := &Provider{
		Name:   name,
		config: cfg.ProviderDetails,
	}

	tmpl, err := template.New("provider").Parse(ProviderTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse provider template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, struct {
		ProviderName string
		Package      string
	}{
		ProviderName: provider.Name,
		Package:      cfg.ProviderDetails.Package,
	})
	if err != nil {
		return nil, fmt.Errorf("execute provider template: %w", err)
	}

	provider.yaml = buf.String()
	return provider, nil
}

// YAML returns the rendered provider manifest.
func (p *Provider) YAML() string {
	return p.yaml
}

func (p *Provider) String() string {
	return p.Name
}
