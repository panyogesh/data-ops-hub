// Install the following
// helm repo add crossplane-stable https://charts.crossplane.io/stable
// helm repo update
// helm install crossplane crossplane-stable/crossplane \
//   --namespace crossplane-system \
//   --create-namespace \
//   --kube-context kind-crossplane-dev

// Only install if not present

package helmmanager

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
	Chart     string
	RepoName  string
	RepoURL   string
	Namespace string
	Context   string
}

func NewHelmManager(cfg *config.Config) *HelmConfig {
	return &HelmConfig{
		Chart:     cfg.Helm.CrossPlane.Chart,
		RepoName:  cfg.Helm.CrossPlane.RepoName,
		RepoURL:   cfg.Helm.CrossPlane.RepoURL,
		Namespace: cfg.Helm.CrossPlane.Namespace,
		Context:   cfg.KubeConfig.Context,
	}
}

func (c *HelmConfig) String() string {
	return fmt.Sprintf("Chart: %s, RepoName: %s, RepoURL: %s, Namespace: %s, Context: %s",
		c.Chart, c.RepoName, c.RepoURL, c.Namespace, c.Context)
}

// isInstalled checks if the crossplane helm release is already present.
func isInstalled(ctx context.Context, namespace, kubeContext string) bool {
	args := []string{"status", crossplaneReleaseName, "-n", namespace}
	if kubeContext != "" {
		args = append(args, "--kube-context", kubeContext)
	}
	cmd := exec.CommandContext(ctx, "helm", args...)
	return cmd.Run() == nil
}

// AddRepo runs `helm repo add <name> <url>`.
func (c *HelmConfig) AddRepo(ctx context.Context) error {
	out, err := exec.CommandContext(ctx, "helm", "repo", "add", c.RepoName, c.RepoURL).CombinedOutput()
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

// UpdateRepo runs `helm repo update`.
func UpdateRepo(ctx context.Context) error {
	out, err := exec.CommandContext(ctx, "helm", "repo", "update").CombinedOutput()
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
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
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}

func (c *HelmConfig) InstallCrossplane(ctx context.Context) error {
	ns := c.Namespace
	if ns == "" {
		ns = crossplaneNamespace
	}

	if isInstalled(ctx, ns, c.Context) {
		fmt.Println("Crossplane is already installed")
		return nil
	}

	fmt.Println("========= repo add crossplane-stable =========")
	if err := c.AddRepo(ctx); err != nil {
		return errors.Wrap(err, "failed to add Crossplane repo")
	}

	fmt.Println("========= repo update =========")
	if err := UpdateRepo(ctx); err != nil {
		return errors.Wrap(err, "failed to update Helm repos")
	}

	fmt.Println("========= install crossplane =========")
	chart := fmt.Sprintf("%s/%s", c.RepoName, c.Chart)
	if err := InstallChart(ctx, crossplaneReleaseName, chart, ns, true, c.Context); err != nil {
		return errors.Wrap(err, "failed to install Crossplane")
	}
	return nil
}

func (c *HelmConfig) UninstallCrossplane(ctx context.Context) error {
	ns := c.Namespace
	if ns == "" {
		ns = crossplaneNamespace
	}

	if !isInstalled(ctx, ns, c.Context) {
		fmt.Println("Crossplane is not installed")
		return nil
	}

	fmt.Println("Uninstalling Crossplane")
	args := []string{"uninstall", crossplaneReleaseName, "-n", ns}
	if c.Context != "" {
		args = append(args, "--kube-context", c.Context)
	}
	out, err := exec.CommandContext(ctx, "helm", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
