package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/gridnode/pkg/config"
	"example.com/gridnode/pkg/crossplane"
	"example.com/gridnode/pkg/helmmanager"
	"example.com/gridnode/pkg/kubemanager"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Kubernetes client
	kubeClient, err := kubemanager.NewClient(kubemanager.Options{
		KubeconfigPath: cfg.KubeConfig.KubeConfigPath,
		Context:        cfg.KubeConfig.Context,
	})
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Verify cluster connectivity
	if err := kubeClient.Ping(); err != nil {
		log.Fatalf("Cannot reach cluster: %v", err)
	}

	// helm repo add + repo update + helm install crossplane
	helmManager := helmmanager.NewHelmManager(cfg)
	fmt.Printf("Helm config: %s\n", helmManager)

	if err := helmManager.InstallCrossplane(ctx); err != nil {
		log.Fatalf("Failed to install Crossplane: %v", err)
	}

	// Wait for crossplane pods to be Running
	fmt.Println("========= Sleep for 5 seconds. Allow cross-plane pods to come up =========")
	fmt.Println("========= Check if the installation is RUNNING =========")
	ns := cfg.Helm.CrossPlane.Namespace
	if err := kubeClient.WaitForPodsRunning(ctx, ns, 2*time.Minute); err != nil {
		log.Fatalf("Crossplane pods did not become ready: %v", err)
	}

	// kubectl get pods -n crossplane-system
	pods, err := kubeClient.ListPods(ctx, ns)
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}
	fmt.Printf("Pods in %q:\n", ns)
	for _, pod := range pods {
		fmt.Printf("  %-40s %s\n", pod.Name, pod.Status.Phase)
	}

	// kubectl config current-context
	fmt.Println("========= Checking the current context =========")
	kubemanager.PrintCurrentContext(ctx)

	// Build provider from config and display its manifest
	provider, err := crossplane.NewProvider(*cfg)
	if err != nil {
		log.Fatalf("Failed to build provider config: %v", err)
	}
	fmt.Printf("========= Provider: %s =========\n", provider)
	fmt.Printf("Generated provider manifest:\n%s\n", provider.YAML())

	// kubectl apply -f 00-provider.yaml
	fmt.Println("========= Apply provider YAML =========")
	if err := kubemanager.Apply(ctx, cfg.YAMLPaths.Provider, cfg.KubeConfig.Context); err != nil {
		log.Fatalf("Failed to apply provider YAML: %v", err)
	}

	// kubectl apply -f 01-gcp-secret.yaml
	fmt.Println("========= Apply GCP secret YAML =========")
	if err := kubemanager.Apply(ctx, cfg.YAMLPaths.Secret, cfg.KubeConfig.Context); err != nil {
		log.Fatalf("Failed to apply GCP secret YAML: %v", err)
	}

	// kubectl get providers
	fmt.Println("========= Check whether services are operational =========")
	if err := kubemanager.GetProviders(ctx, cfg.KubeConfig.Context); err != nil {
		log.Printf("Warning: kubectl get providers: %v", err)
	}
}
