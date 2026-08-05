# Golden Path Service Alert Runbook

This runbook supports the Prometheus alerts defined for the Golden Path
service. Start with the common checks, then follow the procedure for the alert
that is firing.

## Common Checks

Confirm that the cluster and application workloads are reachable:

```bash
kubectl cluster-info
kubectl get pods --namespace golden-path --output wide
kubectl get deployment,service --namespace golden-path
kubectl get servicemonitor,prometheusrule --namespace golden-path
```

Inspect recent application events and logs:

```bash
kubectl get events --namespace golden-path --sort-by=.lastTimestamp
kubectl logs \
  --namespace golden-path \
  deployment/golden-path-service \
  --all-pods \
  --tail=100
```

## GoldenPathServiceTargetDown

### Meaning and Impact

Prometheus cannot scrape one or more service targets. A single unavailable
replica reduces redundancy; all unavailable replicas make the service
unavailable.

### Diagnose

```bash
kubectl get pods --namespace golden-path --output wide
kubectl describe deployment golden-path-service --namespace golden-path
kubectl get endpointslice --namespace golden-path \
  --selector kubernetes.io/service-name=golden-path-service
```

Check whether readiness and metrics are available from inside the cluster:

```bash
kubectl run monitoring-check \
  --namespace golden-path \
  --restart=Never \
  --rm \
  --stdin \
  --tty \
  --image=curlimages/curl:8.17.0 \
  --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"monitoring-check","image":"curlimages/curl:8.17.0","args":["http://golden-path-service/metrics"],"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}]}}'
```

### Recover

- Correct failed image pulls, scheduling failures, or invalid configuration.
- Restore healthy application replicas before changing monitoring resources.
- If the application is healthy, verify the Service selector, ServiceMonitor,
  namespace opt-in label, and network connectivity.

### Verify

```promql
up{namespace="golden-path", service="golden-path-service"}
```

Every returned target should have the value `1`.

## GoldenPathServiceHighErrorRate

### Meaning and Impact

More than 5% of requests are returning HTTP 5xx responses while the service is
receiving meaningful traffic. Users may see failed or incomplete operations.

### Diagnose

Identify affected status codes and routes in Grafana, then inspect application
logs:

```bash
kubectl logs \
  --namespace golden-path \
  deployment/golden-path-service \
  --all-pods \
  --since=15m
```

Use Prometheus to break down the error rate:

```promql
sum by (route, status) (
  rate(golden_path_http_requests_total{status=~"5.."}[5m])
)
```

### Recover

- Identify whether failures began after a deployment or configuration change.
- Correct the failing dependency, request-handling path, or configuration.
- Roll back only when a recent application change is confirmed as the cause.

### Verify

Confirm that the five-minute HTTP 5xx ratio returns below `0.05` and remains
stable for at least five minutes.

## GoldenPathServiceHighLatency

### Meaning and Impact

The p95 request duration is above 500 milliseconds while the service is
receiving meaningful traffic. At least 5% of requests may be taking longer
than the expected response-time threshold.

### Diagnose

Compare latency by route:

```promql
histogram_quantile(
  0.95,
  sum by (le, route) (
    rate(golden_path_http_request_duration_seconds_bucket[5m])
  )
)
```

Check resource usage and recent application logs:

```bash
kubectl top pods --namespace golden-path
kubectl logs \
  --namespace golden-path \
  deployment/golden-path-service \
  --all-pods \
  --since=15m
```

### Recover

- Identify the slow route and correlate it with resource pressure or errors.
- Correct slow application logic, constrained resources, or dependencies.
- Scale replicas only when evidence shows that concurrency or saturation is
  the cause.

### Verify

Confirm that p95 latency returns below `0.5` seconds and remains stable for at
least five minutes.

## Escalation Information

When escalating an unresolved alert, provide:

- Alert name and firing duration
- Affected pods and routes
- Deployment revision and recent changes
- Relevant logs, events, and Prometheus query results
- Remediation already attempted
