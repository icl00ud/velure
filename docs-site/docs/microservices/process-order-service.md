# Process Order Service

The **Process Order Service** is responsible for executing the asynchronous background processing of orders in the Velure platform.

## Tech Stack

- **Language:** Go 1.25+
- **Framework:** `net/http` (Standard Library)
- **Message Broker:** RabbitMQ
- **Database:** None (State is managed by RabbitMQ messages and the Publish Order Service)
- **Port:** 8081

## Core Responsibilities

1. **Order Processing:** Acting as a background worker that consumes `order.created` messages from the `orders` queue in RabbitMQ.
2. **Inventory Validation:** Making synchronous HTTP calls to the **Product Service** to update and validate inventory before processing the order.
3. **Payment Logic:** Simulating payment processing and determining the final state of an order (`COMPLETED` or `FAILED`).
4. **Status Updates:** Publishing order status update events back to RabbitMQ for the **Publish Order Service** to process and notify the user.

## Downstream API Contract

- `PATCH /api/products/{id}/inventory`: Decrements inventory for each order item and returns an error when inventory is invalid or unavailable.

## Architecture & Conventions

The service uses standard Go `net/http` for internal health checks or lightweight routing if needed. It follows Clean Architecture principles (`handler/`, `service/`, `repository/` for RabbitMQ interactions) and utilizes the `velure-shared` module. It operates completely statelessly regarding databases, relying entirely on the message queue.
