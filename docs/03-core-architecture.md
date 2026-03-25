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
