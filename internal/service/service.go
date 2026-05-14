package service

import (
	"context"
	"time"

	"github.com/jpringle6/k8s-pod-monitor/internal/api"
	"github.com/jpringle6/k8s-pod-monitor/internal/logger"
	"github.com/jpringle6/k8s-pod-monitor/internal/monitoring"
)

// Service handles all business logic for the API
type Service struct {
	k8sClient *monitoring.K8sClient
	logger    logger.AsyncLogger
}

// NewService creates a new Service instance
func NewService(k8sClient *monitoring.K8sClient, log logger.AsyncLogger) *Service {
	return &Service{
		k8sClient: k8sClient,
		logger:    log,
	}
}

// GetHealth returns the health status of the service
func (s *Service) GetHealth(ctx context.Context) *api.HealthStatus {
	s.logger.Info("GetHealth called")
	return &api.HealthStatus{
		Status:    api.Healthy,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
}

// GetPodMetrics returns pod metrics for a given namespace
func (s *Service) GetPodMetrics(ctx context.Context, namespace string) (*api.PodMetricsResponse, error) {
	s.logger.Info("GetPodMetrics called for namespace: " + namespace)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pods, err := s.k8sClient.GetPodMetrics(ctx, namespace)
	if err != nil {
		s.logger.Error("Failed to get pod metrics: " + err.Error())
		return nil, err
	}

	return &api.PodMetricsResponse{
		Pods:      pods,
		Namespace: namespace,
		Timestamp: time.Now(),
	}, nil
}

// GetNodeMetrics returns metrics for all nodes
func (s *Service) GetNodeMetrics(ctx context.Context) (*api.NodeMetricsResponse, error) {
	s.logger.Info("GetNodeMetrics called")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	nodes, err := s.k8sClient.GetNodeMetrics(ctx)
	if err != nil {
		s.logger.Error("Failed to get node metrics: " + err.Error())
		return nil, err
	}

	return &api.NodeMetricsResponse{
		Nodes:     nodes,
		Timestamp: time.Now(),
	}, nil
}

// GetEvents returns events for a given namespace
func (s *Service) GetEvents(ctx context.Context, namespace string) (*api.EventsResponse, error) {
	s.logger.Info("GetEvents called for namespace: " + namespace)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	events, err := s.k8sClient.GetEvents(ctx, namespace)
	if err != nil {
		s.logger.Error("Failed to get events: " + err.Error())
		return nil, err
	}

	return &api.EventsResponse{
		Events:    events,
		Namespace: namespace,
		Timestamp: time.Now(),
	}, nil
}

// GetAllMetrics returns all metrics (pods, nodes, events) for a namespace
func (s *Service) GetAllMetrics(ctx context.Context, namespace string) (*api.AllMetricsResponse, error) {
	s.logger.Info("GetAllMetrics called for namespace: " + namespace)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Get all metrics concurrently
	podsChan := make(chan interface{}, 1)
	nodesChan := make(chan interface{}, 1)
	eventsChan := make(chan interface{}, 1)

	go func() {
		pods, err := s.k8sClient.GetPodMetrics(ctx, namespace)
		if err != nil {
			s.logger.Error("Failed to get pods in concurrent call: " + err.Error())
			podsChan <- nil
		} else {
			podsChan <- pods
		}
	}()

	go func() {
		nodes, err := s.k8sClient.GetNodeMetrics(ctx)
		if err != nil {
			s.logger.Error("Failed to get nodes in concurrent call: " + err.Error())
			nodesChan <- nil
		} else {
			nodesChan <- nodes
		}
	}()

	go func() {
		events, err := s.k8sClient.GetEvents(ctx, namespace)
		if err != nil {
			s.logger.Error("Failed to get events in concurrent call: " + err.Error())
			eventsChan <- nil
		} else {
			eventsChan <- events
		}
	}()

	pods := <-podsChan
	nodes := <-nodesChan
	events := <-eventsChan

	var podList []api.PodMetric
	var nodeList []api.NodeMetric
	var eventList []api.Event

	if pods != nil {
		podList = pods.([]api.PodMetric)
	}
	if nodes != nil {
		nodeList = nodes.([]api.NodeMetric)
	}
	if events != nil {
		eventList = events.([]api.Event)
	}

	return &api.AllMetricsResponse{
		Pods:      podList,
		Nodes:     nodeList,
		Events:    eventList,
		Namespace: &namespace,
		Timestamp: time.Now(),
	}, nil
}
