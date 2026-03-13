package kubemanager

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitForPodsRunning polls until all pods in the namespace are Running, or timeout is reached.
func (c *Client) WaitForPodsRunning(ctx context.Context, namespace string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("list pods in %q: %w", namespace, err)
		}

		if len(pods.Items) > 0 {
			allRunning := true
			for _, pod := range pods.Items {
				if pod.Status.Phase != corev1.PodRunning {
					allRunning = false
					break
				}
			}
			if allRunning {
				fmt.Printf("All %d pod(s) in %q are Running\n", len(pods.Items), namespace)
				return nil
			}
		}

		fmt.Printf("Waiting for pods in %q to be Running...\n", namespace)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for pods in %q to be Running", namespace)
}
