# Order Status Update Flow

```mermaid
sequenceDiagram
    participant UI as UI Service
    participant POS as Publish Order Service
    participant RMQ as RabbitMQ
    participant PROS as Process Order Service
    participant PG as PostgreSQL

    Note over UI,PG: Create Order Flow
    UI->>POS: POST /create-order
    POS->>PG: INSERT order (status=CREATED)
    PG-->>POS: Order saved
    POS->>RMQ: Publish order.created event
    POS-->>UI: 201 Created

    Note over UI,PG: Process Order Flow
    RMQ->>PROS: Consume order.created
    PROS->>RMQ: Publish order.processing
    
    Note over POS,PG: Update to PROCESSING
    RMQ->>POS: Consume order.processing
    POS->>PG: UPDATE status=PROCESSING
    PG-->>POS: Updated
    
    Note over PROS: Simulate Payment (1-3s)
    PROS->>PROS: Process payment
    PROS->>RMQ: Publish order.completed
    
    Note over POS,PG: Update to COMPLETED
    RMQ->>POS: Consume order.completed
    POS->>PG: UPDATE status=COMPLETED
    PG-->>POS: Updated
    
    Note over UI,PG: Query Order Status
    UI->>POS: GET /orders
    POS->>PG: SELECT orders
    PG-->>POS: Orders with status
    POS-->>UI: Order list
```

## Components

### Publish Order Service
- **Publisher**: Publica `order.created` events
- **Consumer**: Consome `order.processing` e `order.completed` events
- **Repository**: Persiste pedidos no PostgreSQL

### Process Order Service
- **Consumer**: Consome `order.created` events
- **Publisher**: Publica `order.processing` e `order.completed` events
- **Service**: Simula processamento de pagamento

### RabbitMQ
- **Exchange**: `order_events` (topic)
- **Queues**:
  - `order_created_queue`: Para process-order-service
  - `publish-order-status-updates`: Para publish-order-service

## Event Flow

```
order.created
  ↓
[Process Order Service]
  ↓
order.processing → [Publish Order Service] → UPDATE status=PROCESSING
  ↓
[Payment Processing]
  ↓
order.completed → [Publish Order Service] → UPDATE status=COMPLETED
```

## Status Transitions

```
CREATED → PROCESSING → COMPLETED
  ↑           ↑            ↑
  |           |            |
  |           |            └─ order.completed event
  |           └─────────────── order.processing event
  └──────────────────────────── Initial creation
```
