# Velure

Velure is a cloud-native e-commerce capstone project built to demonstrate modern software and platform engineering practices. It combines Go microservices, a React/TypeScript frontend, asynchronous order processing, containerized local development, Kubernetes deployment assets, Terraform-based AWS EKS infrastructure, and an observability stack.

This is an educational TCC/capstone project, not a production-ready commerce system. The repository is structured to show practical architecture, deployment, CI/CD, testing, and operations patterns without overstating operational maturity.

## Architecture Overview

Velure uses an event-driven microservices architecture. The UI sends user and order requests through a Caddy reverse proxy. Backend services own their data boundaries and coordinate the order lifecycle through RabbitMQ:

1. The UI creates an order through `publish-order-service`.
2. `publish-order-service` persists the order in PostgreSQL and publishes an `order.created` event.
3. `process-order-service` consumes the event, checks inventory through `product-service`, simulates payment logic, and publishes the final status.
4. `publish-order-service` stores the status update and streams it back to the UI with Server-Sent Events.

Local development runs with Docker Compose. Cloud deployment assets target Kubernetes on AWS EKS, with Helm charts for services and Terraform modules for AWS infrastructure.

## Service Map

| Service | Path | Responsibility | Main technologies |
| --- | --- | --- | --- |
| UI service | [`services/ui-service`](./services/ui-service) | Customer-facing web application | React, TypeScript, Vite, Tailwind CSS |
| Auth service | [`services/auth-service`](./services/auth-service) | Registration, login, token validation, user/session APIs | Go, PostgreSQL, Redis |
| Product service | [`services/product-service`](./services/product-service) | Product catalog and inventory lookups | Go, MongoDB, Redis |
| Publish order service | [`services/publish-order-service`](./services/publish-order-service) | Order intake, persistence, event publishing, status streaming | Go, PostgreSQL, RabbitMQ, SSE |
| Process order service | [`services/process-order-service`](./services/process-order-service) | Asynchronous order processing, inventory checks, simulated payment outcome | Go, RabbitMQ |

## Tech Stack

- Backend: Go services using HTTP APIs and service-specific packages.
- Frontend: React, TypeScript, Vite, shadcn-ui, Tailwind CSS.
- Messaging: RabbitMQ for asynchronous order events.
- Data stores: PostgreSQL, MongoDB, and Redis.
- Local runtime: Docker Compose with Caddy as the reverse proxy.
- Kubernetes: Helm charts under [`infrastructure/kubernetes/charts`](./infrastructure/kubernetes/charts).
- Cloud infrastructure: Terraform modules for AWS EKS, VPC, RDS PostgreSQL, Amazon MQ RabbitMQ, Secrets Manager, and related resources under [`infrastructure/terraform`](./infrastructure/terraform).
- Observability: Prometheus, Grafana, Loki, Promtail, AlertManager, exporters, dashboards, and ServiceMonitors.
- CI/CD: GitHub Actions workflows for Go and Node test coverage, SonarCloud analysis, Docker image builds, and Kubernetes deployment workflows.

## Local Quick Start

Prerequisites:

- Docker and Docker Compose
- Make
- Go 1.25+ for backend development
- Node.js/npm or Bun for UI development

Add the local host entry before opening the app:

```bash
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

Start the full local environment:

```bash
make local-up
```

Open the application at:

```text
https://velure.local
```

Useful local development URLs:

- Grafana: `http://localhost:3000` (`admin` / `admin`)
- Prometheus: `http://localhost:9090`
- RabbitMQ management: `http://localhost:15672` (`admin` / `admin_password`)
- cAdvisor: `http://localhost:8080`

These credentials are for the local Docker Compose environment only.

Stop and clean the local environment:

```bash
make local-down
```

## Documentation

- Project overview: [`docs/01-overview.md`](./docs/01-overview.md)
- Local quickstart: [`docs/02-quickstart.md`](./docs/02-quickstart.md)
- Core architecture and event flow: [`docs/03-core-architecture.md`](./docs/03-core-architecture.md)
- Microservice documentation: [`docs/microservices`](./docs/microservices)
- Kubernetes monitoring stack: [`infrastructure/kubernetes/monitoring/README.md`](./infrastructure/kubernetes/monitoring/README.md)
- Terraform AWS infrastructure: [`infrastructure/terraform/README.md`](./infrastructure/terraform/README.md)

## Testing and Quality

The repository includes unit tests across the Go services and frontend coverage support for the UI. CI workflows run service-level tests with coverage and feed results into SonarCloud analysis before building Docker images.

Common local checks:

```bash
# From an individual Go service directory
go test ./...

# From services/ui-service
bun run test:coverage -- --run
```

Additional quality assets include SonarCloud project configuration, k6 load-test scripts for selected services, Prometheus metrics endpoints, Grafana dashboards, and Kubernetes ServiceMonitor definitions.

## Repository Structure

```text
.
|-- docs/                    # Source Markdown documentation
|-- infrastructure/
|   |-- kubernetes/          # Helm charts, manifests, monitoring assets
|   |-- local/               # Docker Compose local environment
|   `-- terraform/           # AWS infrastructure as code
|-- scripts/                 # Deployment automation
|-- services/                # Go microservices and React UI
`-- .github/workflows/       # CI/CD workflows
```

## Project Status

Velure is a portfolio-grade educational project for demonstrating cloud-native application design, microservice boundaries, asynchronous processing, infrastructure as code, observability, and CI/CD. It is suitable for learning, review, and recruiting conversations, but it should be hardened further before any real production use.

## License

This project is licensed under the MIT License. See [`LICENSE`](./LICENSE).
