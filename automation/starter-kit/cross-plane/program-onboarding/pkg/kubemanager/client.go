package kubemanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset  *kubernetes.Clientset
	Discovery  discovery.DiscoveryInterface
	RestConfig *rest.Config
	Source     string
}

type Options struct {
	KubeconfigPath string
	Context        string
	QPS            float32
	Burst          int
	Timeout        int
}

// NewClient returns a new Kubernetes client. Works for:
// InCluster
// Local KubeConfig
// GKE KubeConfig
func NewClient(opts Options) (*Client, error) {
	cfg, source, err := buildConfig(opts)
	if err != nil {
		return nil, err
	}

	if opts.QPS > 0 {
		cfg.QPS = opts.QPS
	}

	if opts.Burst > 0 {
		cfg.Burst = opts.Burst
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes Clientset: %w", err)
	}

	return &Client{
		Clientset:  cs,
		Discovery:  cs.Discovery(),
		RestConfig: cfg,
		Source:     source,
	}, nil
}

func buildFromKubeconfig(path, contextName string) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: path}

	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}

	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	cfg, err := clientCfg.ClientConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func buildConfig(opts Options) (*rest.Config, string, error) {
	if opts.KubeconfigPath != "" {
		cfg, err := buildFromKubeconfig(opts.KubeconfigPath, opts.Context)
		if err != nil {
			return nil, "", fmt.Errorf("load explicit kubeconfig %q: %w", opts.KubeconfigPath, err)
		}
		return cfg, "explicit-kubeconfig", nil
	}

	// Try in-cluster config
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, "incluster", nil
	}

	// 3) Then KUBECONFIG env
	if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
		cfg, err := buildFromKubeconfig(envKubeconfig, opts.Context)
		if err != nil {
			return nil, "", fmt.Errorf("load kubeconfig from KUBECONFIG env %q: %w", envKubeconfig, err)
		}
		return cfg, "env-kubeconfig", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("get user home directory: %w", err)
	}
	defaultKubeconfig := filepath.Join(home, ".kube", "config")
	if _, statErr := os.Stat(defaultKubeconfig); statErr == nil {
		cfg, err = buildFromKubeconfig(defaultKubeconfig, opts.Context)
		if err != nil {
			return nil, "", fmt.Errorf("load kubeconfig from default location %q: %w", defaultKubeconfig, err)
		}
		return cfg, "default-kubeconfig", nil
	}
	return nil, "", errors.New("no usable Kubernetes config found: not running in cluster, KUBECONFIG not set, and ~/.kube/config missing")
}

// Ping verifies we can talk to the cluster.
func (c *Client) Ping() error {
	_, err := c.Discovery.ServerVersion()
	if err != nil {
		return fmt.Errorf("cluster ping failed : %w", err)
	}
	return nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	list, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespacers error %w", err)
	}

	out := make([]string, len(list.Items))
	for i, ns := range list.Items {
		out[i] = ns.Name
	}
	return out, nil
}

// kubectl get pods -n crossplane-system --context kind-crossplane-dev
func (c *Client) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	list, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods in namespace %q: %w", namespace, err)
	}
	return list.Items, nil
}
