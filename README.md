# Vaipex Golden Path

An open-source reference implementation for building secure, observable, and
production-ready services on Kubernetes.

Developed by **Vaipex Labs** for the developer and platform engineering
community.

## Overview

Engineering teams frequently recreate the same foundational capabilities for
every service: build automation, container packaging, deployment
configuration, health checks, observability, security controls, and continuous
integration.

Vaipex Golden Path is intended to provide an opinionated, reusable service
foundation that helps teams adopt consistent engineering standards without
requiring every developer to become an expert in the underlying platform
tooling.

The project will demonstrate how platform teams can deliver reusable
capabilities as a product, with a clear developer interface, sensible defaults,
automated guardrails, and documented extension points.

## Who This Is For

- Platform engineering teams designing internal developer platforms
- Application teams deploying services to Kubernetes
- SRE and DevOps teams standardizing production readiness
- Engineering leaders evaluating golden paths and paved-road strategies
- Developers seeking practical cloud-native reference implementations

## Project Goals

- Reduce the effort required to create a production-ready service
- Provide consistent build, test, package, and deployment workflows
- Establish secure and observable defaults
- Minimize the cognitive load placed on application developers
- Provide reusable patterns without creating a restrictive golden cage
- Demonstrate platform-as-a-product principles through working software
- Encourage community discussion around practical platform engineering

## Platform Principles

### Product-Oriented

Platform capabilities should be designed around developer needs, supported
interfaces, documentation, feedback, and measurable outcomes.

### Secure by Default

The supported path should provide secure defaults so application teams do not
need to discover every control independently.

### Observable by Default

Health, readiness, metrics, and structured logs should be foundational service
capabilities rather than optional additions.

### Opinionated but Extensible

The common path should be intentionally standardized. Supported extension
points should address legitimate workload differences without requiring teams
to duplicate the platform.

### Automated and Repeatable

Local and continuous-integration workflows should apply the same validations
to reduce environmental inconsistencies.

## Responsibility Model

### Platform Team

The platform team owns:

- Supported workflows and interfaces
- Build and deployment standards
- Secure defaults and automated guardrails
- Shared observability capabilities
- Documentation and versioning
- Platform reliability and developer experience

### Application Team

Application teams own:

- Business functionality
- Application tests
- Service-specific resource requirements
- Service-level objectives
- Domain-specific alerts and runbooks
- Operational ownership

## Delivery Roadmap

- [x] Define and document the golden path delivery flow.
- [ ] Define and document the reference architecture.
- [x] Establish the Go service foundation and HTTP API.
- [x] Add liveness and readiness endpoints.
- [x] Add automated endpoint tests.
- [x] Document the supported local developer workflow.
- [x] Add automated code-quality validation.
- [x] Create a secure production container.
- [x] Add a reusable Kubernetes deployment model.
- [x] Expose Prometheus-compatible runtime and process metrics.
- [x] Add low-cardinality application request metrics.
- [x] Add structured request logs.
- [x] Add Prometheus monitoring foundation and service discovery.
- [x] Add a Grafana service dashboard and verify it in the local cluster.
- [x] Add operational alerts and runbooks.
- [ ] Add continuous-integration guardrails.
- [ ] Document customization, governance, and adoption guidance.

Each capability will be delivered as a small, independently verifiable change.

## Delivery

### Golden Path Delivery Flow

![Vaipex Golden Path delivery flow](docs/images/vaipex-golden-path-delivery-flow.png)

> A developer submits code, the platform verifies and packages it, Kubernetes runs it, and operational feedback shows whether it is working properly.

### Local Development

#### Prerequisites

- Go 1.26 or newer

#### Run the Service

```bash
go run ./cmd/golden-path-service
```

The service listens on `http://localhost:8080`.

#### Verify the Endpoints

```bash
curl http://localhost:8080/
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

#### Inspect Metrics

```bash
curl http://localhost:8080/metrics
```

The metrics endpoint exposes Prometheus-compatible Go runtime, process, and
HTTP request measurements. Request metrics use bounded route patterns rather
than raw URLs to avoid high-cardinality labels.

#### Inspect Structured Logs

The service writes JSON logs to standard output. Request events include the
HTTP method, bounded route pattern, response status, and duration. Successful
health probes and metrics scrapes are omitted to reduce operational noise;
failed operational requests are still logged.

#### Run the Tests

```bash
go test ./...
```

#### Validate Changes

```bash
./scripts/validate.sh
```

The validation command checks Go formatting, runs static analysis, and executes
the complete test suite.

Press `Ctrl+C` in the service terminal to stop the application.

### Container

#### Build the Image

```bash
docker build --tag golden-path-service:local .
```

#### Run the Container

```bash
docker run --rm \
  --read-only \
  --cap-drop=ALL \
  --security-opt=no-new-privileges \
  --publish 127.0.0.1:8080:8080 \
  golden-path-service:local
