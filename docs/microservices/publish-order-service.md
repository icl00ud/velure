# Publish Order Service

The **Publish Order Service** is responsible for initiating the order lifecycle and providing real-time status updates to the frontend via Server-Sent Events (SSE).

## Tech Stack

- **Language:** Go 1.25+
- **Framework:** `net/http` (Standard Library)
- **Database:** PostgreSQL (via `lib/pq` and raw SQL)
- **Message Broker:** RabbitMQ
- **Port:** 8080

## Core Responsibilities

1. **Order Creation:** Receiving new order requests from the frontend and saving the initial order state (e.g., `CREATED`) to PostgreSQL using raw SQL queries.
2. **Event Publishing:** Publishing an `order.created` event to the `orders` exchange on RabbitMQ to initiate asynchronous processing by the **Process Order Service**.
3. **Status Synchronization:** Consuming status update events from RabbitMQ and updating the PostgreSQL database with the latest order state.
4. **Real-time Notifications:** Broadcasting real-time order status updates to connected clients (frontend) using Server-Sent Events (SSE).

## Endpoints

- `POST /api/order/create-order`: Initiates a new order and publishes it to RabbitMQ.
- `GET /api/order/user/order/status?id=X`: Establishes an SSE connection to stream real-time order status updates back to the client.

## Architecture & Conventions

The service uses the `net/http` package for routing and raw SQL for database operations. It follows Clean Architecture principles (`handler/`, `service/`, `repository/`) and utilizes the internal `velure-shared` module.