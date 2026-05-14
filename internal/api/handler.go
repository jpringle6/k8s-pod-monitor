package api

import (
	"encoding/json"
	"net/http"

	"github.com/jpringle6/k8s-pod-monitor/internal/service"
)

// HTTPHandler implements ServerInterface and bridges HTTP requests to Service
type HTTPHandler struct {
	svc *service.Service
}

// NewHTTPHandler creates a new HTTPHandler instance
func NewHTTPHandler(svc *service.Service) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

// GetHealth implements ServerInterface.GetHealth
func (h *HTTPHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.svc.GetHealth(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetPodMetrics implements ServerInterface.GetPodMetrics
func (h *HTTPHandler) GetPodMetrics(w http.ResponseWriter, r *http.Request, params GetPodMetricsParams) {
	namespace := "default"
	if params.Namespace != nil {
		namespace = *params.Namespace
	}

	metrics, err := h.svc.GetPodMetrics(r.Context(), namespace)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Code: 500, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetNodeMetrics implements ServerInterface.GetNodeMetrics
func (h *HTTPHandler) GetNodeMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.svc.GetNodeMetrics(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Code: 500, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetEvents implements ServerInterface.GetEvents
func (h *HTTPHandler) GetEvents(w http.ResponseWriter, r *http.Request, params GetEventsParams) {
	namespace := "default"
	if params.Namespace != nil {
		namespace = *params.Namespace
	}

	events, err := h.svc.GetEvents(r.Context(), namespace)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Code: 500, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetAllMetrics implements ServerInterface.GetAllMetrics
func (h *HTTPHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request, params GetAllMetricsParams) {
	namespace := "default"
	if params.Namespace != nil {
		namespace = *params.Namespace
	}

	metrics, err := h.svc.GetAllMetrics(r.Context(), namespace)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Code: 500, Message: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
