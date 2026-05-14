package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jpringle6/k8s-pod-monitor/internal/api"
	"github.com/jpringle6/k8s-pod-monitor/internal/logger"
	"github.com/jpringle6/k8s-pod-monitor/internal/monitoring"
	"github.com/jpringle6/k8s-pod-monitor/internal/service"
)

func main() {
	// Initialize logger
	log, err := logger.NewLogger(100)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Close()

	log.Info("Initializing Kubernetes client...")

	// Initialize Kubernetes client
	k8sClient, err := monitoring.NewK8sClient()
	if err != nil {
		log.Error("Failed to initialize Kubernetes client: " + err.Error())
		panic(err)
	}

	log.Info("Kubernetes client initialized successfully")

	// Create service layer (business logic)
	svc := service.NewService(k8sClient, log)

	// Create HTTP handler (bridges HTTP to service)
	httpHandler := api.NewHTTPHandler(svc)

	// Create Chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register handlers from generated API code
	api.HandlerFromMux(httpHandler, r)

	// Start server
	log.Info("Starting monitor service on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Error("Server error: " + err.Error())
		panic(err)
	}
}
