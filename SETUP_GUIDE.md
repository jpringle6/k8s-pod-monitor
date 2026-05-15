# K8s Pod Monitor - Complete Setup Guide

A comprehensive guide to setting up the development environment and deploying the Kubernetes pod monitoring service.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Windows Setup](#windows-setup)
3. [WSL2 Setup](#wsl2-setup)
4. [Kubernetes Cluster Setup](#kubernetes-cluster-setup)
5. [Project Structure](#project-structure)
6. [Project Development](#project-development)
7. [Building and Deployment](#building-and-deployment)
8. [Testing](#testing)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before starting, ensure you have:
- Windows 10/11 with admin access
- Administrator privileges in PowerShell
- Internet connection for downloading components
- At least 16GB RAM (8GB minimum for WSL2 + K8s)

---

## Windows Setup

### 1. Install Chocolatey Package Manager

Chocolatey is a package manager for Windows that makes installing tools easy.

**In PowerShell (as Administrator):**

```powershell
# Set execution policy for current session
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process -Force

# Install Chocolatey
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Verify installation
choco --version
```

### 2. Install Docker Desktop

Docker Desktop provides containerization for development and deployment.

```powershell
# Install Docker Desktop
choco install docker-desktop -y

# Wait for installation to complete
# Restart computer when prompted
```

**Note:** Docker Desktop will initially use Hyper-V backend, but we'll switch it to WSL2 later.

### 3. Install Git

Git is required for version control.

```powershell
choco install git -y

# Verify
git --version
```

---

## WSL2 Setup

WSL2 (Windows Subsystem for Linux 2) provides a native Linux environment on Windows.

### 1. Install WSL2 with Ubuntu

```powershell
# In PowerShell as Administrator
wsl --install -d Ubuntu

# This will:
# 1. Enable WSL2 feature
# 2. Download Ubuntu
# 3. Install it
# 4. Prompt for restart
```

**When Ubuntu launches:**
- Create a username: `justinp` (or your preferred name)
- Create a password: (your choice)

### 2. Update Ubuntu System

```bash
# In Ubuntu terminal
sudo apt update
sudo apt upgrade -y
```

### 3. Configure Docker for WSL2

**Stop Docker Desktop first:**
- Click Docker icon in system tray
- Select "Quit Docker Desktop"

**Then in Ubuntu terminal:**

```bash
# Install Docker in Ubuntu
sudo apt install -y docker.io

# Add your user to docker group
sudo usermod -aG docker $USER

# Apply new group membership
newgrp docker

# Start Docker service
sudo service docker start

# Enable Docker to start on boot
sudo systemctl enable docker

# Verify Docker works
docker ps
```

### 4. Install Kubernetes Tools

#### Install kubectl (Kubernetes CLI)

```bash
# Download latest kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Make it executable
chmod +x kubectl

# Install it
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Verify
kubectl version --client
```

#### Install Kind (Kubernetes in Docker)

```bash
# Download Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64

# Make it executable
chmod +x ./kind

# Install it
sudo mv ./kind /usr/local/bin/kind

# Verify
kind version
```

### 5. Install Go

```bash
# Install Go
sudo apt install -y golang-go

# Verify
go version
```

### 6. Install Additional Tools

```bash
# Essential development tools
sudo apt install -y \
  build-essential \
  curl \
  wget \
  git \
  make
```

### 7. Set Up VSCode Remote WSL Extension

1. Open VSCode on Windows
2. Install extension: **Remote - WSL** (by Microsoft)
3. Press `Ctrl + Shift + P` and search: `Remote-WSL: New Window`
4. Select **Ubuntu** from the dropdown
5. VSCode reopens connected to WSL Ubuntu

---

## Kubernetes Cluster Setup

### 1. Create Kind Cluster Configuration

```bash
# Create configuration file
cat > kind-config.yaml <<'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
EOF
```

### 2. Create the Cluster

```bash
# Create the Kind cluster
kind create cluster --name dev --config kind-config.yaml

# Wait for all nodes to be ready (2-3 minutes)
kubectl wait --for=condition=Ready node --all --timeout=300s

# Verify all nodes are ready
kubectl get nodes
```

**Expected output:**
```
NAME                STATUS   ROLES           AGE   VERSION
dev-control-plane   Ready    control-plane   2m    v1.27.3
dev-worker          Ready    <none>          2m    v1.27.3
dev-worker2         Ready    <none>          2m    v1.27.3
dev-worker3         Ready    <none>          2m    v1.27.3
```

### 3. Create Monitoring Namespace

```bash
# Create namespace for our monitoring service
kubectl create namespace monitoring

# Deploy test applications to monitor
kubectl create deployment nginx --image=nginx --replicas=3 -n monitoring
kubectl create deployment redis --image=redis --replicas=2 -n monitoring

# Verify pods are running
kubectl get pods -n monitoring
```

---

## Project Structure

### 1. Create GitHub Repository

Go to: https://github.com/new

**Settings:**
- Repository name: `k8s-pod-monitor`
- Description: `Kubernetes pod resource monitoring service with REST API`
- Visibility: `Public`
- Initialize with:
  - ✓ Add README.md
  - ✓ Add .gitignore (select Go)
  - ✓ Add license (select MIT)

### 2. Clone Repository Locally

```bash
# Navigate to projects directory
mkdir -p ~/projects
cd ~/projects

# Clone the repository (replace with YOUR username)
git clone https://github.com/yourusername/k8s-pod-monitor.git

# Navigate into project
cd k8s-pod-monitor
```

### 3. Create Project Directory Structure

```bash
# Create all necessary directories
mkdir -p cmd/monitor-service
mkdir -p cmd/monitor-cli
mkdir -p internal/monitoring
mkdir -p internal/api
mkdir -p internal/config
mkdir -p k8s
mkdir -p examples
mkdir -p tests/{unit,integration}
mkdir -p bin
```

### 4. Fix File Permissions (If Needed)

If you created files as root:

```bash
# Change ownership to your user
sudo chown -R justinp:justinp ~/projects/k8s-pod-monitor

# Make files writable
sudo chmod -R u+w ~/projects/k8s-pod-monitor
```

---

## Project Development

### 1. Initialize Go Module

```bash
# Navigate to project directory
cd ~/projects/k8s-pod-monitor

# Initialize Go module
go mod init github.com/yourusername/k8s-pod-monitor

# Add Kubernetes dependencies
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/metrics@latest

# Add logging dependency
go get github.com/sirupsen/logrus@latest

# Clean up dependencies
go mod tidy

# Verify go.mod
cat go.mod
```

### 2. Create Dockerfile

```bash
cat > Dockerfile <<'EOF'
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o monitor ./cmd/monitor-service

# Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/monitor .

EXPOSE 8080

CMD ["./monitor"]
EOF
```

**Key points:**
- Multi-stage build reduces final image size
- `go.mod go.sum ./` explicitly copies files (not `go.*`)
- Alpine base image is lightweight and secure
- Exposes port 8080 for REST API

### 3. Create Makefile

```bash
cat > Makefile <<'EOF'
.PHONY: help setup build test clean docker-build docker-load deploy logs

help:
	@echo "Available commands:"
	@echo "  make setup          - Setup development environment"
	@echo "  make build          - Build binaries"
	@echo "  make test           - Run tests"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-load    - Load image into Kind"
	@echo "  make deploy         - Deploy to K8s cluster"
	@echo "  make logs           - View service logs"
	@echo "  make clean          - Clean up"

setup:
	go mod download
	go mod tidy

build:
	go build -o bin/monitor ./cmd/monitor-service
	go build -o bin/monitor-cli ./cmd/monitor-cli

test:
	go test -v ./...

docker-build:
	docker build -t k8s-monitor:latest .

docker-load: docker-build
	kind load docker-image k8s-monitor:latest --name dev

deploy: docker-load
	kubectl apply -f k8s/rbac.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl -n monitoring rollout status deployment/k8s-monitor

logs:
	kubectl -n monitoring logs -f deployment/k8s-monitor

clean:
	kubectl delete -f k8s/ 2>/dev/null || true
	go clean
	rm -rf bin/

.DEFAULT_GOAL := help
EOF
```

### 4. Create Kubernetes Manifests

#### RBAC (Role-Based Access Control)

```bash
cat > k8s/rbac.yaml <<'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-monitor
  namespace: monitoring
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-monitor
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "events"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-monitor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-monitor
subjects:
- kind: ServiceAccount
  name: k8s-monitor
  namespace: monitoring
EOF
```

#### Deployment

```bash
cat > k8s/deployment.yaml <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-monitor
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8s-monitor
  template:
    metadata:
      labels:
        app: k8s-monitor
    spec:
      serviceAccountName: k8s-monitor
      containers:
      - name: monitor
        image: k8s-monitor:latest
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
EOF
```

#### Service

```bash
cat > k8s/service.yaml <<'EOF'
apiVersion: v1
kind: Service
metadata:
  name: k8s-monitor
  namespace: monitoring
spec:
  selector:
    app: k8s-monitor
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
EOF
```

### 5. Create Go Source Files

#### Service Main

```bash
cat > cmd/monitor-service/main.go <<'EOF'
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Health struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/metrics/pods", podsHandler)
	
	log.Println("Starting monitor service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Health{
		Status:    "healthy",
		Timestamp: time.Now(),
	})
}

func podsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "TODO: Implement pod metrics",
	})
}
EOF
```

#### CLI Main

```bash
cat > cmd/monitor-cli/main.go <<'EOF'
package main

import (
	"fmt"
)

func main() {
	fmt.Println("K8s Pod Monitor CLI")
	fmt.Println("TODO: Implement CLI")
}
EOF
```

---

## Building and Deployment

### 1. Build Service Locally

```bash
# Build Go binaries
make build

# Verify binaries were created
ls -la bin/
```

### 2. Build Docker Image

```bash
# Build Docker image
make docker-build

# Verify image was created
docker images | grep k8s-monitor
```

### 3. Load Image into Kind

```bash
# Load image into Kind cluster
make docker-load

# This distributes the image to all nodes
```

### 4. Deploy to Kubernetes

```bash
# Deploy all manifests
make deploy

# Watch deployment status
kubectl rollout status deployment/k8s-monitor -n monitoring

# Verify pods are running
kubectl get pods -n monitoring
```

### 5. View Logs

```bash
# View service logs
make logs

# Should show: "Starting monitor service on :8080"
```

### 6. Upgrading the Service Image

When you update the code and need to deploy a new version:

```bash
# 1. Build the image with a NEW VERSION TAG (not "latest")
#    This is important - Kind caches "latest", so use incremented versions
cd ~/projects/k8s-pod-monitor
docker build -t k8s-pod-monitor:v3 .

# 2. Load it into Kind with the new tag
kind load docker-image k8s-pod-monitor:v3 --name dev

# 3. Update the deployment to use the new tag
#    This automatically restarts the pod with the new image
kubectl set image deployment/k8s-monitor monitor=k8s-pod-monitor:v3

# 4. Watch the rollout
kubectl get pods -w

# 5. Verify the new code is running
kubectl logs k8s-monitor -f
```

**Key Points:**
- Always use version tags (v1, v2, v3, etc.) instead of `latest`
- `latest` tag gets cached by Kind and won't update
- Version tags force a fresh image load
- `kubectl set image` automatically rolls out the new version
- Use `kubectl get pods -w` to watch the pod restart
- Check logs to verify the new code is running

**Example Upgrade Workflow:**
```bash
# Edit your code
vim cmd/monitor-service/main.go

# Build and deploy with new version
docker build -t k8s-pod-monitor:v4 .
kind load docker-image k8s-pod-monitor:v4 --name dev
kubectl set image deployment/k8s-monitor monitor=k8s-pod-monitor:v4

# Wait for pod to restart
kubectl get pods -w

# Verify
kubectl logs k8s-monitor -f
```

---

## Testing

### 1. Port Forward Service

```bash
# In a terminal, port forward the service
kubectl port-forward svc/k8s-monitor 8080:8080 -n monitoring &

# Wait a moment for port forward to start
sleep 2
```

### 2. Test Health Endpoint

```bash
# Test health check
curl http://localhost:8080/health

# Expected response:
# {"Status":"healthy","Timestamp":"2025-01-XX..."}
```

### 3. Test Metrics Endpoint

```bash
# Test metrics endpoint
curl http://localhost:8080/api/v1/metrics/pods?namespace=monitoring

# Expected response:
# {"message":"TODO: Implement pod metrics"}
```

### 4. Stop Port Forward

```bash
# Stop port forward
kill %1
```

---

## Committing to GitHub

```bash
# Check git status
git status

# Add all files
git add .

# Commit with descriptive message
git commit -m "Initial project structure with working REST service

- Initialize Go module with Kubernetes dependencies
- Create REST API service with health and metrics endpoints
- Add Dockerfile for containerization (multi-stage build)
- Add Kubernetes manifests (RBAC, Deployment, Service)
- Add Makefile for build and deployment automation
- Deploy to Kind cluster successfully
- Service responds to health and metrics endpoints"

# Push to GitHub
git push origin main
```

---

## Troubleshooting

### Docker Permission Denied

```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker ps
```

### Cluster Health Issues

```bash
# Check cluster info
kubectl cluster-info

# Check node status
kubectl get nodes

# If issues persist, restart cluster
kind delete cluster --name dev
kind create cluster --name dev --config kind-config.yaml
```

### File Permission Issues (Created as Root)

```bash
# Fix ownership
sudo chown -R justinp:justinp ~/projects/k8s-pod-monitor

# Fix permissions
sudo chmod -R u+w ~/projects/k8s-pod-monitor
```

### Docker Image Not Loading

```bash
# Clean up old images
docker system prune -f

# Rebuild image
make clean
make docker-build
make docker-load
```

### Service Not Starting

```bash
# Check pod logs
kubectl logs deployment/k8s-monitor -n monitoring

# Check pod events
kubectl describe pod -l app=k8s-monitor -n monitoring

# Check if pod is running
kubectl get pods -n monitoring
```

### WSL Docker Daemon Not Running

```bash
# Start Docker service
sudo service docker start

# Make it start on boot
sudo systemctl enable docker

# Verify
docker ps
```

---

## Environment Summary

| Component | Version | Purpose |
|-----------|---------|---------|
| Windows | 10/11 | Host OS |
| WSL2 + Ubuntu | 22.04 | Linux development environment |
| Docker | Latest | Container runtime |
| Kubernetes (Kind) | 1.27.3 | Container orchestration |
| Go | 1.21 | Application development |
| kubectl | 1.28 | Kubernetes CLI |
| Kind | 0.20.0 | Local K8s cluster |
| Git | Latest | Version control |
| Make | Latest | Build automation |
| VSCode | Latest | IDE (with Remote WSL) |

---

## Key Learnings

### Architecture Pattern
- **Event-driven**: Async processing pattern
- **Cloud-native**: Runs as Kubernetes service
- **RESTful**: HTTP API for external access
- **Containerized**: Docker for consistent deployment

### Development Best Practices
- Use WSL2 for native Linux environment
- Keep all commands declarative (Makefile, K8s manifests)
- Port-forward for testing, not exposing services publicly
- Use namespaces for logical separation
- Apply RBAC for principle of least privilege
- Use environment variables for configuration

### Common Pitfalls to Avoid
- Don't use Windows paths in WSL commands
- Don't run as root permanently (use sudo for specific commands)
- Don't hardcode configuration (use env vars, ConfigMaps)
- Don't expose services without authentication
- Don't commit binary files or credentials

---

## Next Steps

1. Implement actual pod metrics collection from Kubernetes API
2. Add database storage (SQLite) for metrics history
3. Implement alerting system
4. Add web UI for visualization
5. Add comprehensive error handling
6. Add unit and integration tests
7. Add performance monitoring
8. Deploy to cloud Kubernetes cluster

---

End of Setup Guide
