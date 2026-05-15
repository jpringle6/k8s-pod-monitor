
EOF.PHONY: help setup build test clean docker-build docker-load deploy logs generate

help:
	@echo "Available commands:"
	@echo "  make setup          - Setup development environment"
	@echo "  make generate       - Generate code from OpenAPI spec"
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
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

.PHONY: generate
generate:
	@echo "Generating code from OpenAPI spec..."
	oapi-codegen -package api -generate types,chi-server,client -o internal/api/generated.go api/openapi.yaml
	@echo "✓ Code generated successfully"

build: generate
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
