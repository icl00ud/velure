# Core Architecture & Event Flow

Velure relies on an event-driven microservices architecture to handle its core business processes, particularly the order lifecycle. This approach ensures high availability, loose coupling, and scalability across the platform without relying on a single centralized database.

## Order Lifecycle Event Flow

The most critical flow in Velure is the order creation and processing pipeline. It utilizes HTTP for synchronous initial requests and Server-Sent Events (SSE), while relying on RabbitMQ for asynchronous processing between backend services.

Below is the sequence diagram illustrating the complete order lifecycle:

```mermaid
sequenceDiagram
    participant Frontend as UI Service (Frontend)
    participant PublishService as Publish Order Service
    participant DB as PostgreSQL (Publish DB)
    participant RabbitMQ as RabbitMQ (Exchange: orders)
    participant ProcessService as Process Order Service
    participant ProductService as Product Service

    Frontend->>PublishService: POST /api/order/create-order
    PublishService->>DB: Save order (Status: CREATED)
    PublishService->>RabbitMQ: Publish event `order.created`
    PublishService-->>Frontend: Return Order ID

    Frontend->>PublishService: GET /api/order/user/order/status?id=X
    PublishService-->>Frontend: Open SSE Stream

    RabbitMQ->>ProcessService: Consume `order.created` queue
    ProcessService->>ProductService: HTTP GET Inventory Check
    ProductService-->>ProcessService: Inventory Status
    
    alt Inventory Available & Payment Success
        ProcessService->>RabbitMQ: Publish `order.completed`
    else Inventory Unavailable or Payment Failed
        ProcessService->>RabbitMQ: Publish `order.failed`
    end

    RabbitMQ->>PublishService: Consume status update
    PublishService->>DB: Update order status (COMPLETED / FAILED)
    PublishService->>Frontend: Broadcast via SSE (New Status)
```

## Distributed State Management

In this architecture, Velure avoids a monolithic centralized database. Instead, state is managed across services using an event-driven approach.

The order transitions through the following states:
1. **CREATED**: The initial state when the `publish-order-service` receives the request and saves it to its local PostgreSQL database.
2. **PROCESSING**: The state when the `process-order-service` picks up the event from RabbitMQ and begins validating inventory via the `product-service` and handling simulated payment logic.
3. **COMPLETED / FAILED**: The terminal states. Once the `process-order-service` finishes its operations, it publishes a final event back to RabbitMQ. The `publish-order-service` consumes this, updates the local database, and pushes the final state to the frontend via SSE.

This decoupled design ensures that if the processing or product services are temporarily unavailable, orders are not lost—they remain safely queued in RabbitMQ until they can be processed.

## High-Level Architecture

The diagram below shows the runtime topology: client traffic enters through Caddy, services own private data stores, RabbitMQ brokers async work, and every consumer queue is paired with a Dead Letter Queue (DLQ) so poison messages are quarantined instead of looping forever.

```mermaid
flowchart LR
    Browser([Browser / SPA])

    subgraph Edge
        Caddy[Caddy<br/>Reverse Proxy]
    end

    subgraph Services
        UI[ui-service<br/>React + Vite]
        Auth[auth-service<br/>Go + Gin]
        Product[product-service<br/>Go + Fiber]
        Publish[publish-order-service<br/>Go + net/http<br/>+ SSE]
        Process[process-order-service<br/>Go + net/http]
    end

    subgraph Data
        AuthDB[(PostgreSQL<br/>auth)]
        Redis[(Redis<br/>token cache)]
        Mongo[(MongoDB<br/>products)]
        OrderDB[(PostgreSQL<br/>orders)]
    end

    subgraph Broker[RabbitMQ]
        Ex{{exchange<br/>orders topic}}
        QProc[[process-order-queue]]
        QPub[[publish-order-status-updates]]
        DlxA{{orders.dlx fanout}}
        DlxB{{publish.dlx fanout}}
        DlqA[[process-order-queue.dlq]]
        DlqB[[publish-order-status-updates.dlq]]
    end

    Browser <--> Caddy
    Caddy --> UI
    Caddy --> Auth
    Caddy --> Product
    Caddy --> Publish

    Auth --- AuthDB
    Auth --- Redis
    Product --- Mongo
    Publish --- OrderDB

    Publish -- "order.created" --> Ex
    Process -- "order.processing<br/>order.completed<br/>order.failed" --> Ex

    Ex -- "order.created" --> QProc
    Ex -- "order.processing/completed/failed" --> QPub

    QProc --> Process
    QPub --> Publish
    Process -- "HTTP inventory check" --> Product

    QProc -. "nack no-requeue<br/>(poison / max retries)" .-> DlxA --> DlqA
    QPub  -. "nack no-requeue<br/>(poison / max retries)" .-> DlxB --> DlqB

    Publish == "SSE status stream" ==> Browser
```

### Reading the diagram

- **Synchronous path:** Browser → Caddy → service. Caddy is the single ingress; never hit container ports directly.
- **Async path:** `publish-order-service` writes the `order.created` event to the `orders` topic exchange. `process-order-service` consumes it from `process-order-queue`, calls `product-service` for inventory, then republishes a status event. `publish-order-service` consumes that status from `publish-order-status-updates` and pushes it to the browser via SSE.
- **DLQ pattern:** Each consumer queue has `x-dead-letter-exchange` set. Permanent errors, parse failures and messages exceeding `maxRetries` are `Nack(false, false)` → routed to the per-stream DLX (`orders.dlx`, `publish.dlx`) → land in the matching DLQ for inspection/replay instead of looping back into the main queue.
- **State isolation:** Each service owns its own data store; cross-service queries happen over HTTP (`process → product`), never via shared schemas.
