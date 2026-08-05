# Vaipex Golden Path

An open reference implementation of a Kubernetes golden path for delivering
secure, observable, and operable services, demonstrated with Go.

Developed by **Vaipex Labs** for the developer and platform engineering
community.

[![Validate](https://github.com/vaipexlabs/platform-lab-01-golden-path/actions/workflows/validate.yaml/badge.svg)](https://github.com/vaipexlabs/platform-lab-01-golden-path/actions/workflows/validate.yaml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Kustomize-326CE5?logo=kubernetes&logoColor=white)
![Observability](https://img.shields.io/badge/Observability-Prometheus%20%2B%20Grafana-E6522C)

## Project at a Glance

| Area | What the project delivers |
| --- | --- |
| Developer experience | A working Go service that can run locally in approximately two minutes |
| Secure runtime | A minimal non-root container and restricted Kubernetes workload |
| Deployment | Docker, Kubernetes, Kustomize, and a local [kind](https://kind.sigs.k8s.io/) environment |
| Observability | Health probes, structured logs, Prometheus metrics, and a Grafana dashboard |
| Operations | Availability, error-rate, and latency alerts linked to practical runbooks |
| Quality controls | Automated source, container, Kubernetes, Helm, and PromQL validation |
| Platform product | Documented contracts, customization points, governance, and adoption guidance |

## Explore the Project

| I want to… | Start here |
| --- | --- |
| See the service working immediately | [Try It in 2 Minutes](#try-it-in-2-minutes) |
| Understand the technical design | [Reference Architecture](#reference-architecture) |
| Understand the delivery journey | [How the Golden Path Works](#how-the-golden-path-works) |
| Develop and test locally | [Local Development](#local-development) |
| Build the secure image | [Production Container](#production-container) |
| Deploy it to Kubernetes | [Kubernetes Deployment](#kubernetes-deployment) |
| Install Prometheus and Grafana | [Observability Stack](#observability-stack) |
| Inspect metrics, logs, dashboards, and alerts | [Operations](#operations) |
| Understand automated controls | [Continuous-Integration Guardrails](#continuous-integration-guardrails) |
| Understand the repository | [Repository Structure](#repository-structure) |
| Customize or adopt the pattern | [Customization and Adoption](#customization-and-adoption) |
| Understand ownership | [Responsibility Model](#responsibility-model) |

## Try It in 2 Minutes

With Git and Go 1.26 or newer installed, clone and start the service:

~~~bash
git clone https://github.com/vaipexlabs/platform-lab-01-golden-path.git
cd platform-lab-01-golden-path
go run ./cmd/golden-path-service
~~~

The first run may download Go dependencies. Leave the service running and use a
second terminal to verify it:

~~~bash
curl http://localhost:8080/
curl http://localhost:8080/health/ready
curl --silent http://localhost:8080/metrics \
  | grep '^golden_path_http_requests_total'
~~~

Expected application responses:

~~~json
{"service":"golden-path-service","status":"running"}
{"status":"ready"}
~~~

You now have a tested service exposing health, request, runtime, and process
telemetry. Press <code>Ctrl+C</code> in the service terminal when finished.

## What Is a Golden Path?

A golden path is a supported approach for delivering a common class of
software. It combines sensible defaults, reusable components, automated
guardrails, documentation, and operational feedback so teams do not need to
recreate the same foundation for every service.

This project demonstrates how a platform team can provide that experience as a
product without hiding application ownership or preventing supported
customization.

## Reference Architecture

![Vaipex Golden Path reference architecture](docs/images/vaipex-golden-path-reference-architecture.png)

The implementation separates four concerns:

- **Developer and source experience:** local tools, GitHub, and automated
  validation.
- **Application runtime:** Kubernetes Service, Deployment, secure pods, health
  endpoints, metrics, and JSON logs.
- **Monitoring control plane:** Prometheus Operator, Prometheus, ServiceMonitor,
  PrometheusRule, and Grafana.
- **Operational experience:** the service API, Prometheus UI, Grafana dashboard,
  alerts, and runbooks.

Solid arrows represent runtime and data flows. Dashed arrows represent
configuration and discovery relationships.

## How the Golden Path Works

![Vaipex Golden Path delivery flow](docs/images/vaipex-golden-path-delivery-flow.png)

The intended delivery journey is:

1. A developer works through a documented service interface.
2. The golden path supplies supported defaults and reusable components.
3. Automated controls build, test, and validate the change.
4. A production container is created for promotion through a registry.
5. Kubernetes runs the service with secure and observable defaults.
6. Health, logs, metrics, dashboards, and alerts provide operational feedback.

The repository implements source validation and container building but
intentionally does not publish an image or deploy from CI. Registry promotion
and deployment authorization remain organization-specific controls.

## Who This Is For

| Audience | Useful starting points |
| --- | --- |
| Application developers | [Two-minute demo](#try-it-in-2-minutes), [local development](#local-development), and [container build](#production-container) |
| Platform engineers | [Architecture](#reference-architecture), [Kubernetes](#kubernetes-deployment), [observability](#observability-stack), and [CI](#continuous-integration-guardrails) |
| SRE and operations teams | [Operations](#operations), [alerts](#prometheus-alerts), and [runbooks](docs/runbooks/golden-path-service-alerts.md) |
| Engineering leaders | [Responsibility model](#responsibility-model), [adoption guidance](#customization-and-adoption), and [scope](#scope-and-non-goals) |

## Local Development

### Prerequisites

- Go 1.26 or newer

### Run the Service

~~~bash
go run ./cmd/golden-path-service
~~~

The service listens on <code>http://localhost:8080</code>.

### Verify the Endpoints

~~~bash
curl http://localhost:8080/
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
curl http://localhost:8080/metrics
~~~

| Endpoint | Purpose |
| --- | --- |
| <code>GET /</code> | Confirms the application is serving requests |
| <code>GET /health/live</code> | Indicates whether the process is alive |
| <code>GET /health/ready</code> | Indicates whether the service is ready for traffic |
| <code>GET /metrics</code> | Exposes Prometheus-compatible telemetry |

Request metrics use bounded route patterns rather than raw URLs to prevent
high-cardinality labels.

### Inspect Structured Logs

The service writes JSON logs to standard output. Request records include the
HTTP method, bounded route, response status, and duration. Successful probes
and metrics scrapes are suppressed to reduce noise; failed operational requests
are still logged.

### Test and Validate

~~~bash
go test ./...
./scripts/validate.sh
~~~

The validation script checks formatting, runs static analysis, and executes the
complete test suite.

## Production Container

### Prerequisites

- Docker

### Build

~~~bash
docker build --tag golden-path-service:local .
~~~

### Run Securely

~~~bash
docker run --rm \
  --read-only \
  --cap-drop=ALL \
  --security-opt=no-new-privileges \
  --publish 127.0.0.1:8080:8080 \
  golden-path-service:local
~~~

The multi-stage build produces a pinned, minimal distroless image. The service
runs as a non-root user with a read-only filesystem, no Linux capabilities, and
no path for acquiring additional privileges.

Use the [endpoint verification commands](#verify-the-endpoints) to confirm the
container is working.

## Kubernetes Deployment

### Prerequisites

- Docker
- [kind](https://kind.sigs.k8s.io/)
- <code>kubectl</code>

### Create a Local Cluster

Skip this command if the <code>kind</code> cluster already exists:

~~~bash
kind create cluster --name kind
~~~

### Build and Load the Image

~~~bash
docker build --tag golden-path-service:local .
kind load docker-image golden-path-service:local --name kind
~~~

### Deploy

Use Kustomize through <code>kubectl</code>; do not apply the
<code>kustomization.yaml</code> file directly:

~~~bash
kubectl apply -k deploy/kubernetes/overlays/local
kubectl rollout status \
  deployment/golden-path-service \
  --namespace golden-path
~~~

### Inspect and Access

~~~bash
kubectl get pods,service --namespace golden-path
kubectl port-forward \
  --namespace golden-path \
  service/golden-path-service \
  8081:80
~~~

The service is available at <code>http://localhost:8081</code> while the
port-forward is running.

### Remove the Local Environment

Remove only the deployed application:

~~~bash
kubectl delete -k deploy/kubernetes/overlays/local
~~~

Or delete the complete local cluster:

~~~bash
kind delete cluster --name kind
~~~

## Observability Stack

### Prerequisites

- A running Kubernetes deployment from the previous section
- Helm 4

### Install Prometheus and Grafana

~~~bash
helm repo add \
  prometheus-community \
  https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community

kubectl apply -f deploy/monitoring/namespace.yaml
kubectl apply -k deploy/kubernetes/overlays/local

helm upgrade --install monitoring \
  prometheus-community/kube-prometheus-stack \
  --version 88.1.5 \
  --namespace monitoring \
  --values deploy/monitoring/kube-prometheus-stack-values.yaml \
  --wait \
  --timeout 5m
~~~

The pinned stack installs Prometheus, Prometheus Operator, Grafana,
kube-state-metrics, and node-exporter with local resource limits.

### Install Service Observability

~~~bash
kubectl apply -f deploy/monitoring/service-monitor.yaml
kubectl apply -f deploy/monitoring/grafana-dashboard.yaml
kubectl apply -f deploy/monitoring/prometheus-rules.yaml
~~~

Prometheus discovers monitors and rules only from namespaces explicitly labeled
<code>monitoring.vaipex.io/enabled=true</code>.

### Verify the Stack

~~~bash
kubectl get pods --namespace monitoring
kubectl get servicemonitor,prometheusrule --all-namespaces
kubectl get configmap \
  golden-path-service-dashboard \
  --namespace monitoring
~~~

## Operations

### Prometheus Metrics

Start a port-forward:

~~~bash
kubectl port-forward \
  --namespace monitoring \
  service/monitoring-kube-prometheus-prometheus \
  9090:9090
~~~

Open <code>http://localhost:9090</code> and try:

~~~promql
up{namespace="golden-path"}
golden_path_http_requests_total
golden_path_http_request_duration_seconds_bucket
golden_path_http_requests_in_flight
~~~

### Grafana Dashboard

Retrieve the generated administrator password:

~~~bash
kubectl get secret monitoring-grafana \
  --namespace monitoring \
  --output jsonpath='{.data.admin-password}' \
  | base64 --decode
~~~

Start a new port-forward whenever the Grafana pod is replaced:

~~~bash
kubectl port-forward \
  --namespace monitoring \
  service/monitoring-grafana \
  3000:80
~~~

Open <code>http://localhost:3000</code>, sign in as <code>admin</code>, and
select **Golden Path Service**. The dashboard presents:

- Service availability
- Request rate by route
- HTTP 5xx error rate
- Requests in flight
- p95 request latency by route
- Request rate by status code

### Prometheus Alerts

| Alert | Condition | Severity |
| --- | --- | --- |
| <code>GoldenPathServiceTargetDown</code> | One or more targets are unavailable for two minutes | Critical |
| <code>GoldenPathServiceHighErrorRate</code> | User-facing HTTP 5xx responses exceed 5% for five minutes under meaningful traffic | Warning |
| <code>GoldenPathServiceHighLatency</code> | User-facing p95 latency exceeds 500 ms for five minutes under meaningful traffic | Warning |

Prometheus evaluates the rules locally. Alertmanager notification routing is
intentionally outside this reference implementation.

Use the
[Golden Path service alert runbook](docs/runbooks/golden-path-service-alerts.md)
for impact, diagnosis, recovery, verification, and escalation guidance.

### Application Logs

~~~bash
kubectl logs \
  --namespace golden-path \
  deployment/golden-path-service \
  --all-pods \
  --tail=100
~~~

## Continuous-Integration Guardrails

The [Validate workflow](.github/workflows/validate.yaml) runs for pull requests
and pushes to <code>main</code>.

| Job | Controls |
| --- | --- |
| Go quality | Formatting, static analysis, and unit tests |
| Container build | Complete production Dockerfile build |
| Platform configuration | Kustomize rendering, PromQL validation, Helm linting, and monitoring rendering |

The workflow uses:

- Read-only repository permissions
- Immutable GitHub Action references
- Explicit tool and image versions
- Job timeouts and superseded-run cancellation
- No publishing credentials or deployment permissions

Run the primary source checks locally:

~~~bash
./scripts/validate.sh
~~~

## Repository Structure

~~~text
.
├── .github/workflows/       GitHub Actions guardrails
├── cmd/                     Service entry point
├── internal/httpapi/        HTTP routes, metrics, and logging
├── deploy/
│   ├── kubernetes/
│   │   ├── base/            Reusable Kubernetes resources
│   │   └── overlays/local/  Local kind customization
│   └── monitoring/          Prometheus, Grafana, and alert resources
├── docs/
│   ├── images/              Delivery and architecture illustrations
│   ├── runbooks/            Operational response procedures
│   └── adoption-guide.md    Customization and governance guidance
├── scripts/validate.sh      Local source-quality checks
├── Dockerfile               Secure multi-stage production image
├── go.mod
└── go.sum
~~~

## Customization and Adoption

The [Golden Path adoption guide](docs/adoption-guide.md) defines:

- The stable runtime, Kubernetes, observability, and delivery contracts
- Supported customization points and their coupled dependencies
- Platform governance, review, versioning, and exception handling
- A pilot-based adoption approach and readiness checklist
- Capabilities organizations commonly add before production use

Prefer environment overlays over copies of the reusable Kubernetes base.
Coordinate any change to operational endpoints, labels, metric names, or metric
dimensions across their consumers.

## Responsibility Model

| Platform team owns | Application team owns |
| --- | --- |
| Supported workflows and interfaces | Business functionality |
| Build and deployment standards | Application and domain tests |
| Secure defaults and automated guardrails | Workload resource requirements |
| Shared observability capabilities | Service-level objectives |
| Platform documentation and versioning | Domain-specific telemetry and runbooks |
| Platform reliability and developer experience | Production operation of the service |

## Project Status

The planned reference implementation is complete and continuously validated.

| Capability | Status |
| --- | --- |
| Service foundation, health endpoints, tests, metrics, and structured logs | Available |
| Secure production container | Available |
| Reusable Kubernetes base and local overlay | Available |
| Prometheus and Grafana monitoring foundation | Available |
| Service dashboard, alerts, and runbooks | Available |
| Continuous-integration guardrails | Available |
| Reference architecture and delivery flow | Documented |
| Customization, governance, and adoption | Documented |

Future work can extend this foundation without changing its core purpose or
developer-facing contract.

## Scope and Non-Goals

This project is not intended to:

- Replace a complete enterprise internal developer platform
- Prescribe one technology stack for every workload
- Eliminate application-team operational ownership
- Hide operational responsibilities behind automation
- Demonstrate production-scale multi-cluster operations
- Publish artifacts or deploy to production environments

Organizations commonly add artifact signing, vulnerability scanning, workload
identity, external secrets, policy enforcement, environment promotion,
deployment approvals, centralized logs and traces, Alertmanager routing,
autoscaling, disruption controls, and recovery procedures.

## Contributing

Community feedback and contributions are welcome. Before opening a pull request:

~~~bash
./scripts/validate.sh
~~~

Describe the problem being solved, contract or operational impact, validation
evidence, and documentation changes. Platform changes should preserve secure
and observable defaults or document a narrowly scoped exception.

## About Vaipex Labs

**Vaipex Labs** develops open reference implementations, engineering patterns,
and practical tools that help the developer community build reliable, secure,
and scalable software platforms.
