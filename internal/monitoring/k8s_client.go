package monitoring

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsClientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8sClient wraps Kubernetes client operations
type K8sClient struct {
	clientset        kubernetes.Interface
	metricsClientset metricsClientset.Interface
}

// NewK8sClient creates a new Kubernetes client
// It first tries to use in-cluster config, then falls back to kubeconfig
func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (for running in pod)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config: %w", err)
		}
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create metrics clientset
	metricsClientset, err := metricsClientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics clientset: %w", err)
	}

	return &K8sClient{
		clientset:        clientset,
		metricsClientset: metricsClientset,
	}, nil
}

// GetPodMetrics retrieves metrics for all pods in a namespace
func (k *K8sClient) GetPodMetrics(ctx context.Context, namespace string) ([]PodMetrics, error) {
	var podMetrics []PodMetrics

	// Get pods
	pods, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Build pod metrics list
	now := time.Now()
	for _, pod := range pods.Items {
		pm := PodMetrics{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Age:       formatAge(pod.CreationTimestamp.Time),
			Timestamp: now,
		}

		// Get restart count from all containers
		for _, containerStatus := range pod.Status.ContainerStatuses {
			pm.Restarts += containerStatus.RestartCount
		}

		// For now, estimate CPU and memory from requests/limits
		if pod.Spec.Containers != nil && len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[0]
			if container.Resources.Requests != nil {
				if cpu, ok := container.Resources.Requests["cpu"]; ok {
					pm.CPU = cpu.String()
				}
				if memory, ok := container.Resources.Requests["memory"]; ok {
					pm.Memory = memory.String()
				}
			}
		}

		podMetrics = append(podMetrics, pm)
	}

	return podMetrics, nil
}

// GetNodeMetrics retrieves metrics for all nodes
func (k *K8sClient) GetNodeMetrics(ctx context.Context) ([]NodeMetrics, error) {
	var nodeMetrics []NodeMetrics

	// Get nodes
	nodes, err := k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Count pods per node
	pods, _ := k.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	podCountMap := make(map[string]int)
	if pods != nil {
		for _, pod := range pods.Items {
			if pod.Spec.NodeName != "" {
				podCountMap[pod.Spec.NodeName]++
			}
		}
	}

	now := time.Now()
	for _, node := range nodes.Items {
		status := "NotReady"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}

		nm := NodeMetrics{
			Name:      node.Name,
			Status:    status,
			Pods:      podCountMap[node.Name],
			Timestamp: now,
		}

		// Get allocatable resources from node status
		if node.Status.Allocatable != nil {
			if cpu, ok := node.Status.Allocatable["cpu"]; ok {
				nm.CPU = cpu.String()
			}
			if memory, ok := node.Status.Allocatable["memory"]; ok {
				nm.Memory = memory.String()
			}
		}

		nodeMetrics = append(nodeMetrics, nm)
	}

	return nodeMetrics, nil
}

// GetEvents retrieves recent events from a namespace
func (k *K8sClient) GetEvents(ctx context.Context, namespace string) ([]Event, error) {
	var events []Event

	eventList, err := k.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	for _, e := range eventList.Items {
		event := Event{
			Name:      e.Name,
			Namespace: e.Namespace,
			Kind:      e.InvolvedObject.Kind,
			Message:   e.Message,
			Reason:    e.Reason,
			Type:      e.Type,
			Timestamp: e.LastTimestamp.Time,
		}
		events = append(events, event)
	}

	return events, nil
}

// formatAge converts a creation timestamp to a human-readable age
func formatAge(creationTime time.Time) string {
	age := time.Since(creationTime)

	if age.Hours() > 24 {
		days := int(age.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if age.Hours() > 1 {
		return fmt.Sprintf("%dh", int(age.Hours()))
	}
	if age.Minutes() > 1 {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	}
	return fmt.Sprintf("%ds", int(age.Seconds()))
}
