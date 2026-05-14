# Testing Guide for K8s Pod Monitor

## Prerequisites

- Kind cluster running with monitoring namespace
- Service deployed to K8s
- kubectl configured
- curl installed

## Manual Testing Steps

### 1. Check Service is Running

```bash
# Check if pod is running
kubectl get pods -n monitoring

# Should show k8s-monitor pod in Running state
```

### 2. Check Service Logs

```bash
# View current logs
kubectl logs deployment/k8s-monitor -n monitoring

# Watch logs in real-time
kubectl logs deployment/k8s-monitor -n monitoring -f
```

### 3. Port Forward Service

```bash
# In one terminal, port forward
kubectl port-forward svc/k8s-monitor 8080:8080 -n monitoring

# In another terminal, you're ready to test
```

### 4. Test Health Endpoint

```bash
# Test health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","timestamp":"2025-01-XX...","version":"1.0.0"}
```

### 5. Test Pod Metrics Endpoint

```bash
# Get pod metrics from monitoring namespace
curl "http://localhost:8080/api/v1/metrics/pods?namespace=monitoring"

# Expected response includes pods with metrics:
# {
#   "pods": [
#     {
#       "name": "nginx-xxx",
#       "namespace": "monitoring",
#       "status": "Running",
#       "cpu": "10m",
#       "memory": "2.5Mi",
#       "restarts": 0,
#       "age": "5m",
#       "timestamp": "2025-01-XX..."
#     }
#   ],
#   "namespace": "monitoring",
#   "timestamp": "2025-01-XX..."
# }
```

### 6. Test Node Metrics Endpoint

```bash
# Get metrics for all cluster nodes
curl http://localhost:8080/api/v1/metrics/nodes

# Expected response includes all nodes with metrics
```

### 7. Test Events Endpoint

```bash
# Get events from monitoring namespace
curl "http://localhost:8080/api/v1/metrics/events?namespace=monitoring"

# Expected response includes recent events
```

### 8. Test All Metrics Endpoint

```bash
# Get all metrics at once
curl "http://localhost:8080/api/v1/metrics/all?namespace=monitoring" | jq .

# This is the most comprehensive endpoint
```

## Using the Test Script

```bash
# Make script executable
chmod +x examples/test-api.sh

# Run tests
./examples/test-api.sh

# Or test specific namespace
./examples/test-api.sh custom-namespace
```

## Troubleshooting

### Pod Not Running

```bash
# Check pod status
kubectl describe pod -l app=k8s-monitor -n monitoring

# Check events
kubectl get events -n monitoring

# Check logs for errors
kubectl logs deployment/k8s-monitor -n monitoring
```

### Connection Refused

```bash
# Make sure port forward is running in another terminal
kubectl port-forward svc/k8s-monitor 8080:8080 -n monitoring

# Make sure service is created
kubectl get svc -n monitoring
```

### Metrics Not Available

```bash
# Metrics server might not be installed
# Check if metrics are available
kubectl get --raw /apis/metrics.k8s.io/v1beta1/nodes

# If not available, Kind might need metrics-server
# For Kind, metrics are usually pre-installed
```

### 401/403 Unauthorized

```bash
# Check RBAC permissions
kubectl get rolebinding,clusterrolebinding -n monitoring

# Check service account
kubectl get sa -n monitoring

# Verify service account has correct permissions
kubectl describe clusterrole k8s-monitor
```

## Performance Testing

### Load Testing

```bash
# Simple load test with ab (Apache Bench)
ab -n 100 -c 10 http://localhost:8080/health

# Or use wrk
wrk -t4 -c100 -d30s http://localhost:8080/health
```

### Latency Testing

```bash
# Measure response time
time curl "http://localhost:8080/api/v1/metrics/pods?namespace=monitoring" > /dev/null

# Or use curl -w for detailed timing
curl -w "\nTotal: %{time_total}s\n" http://localhost:8080/health
```

## Common Test Cases

### Test 1: Service Health

```bash
curl http://localhost:8080/health
```

**Expected:** Status is "healthy"

### Test 2: Pod Discovery

```bash
curl "http://localhost:8080/api/v1/metrics/pods?namespace=monitoring" | jq '.pods | length'
```

**Expected:** Number > 0 (should find running pods)

### Test 3: Node Status

```bash
curl http://localhost:8080/api/v1/metrics/nodes | jq '.nodes[] | .status' | grep -c Ready
```

**Expected:** 4 (all nodes Ready in Kind cluster)

### Test 4: Events Collection

```bash
curl "http://localhost:8080/api/v1/metrics/events?namespace=monitoring" | jq '.events | length'
```

**Expected:** Number >= 0 (may have events)

### Test 5: Combined Metrics

```bash
curl "http://localhost:8080/api/v1/metrics/all?namespace=monitoring" | jq 'keys'
```

**Expected:** ["events", "namespace", "nodes", "pods", "timestamp"]

## Manual Testing Without Port Forward

You can also test from inside the cluster:

```bash
# Port forward to local machine
kubectl port-forward svc/k8s-monitor 8080:8080 -n monitoring &

# Or exec into a pod in the cluster
kubectl exec -it deployment/nginx -n monitoring -- sh

# Then test from inside:
curl http://k8s-monitor.monitoring.svc.cluster.local:8080/health
```

## Cleanup After Testing

```bash
# Kill port forward
kill %1

# Or if in background, find and kill
ps aux | grep "port-forward"
kill <PID>
```

## Next Steps

Once testing is complete:
1. Fix any issues found
2. Add logging for debugging
3. Create unit tests
4. Add integration tests
5. Performance optimize
6. Commit to GitHub