```

The runtime image is pinned, minimal, and configured to run as a non-root user.
The runtime command also uses a read-only filesystem, removes Linux
capabilities, and prevents the process from gaining additional privileges.

Use the endpoint verification commands from the local development workflow to
confirm the container is running correctly.

### Kubernetes

The reusable Kubernetes base is adapted for a local
[kind](https://kind.sigs.k8s.io/) cluster through a Kustomize overlay.

#### Load the Local Image

```bash
kind load docker-image golden-path-service:local --name kind
```

#### Deploy the Service

```bash
kubectl apply -k deploy/kubernetes/overlays/local
kubectl rollout status deployment/golden-path-service --namespace golden-path
```

#### Inspect the Workload

```bash
kubectl get pods,service --namespace golden-path
```

#### Access the Service

```bash
kubectl port-forward --namespace golden-path service/golden-path-service 8081:80
```

In another terminal, use the endpoint verification commands with
`http://localhost:8081`.

### Monitoring Foundation

The local monitoring foundation uses the pinned `kube-prometheus-stack` Helm
chart to run Prometheus, the Prometheus Operator, Grafana, kube-state-metrics,
and node-exporter.

#### Add the Chart Repository

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
```

#### Install the Monitoring Stack

```bash
kubectl apply -f deploy/monitoring/namespace.yaml
kubectl apply -k deploy/kubernetes/overlays/local

helm upgrade --install monitoring \
  prometheus-community/kube-prometheus-stack \
  --version 88.1.5 \
  --namespace monitoring \
  --values deploy/monitoring/kube-prometheus-stack-values.yaml \
  --wait \
  --timeout 5m
```

#### Enable Service Discovery

```bash
kubectl apply -f deploy/monitoring/service-monitor.yaml
```

Prometheus selects ServiceMonitors only from namespaces labeled
`monitoring.vaipex.io/enabled=true`. The application ServiceMonitor discovers
both service pods and scrapes `/metrics` every 15 seconds.

#### Inspect the Monitoring Workloads

```bash
kubectl get pods --namespace monitoring
kubectl get servicemonitor --all-namespaces
```

#### Access Prometheus

```bash
kubectl port-forward \
  --namespace monitoring \
  service/monitoring-kube-prometheus-prometheus \
  9090:9090
```

Open `http://localhost:9090` and query:

```promql
up{namespace="golden-path"}
golden_path_http_requests_total
```

### Grafana Service Dashboard

The Grafana sidecar discovers dashboard ConfigMaps labeled
`grafana_dashboard=1`. The version-controlled service dashboard provides views
of availability, request rate, HTTP 5xx error rate, in-flight requests, p95
latency, and response status codes.

#### Install the Dashboard

After installing or upgrading the monitoring stack with the repository values,
apply the dashboard ConfigMap:

```bash
kubectl apply -f deploy/monitoring/grafana-dashboard.yaml
```

#### Access Grafana

```bash
kubectl port-forward \
  --namespace monitoring \
  service/monitoring-grafana \
  3000:80
```

Retrieve the generated administrator password:

```bash
kubectl get secret \
  --namespace monitoring \
  monitoring-grafana \
  --output jsonpath='{.data.admin-password}' | base64 --decode
```

Open `http://localhost:3000`, sign in as `admin`, and select the
**Golden Path Service** dashboard.

### Operational Alerts and Runbooks

Prometheus evaluates service availability, HTTP 5xx error rate, and p95 latency
rules from monitoring-enabled namespaces. Each alert links to the
[Golden Path service alert runbook](docs/runbooks/golden-path-service-alerts.md),
which documents impact, diagnosis, recovery, verification, and escalation.

After installing or upgrading the monitoring stack with the repository values,
apply the alert rules:

```bash
kubectl apply -f deploy/monitoring/prometheus-rules.yaml
```

Inspect the deployed rules:

```bash
kubectl get prometheusrule --namespace golden-path
```

Open Prometheus and select **Alerts** to inspect each rule's current state.
Alertmanager delivery and notification routing are intentionally deferred from
this local foundation.

## Success Criteria

The project will be successful when a developer can:

- Clone the repository and understand its purpose quickly
- Test and run the reference service locally
- Build a secure container image
- Deploy the service to a local Kubernetes cluster
- Observe application health and operational metrics
- Validate changes through automated controls
- Customize documented settings without modifying platform internals

## Non-Goals

This project is not intended to:

- Replace a complete enterprise internal developer platform
- Prescribe one technology stack for every workload
- Eliminate the need for application ownership
- Hide operational responsibilities behind automation
- Demonstrate production-scale multi-cluster operations

It provides a focused reference implementation that teams can evaluate,
extend, and adapt to their environments.

## Contributing

Community feedback and contributions are welcome. Contribution guidance,
development setup, and review expectations will be added as the implementation
evolves.

## About Vaipex Labs

**Vaipex Labs** develops open reference implementations, engineering patterns,
and practical tools that help the developer community build reliable, secure,
and scalable software platforms.
