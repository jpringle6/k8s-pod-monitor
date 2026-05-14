package monitoring

import "time"

// PodMetrics represents metrics for a single pod
type PodMetrics struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Status     string    `json:"status"`
	CPU        string    `json:"cpu"`
	Memory     string    `json:"memory"`
	Restarts   int32     `json:"restarts"`
	Age        string    `json:"age"`
	Timestamp  time.Time `json:"timestamp"`
}

// NodeMetrics represents metrics for a single node
type NodeMetrics struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CPU       string    `json:"cpu"`
	Memory    string    `json:"memory"`
	Pods      int       `json:"pods"`
	Timestamp time.Time `json:"timestamp"`
}

// Event represents a Kubernetes event
type Event struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Kind       string    `json:"kind"`
	Message    string    `json:"message"`
	Reason     string    `json:"reason"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
}

// MetricsResponse wraps metrics data
type MetricsResponse struct {
	Pods      []PodMetrics `json:"pods"`
	Nodes     []NodeMetrics `json:"nodes"`
	Events    []Event `json:"events"`
	Timestamp time.Time `json:"timestamp"`
	Error     string `json:"error,omitempty"`
}
