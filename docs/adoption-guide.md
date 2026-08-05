# Golden Path Adoption Guide

This guide helps platform teams evaluate, customize, govern, and adopt the
Vaipex Golden Path while preserving its secure and observable defaults.

The repository is a reference implementation, not a universal production
platform. Adopters should keep the developer-facing contract stable, customize
the documented extension points, and add organization-specific capabilities
through reviewed overlays and policies.

## Platform Contract

The platform contract is the small set of behaviors that connects application
code, Kubernetes, observability, and continuous integration. Changes to these
behaviors must be coordinated across every dependent component.

### Runtime Interface

The reference service provides:

| Interface | Contract | Consumers |
| --- | --- | --- |
| Application listener | HTTP on port `8080` | Container, Deployment, Service |
| Liveness | `GET /health/live` | Kubernetes liveness probe |
| Readiness | `GET /health/ready` | Kubernetes readiness probe |
| Metrics | `GET /metrics` | ServiceMonitor and Prometheus |
| Logs | Structured JSON on standard output | Kubernetes and log collectors |

Applications may add routes and business metrics without changing these
operational interfaces.

### Kubernetes Interface

The Kubernetes resources depend on these conventions:

- `app.kubernetes.io/name` consistently identifies the workload and Service.
- The container and Service expose a port named `http`.
- Service selectors match pod labels.
- Health probes use the documented liveness and readiness paths.
- Monitoring-enabled namespaces use
  `monitoring.vaipex.io/enabled=true`.
- ServiceMonitor selection labels match the application Service.

The default workload security controls are part of the supported path:

- Non-root execution
- Read-only root filesystem
- No privilege escalation
- All Linux capabilities dropped
- Runtime-default seccomp profile
- Explicit CPU and memory requests and limits

An application requiring different privileges should use a reviewed exception
rather than silently weakening the shared baseline.

### Observability Interface

The dashboard and alerts depend on the `golden_path_http_*` metric family and
its bounded `method`, `route`, and `status` labels. Route values must remain
templates such as `/orders/{id}` rather than raw user-controlled URLs. This
prevents unbounded Prometheus cardinality.

Changes to metric names or labels require coordinated updates to:

- Application instrumentation
- Grafana dashboard queries
- Prometheus alert expressions
- Runbooks and operational documentation

### Delivery Interface

Every proposed change is expected to pass the repository's GitHub Actions
workflow. The supported local entry point for source validation is:

```bash
./scripts/validate.sh
```

CI validates source quality, unit tests, the production container build,
Kustomize rendering, Prometheus rules, and monitoring Helm configuration.

## Supported Customization Points

Use this map to identify the resources affected by a customization.

| Customization | Primary locations | Coupled updates | Verification |
| --- | --- | --- | --- |
| Service identity | Go response, Kubernetes names and labels | ServiceMonitor, dashboard, alerts, runbooks | Endpoint tests, Kustomize render, Prometheus queries |
| Namespace | Local overlay namespace and Kustomization | ServiceMonitor, PrometheusRule expressions, monitoring opt-in label | Kustomize render and Prometheus discovery |
| Container image | Kubernetes base and environment overlay | Registry authentication and image policy | Container build and rollout status |
| Replica count | Deployment or environment overlay | Capacity assumptions and disruption policy | Pod readiness and target discovery |
| CPU and memory | Deployment resources or environment overlay | Capacity planning and autoscaling thresholds | Resource usage and load testing |
| Alert thresholds | `deploy/monitoring/prometheus-rules.yaml` | Runbook expectations and service objectives | `promtool` and controlled test traffic |
| Dashboard panels | `deploy/monitoring/grafana-dashboard.yaml` | Metric names, labels, and datasource UID | Grafana API and visual inspection |
| Monitoring stack | `deploy/monitoring/kube-prometheus-stack-values.yaml` | Resource usage, retention, discovery, and chart compatibility | Helm lint, render, and cluster health |

### Prefer Overlays for Environment Differences

Keep reusable defaults in `deploy/kubernetes/base` and represent
environment-specific changes in overlays. Typical overlay changes include:

- Namespace
- Image repository and tag
- Replica count
- Resource sizing
- Environment-specific configuration

Avoid copying the complete base into each environment. Copies drift and make
platform upgrades harder to review.

### Keep Application Changes Inside the Contract

Application teams can add business routes, packages, tests, and domain metrics.
They should preserve the operational endpoints and low-cardinality telemetry
contract unless a coordinated platform change intentionally replaces them.

### Treat Security Changes as Exceptions

When a workload cannot operate with a default security control:

1. Document the technical reason and affected workload.
2. Identify the smallest possible exception.
3. Add a compensating control when practical.
4. Assign an owner and review date.
5. Keep the exception in an environment or workload-specific overlay.

Do not weaken the reusable base to accommodate a single exceptional workload.

