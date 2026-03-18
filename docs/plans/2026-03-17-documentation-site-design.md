# Velure Documentation Site Design

## Overview
Create a professional, static documentation site using Docusaurus (React-based) to document the Velure e-commerce microservices platform. This site will replace the single `README.md` file with a structured, searchable, and easily navigable portal.

## Architecture & Structure

The documentation will be hosted in a new `docs-site/` directory (or similar, to be determined during implementation) and built using Docusaurus.

### 1. Home & Overview
*   **Landing Page:** Hero section introducing Velure as a cloud-native, event-driven e-commerce platform built with Go and React.
*   **Architecture Diagram:** Inclusion of the existing AWS infrastructure diagram (`diagrams/architecture-Velure AWS Infrastructure.drawio.png`).
*   **Tech Stack:** High-level overview of the technologies used (Go, React, Postgres, MongoDB, RabbitMQ, Docker, Caddy, Kubernetes).

### 2. Getting Started (Quickstart)
*   **Prerequisites:** Docker, Make, Node, Go.
*   **Local Setup:** Step-by-step instructions using `make local-up`.
*   **Critical Warning:** Mandatory `/etc/hosts` configuration for `velure.local` to ensure Caddy routing works correctly and prevents 405 errors.
*   **Cloud Deployment:** Instructions for AWS deployment using `make cloud-up` and teardown with `make cloud-down`.

### 3. Core Architecture: Event-Driven Order Flow
*   **Visual Flow:** A Mermaid sequence diagram detailing the order lifecycle.
*   **Process:**
    1.  Frontend `POST` -> `publish-order-service`.
    2.  Save to Postgres -> Publish `order.created` to RabbitMQ.
    3.  `process-order-service` consumes -> Calls `product-service` (HTTP) -> Processes payment/inventory.
    4.  Publish status update to RabbitMQ.
    5.  `publish-order-service` consumes -> Updates Postgres -> Broadcasts via SSE.
    6.  Frontend receives real-time SSE updates.
*   **State Management:** Explanation of `CREATED` -> `PROCESSING` -> `COMPLETED`/`FAILED` states.

### 4. Microservices Reference
A dedicated section detailing each service:
*   **Shared Module (`shared/`):** Go module `replace` directive, `logger/`, and `models/`.
*   **Auth Service:** Go + Gin, Postgres, Redis (JWT caching). Endpoints (`/register`, `/login`), environment variables, and test commands.
*   **Product Service:** Go + Fiber, MongoDB. Endpoints (`/api/product/`), environment variables, and test commands.
*   **Publish Order Service:** Go + net/http, Postgres, RabbitMQ publisher, SSE implementation.
*   **Process Order Service:** Go + net/http, RabbitMQ consumer, synchronous HTTP calls to Product Service.
*   **UI Service:** React + Vite, React Router v6, React Query v5, Radix UI/shadcn. Commands (`dev`, `build`, `lint` via Biome).

### 5. DevOps & Operations
*   **Docker:** Multi-stage builds for Go, Bun/Nginx for UI.
*   **Networking:** Local Docker networks (`local_auth`, `local_order`, `local_frontend`) and Caddy reverse proxy routing.
*   **Makefiles:** Comprehensive list of available commands.

## Implementation Steps
1.  Initialize Docusaurus project in the repository root (e.g., `npx create-docusaurus@latest docs-site classic`).
2.  Configure `docusaurus.config.js` (title, logo, GitHub links).
3.  Create the directory structure mapping to the design above within the Docusaurus `docs/` folder.
4.  Migrate content from the existing `README.md` and `CLAUDE.md` into the new structured Markdown files.
5.  Create the Mermaid diagrams for the event flow.
6.  Ensure all code blocks, environment variables, and commands are accurately formatted.
7.  Update the root `README.md` to point to the new documentation site and provide basic "how to start the docs" instructions.
