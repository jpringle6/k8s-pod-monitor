#!/bin/bash

# Test script for K8s Pod Monitor REST API
# Usage: ./examples/test-api.sh

set -e

BASE_URL="http://localhost:8080"
NAMESPACE="${1:-monitoring}"

echo "=========================================="
echo "K8s Pod Monitor API Test Suite"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo "Namespace: $NAMESPACE"
echo ""

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test health endpoint
echo -e "${BLUE}Testing /health endpoint...${NC}"
response=$(curl -s "$BASE_URL/health")
echo "$response" | jq .
echo ""

# Test pods endpoint
echo -e "${BLUE}Testing /api/v1/metrics/pods endpoint...${NC}"
response=$(curl -s "$BASE_URL/api/v1/metrics/pods?namespace=$NAMESPACE")
echo "$response" | jq .
echo ""

# Test nodes endpoint
echo -e "${BLUE}Testing /api/v1/metrics/nodes endpoint...${NC}"
response=$(curl -s "$BASE_URL/api/v1/metrics/nodes")
echo "$response" | jq .
echo ""

# Test events endpoint
echo -e "${BLUE}Testing /api/v1/metrics/events endpoint...${NC}"
response=$(curl -s "$BASE_URL/api/v1/metrics/events?namespace=$NAMESPACE")
echo "$response" | jq .
echo ""

# Test all metrics endpoint
echo -e "${BLUE}Testing /api/v1/metrics/all endpoint...${NC}"
response=$(curl -s "$BASE_URL/api/v1/metrics/all?namespace=$NAMESPACE")
echo "$response" | jq .
echo ""

echo -e "${GREEN}All tests completed!${NC}"