## Governance Model

Governance protects the contract without turning the golden path into a golden
cage. Controls should be proportional, transparent, and supported by automated
feedback.

### Ownership

Platform maintainers own:

- Runtime and deployment contracts
- Secure and observable defaults
- CI guardrails and shared tooling
- Versioning, migration guidance, and platform documentation
- Reliability of the supported developer experience

Application owners retain responsibility for:

- Business behavior and application tests
- Workload sizing and service objectives
- Domain-specific telemetry and runbooks
- Production operation of their service

### Change Classification

Classify changes before review:

| Change type | Example | Expected treatment |
| --- | --- | --- |
| Compatible | Add a test or dashboard panel | Normal review and CI validation |
| Customization | Change replicas in an overlay | Application-owner review and environment verification |
| Contract change | Rename a health endpoint or metric | Platform review, migration plan, and coordinated release |
| Exception | Add a Linux capability for one workload | Risk review, compensating control, owner, and expiry |

### Review Expectations

A platform change should explain:

- The developer or operational problem being solved
- Whether the platform contract changes
- Security and reliability consequences
- Validation evidence
- Upgrade or rollback requirements
- Documentation and runbook impact

Automated checks provide fast feedback, but they do not replace review of
architecture, operational impact, or user experience.

### Versioning and Upgrades

Adopting teams should release the golden path as a versioned product rather
than asking services to follow an unpinned branch.

- Use semantic versions for supported releases.
- Publish release notes for contract and default changes.
- Provide migration instructions for breaking changes.
- Test upgrades against representative pilot services.
- Define a support window and deprecation policy.
- Track application adoption so unsupported versions are visible.

When consuming the reference through a fork, regularly compare upstream
changes and merge them through the same review and CI process as internal
changes.

### Exceptions and Feedback

An exception is useful product feedback. Review recurring exceptions to decide
whether the platform needs a supported extension point. Record exceptions with
an owner, rationale, scope, compensating controls, and review date.

Maintain a visible feedback channel and publish decisions so application teams
understand what is supported and why.

## Adoption Approach

Adoption should be evidence-driven and incremental.

### 1. Define the Target Outcome

Before introducing the golden path, document the problems it should improve,
such as inconsistent service setup, security gaps, slow onboarding, or weak
operational visibility. Select a small number of measurable outcomes.

Useful measures include:

- Time from repository creation to a working deployment
- Percentage of services passing platform controls
- Deployment failure and rollback rate
- Time required to diagnose a service failure
- Platform adoption, retention, and developer satisfaction

### 2. Establish Organizational Defaults

Review the contract and replace only the organization-specific values needed
for a pilot:

- Source and container registries
- Namespace and identity conventions
- Resource policies
- Observability ownership and retention
- Required CI and security controls
- Support and escalation channels

Document every intentional difference from this reference.

### 3. Run a Low-Risk Pilot

Choose one representative, low-risk service and a willing application team.
Validate the complete journey:

1. Create and understand the service.
2. Run tests and local validation.
3. Build the production container.
4. Deploy through the supported overlay.
5. Confirm health, metrics, dashboard, alerts, and logs.
6. Exercise the runbook for a controlled target failure.
7. Record friction, workarounds, and missing capabilities.

The pilot is successful when the team can operate the service through the
documented interfaces, not merely when Kubernetes reports a running pod.

### 4. Improve the Product

Prioritize feedback that removes repeated developer effort or operational
risk. Prefer improvements to defaults, automation, and documentation over
additional process. Re-run the pilot journey after material changes.

### 5. Expand in Cohorts

Adopt the golden path with small groups of services that have similar needs.
Track exceptions and support demand. Expand only when the platform team can
maintain the experience and respond to feedback.

## Adoption Readiness Checklist

A team is ready to broaden adoption when it can answer yes to the following:

- [ ] The supported platform contract is documented and versioned.
- [ ] Application and platform ownership boundaries are understood.
- [ ] Secure defaults and exception handling are defined.
- [ ] CI guardrails run consistently for proposed changes.
- [ ] Health, metrics, dashboards, alerts, logs, and runbooks are verified.
- [ ] A pilot service has completed the full delivery and operational journey.
- [ ] Success measures and a developer feedback channel exist.
- [ ] Upgrade, deprecation, and support expectations are published.

## What to Add for Production Use

Organizations commonly extend this reference with:

- Artifact signing, provenance, vulnerability scanning, and registry policy
- External secrets and workload identity
- Environment promotion and deployment approvals
- Policy enforcement and admission controls
- Ingress, DNS, certificates, and network policies
- Autoscaling, disruption budgets, and topology constraints
- Centralized logs and traces
- Alertmanager routing and notification ownership
- Backup, recovery, and multi-cluster operating procedures

These capabilities should follow the same design principle: provide a secure,
observable default with a small supported interface and clear ownership.
