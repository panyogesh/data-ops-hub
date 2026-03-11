// Install the following
// helm repo add crossplane-stable https://charts.crossplane.io/stable
// helm repo update
// helm install crossplane crossplane-stable/crossplane \
//   --namespace crossplane-system \
//   --create-namespace \
//   --kube-context kind-crossplane-dev

// Only install if not present

package helmops

import (
	"context"
	"fmt"
	"os/exec"

	"example.com/gridnode/pkg/config"
	"github.com/pkg/errors"
)

const (
	crossplaneReleaseName = "crossplane"
	crossplaneNamespace   = "crossplane-system"
)

type HelmConfig struct {
	Chart    string `json:"chart"`
	RepoName string `json:"repoName"`
	Context  string `json:"context"`
}

func NewHelmConfig(cfg *config.Config) *HelmConfig {
	return &HelmConfig{
		Chart:    cfg.Helm.CrossPlane.Chart,
		RepoName: cfg.Helm.CrossPlane.RepoName,
		Context:  cfg.KubeConfig.Context,
	}
}

func (c *HelmConfig) String() string {
	return fmt.Sprintf("Chart: %s, RepoName: %s, Context: %s", c.Chart, c.RepoName, c.Context)
}

// isInstalled checks if the crossplane helm release is already present.
func isInstalled(ctx context.Context, kubeContext string) bool {
	args := []string{"status", crossplaneReleaseName, "-n", crossplaneNamespace}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	cmd := exec.CommandContext(ctx, "helm", args...)
	return cmd.Run() == nil
}

// InstallChart runs `helm install` for the given release.
func InstallChart(ctx context.Context, release, chart, namespace string, createNamespace bool, kubeContext string) error {
	args := []string{"install", release, chart, "-n", namespace}
	if createNamespace {
		args = append(args, "--create-namespace")
	}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	out, err := exec.CommandContext(ctx, "helm", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

func (c *HelmConfig) InstallCrossplane(ctx context.Context) error {
	if isInstalled(ctx, c.Context) {
		fmt.Println("Crossplane is already installed")
		return nil
	}

	fmt.Println("Installing Crossplane")
	err := InstallChart(ctx, crossplaneReleaseName, "crossplane-stable/crossplane", crossplaneNamespace, true, c.Context)
	if err != nil {
		return errors.Wrap(err, "failed to install Crossplane")
	}
	return nil
}

func (c *HelmConfig) UninstallCrossplane(ctx context.Context) error {
	if !isInstalled(ctx, c.Context) {
		fmt.Println("Crossplane is not installed")
		return nil
	}

	fmt.Println("Uninstalling Crossplane")
	args := []string{"uninstall", crossplaneReleaseName, "-n", crossplaneNamespace}
	if c.Context != "" {
		args = append(args, "--kube-context", c.Context)
	}
	out, err := exec.CommandContext(ctx, "helm", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
