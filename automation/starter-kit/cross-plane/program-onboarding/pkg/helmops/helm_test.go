package helmops

import (
	"context"
	"testing"
)

// Write the testcases to check if the helm chart is installed and if not, install it.
func TestInstallCrossplane(t *testing.T) {
	ctx := context.Background()
	helmConfig := HelmConfig{
		Context: "kind-crossplane-dev",
	}

	err := helmConfig.InstallCrossplane(ctx)
	if err != nil {
		t.Fatalf("Failed to install Crossplane: %v", err)
	}
}
