package kubemanager

import (
	"context"
	"fmt"
	"os/exec"
)

// Apply runs `kubectl apply -f <filePath>` against the given context.
func Apply(ctx context.Context, filePath, kubeContext string) error {
	args := []string{"apply", "-f", filePath}
	if kubeContext != "" {
		args = append(args, "--context", kubeContext)
	}
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return fmt.Errorf("kubectl apply -f %s: %w", filePath, err)
	}
	return nil
}

// GetProviders runs `kubectl get providers` and prints the output.
func GetProviders(ctx context.Context, kubeContext string) error {
	args := []string{"get", "providers"}
	if kubeContext != "" {
		args = append(args, "--context", kubeContext)
	}
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return fmt.Errorf("kubectl get providers: %w", err)
	}
	return nil
}

// PrintCurrentContext runs `kubectl config current-context` and prints the result.
func PrintCurrentContext(ctx context.Context) {
	out, err := exec.CommandContext(ctx, "kubectl", "config", "current-context").CombinedOutput()
	if err != nil {
		fmt.Printf("kubectl config current-context: %v\n", err)
		return
	}
	fmt.Printf("Current context: %s\n", string(out))
}
