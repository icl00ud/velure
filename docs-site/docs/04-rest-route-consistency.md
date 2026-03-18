---
sidebar_position: 4
---

# REST Route Consistency

This document captures a REST consistency audit across Velure services and defines a migration-safe path from legacy RPC-style routes to canonical REST endpoints.

## Why This Matters

The project currently mixes resource-oriented routes and action-oriented routes (for example: `create-order`, `update-order-status`, `getProductsByPage`). This makes APIs harder to reason about, increases client coupling, and introduces inconsistencies between services.

## External References Used

- Microsoft API design guidance: use nouns, plural collections, stable URI structure.
- Google API Improvement Proposals (AIP-121/AIP-122): resource-oriented naming and consistent hierarchical paths.
- RESTful API naming conventions: avoid verbs in URI paths, use query params for filtering/pagination.

## Current High-Impact Inconsistencies

### Auth Service

- Legacy action routes: `/register`, `/login`, `/logout/:refreshToken`, `/validateToken`.
- Redundant user segments: `/user/id/:id`, `/user/email/:email`.

### Product Service

- Verb/camelCase routes: `/getProductsByPage`, `/getProductsByPageAndCategory`, `/getProductsCount`, `/updateQuantity`.
- Duplicate semantics between list/search routes.

### Publish Order Service

- Repeated segment shape: `/api/order/user/order/status?id=X`.
- Action endpoints: `/create-order`, `/update-order-status`.
- Query-ID for singleton resource retrieval (`?id=`) instead of path params.

### Process Order Service

- Internal worker service is mostly clean, but downstream contract consumed includes non-REST action route (`/product/updateQuantity`).

## Canonical Endpoint Contract

### Orders

- `POST /api/orders`
- `GET /api/orders`
- `GET /api/me/orders`
- `GET /api/me/orders/{id}`
- `GET /api/me/orders/{id}/events` (SSE)
- `PATCH /api/orders/{id}/status`

### Products (target)

- `GET /api/products?page=&limit=&category=&search=`
- `GET /api/products/{id}`
- `POST /api/products`
- `PUT /api/products/{id}`
- `DELETE /api/products/{id}`
- `PATCH /api/products/{id}/inventory`

### Auth (target)

- `POST /api/sessions`
- `DELETE /api/sessions/current`
- `POST /api/users`
- `GET /api/users/{id}`
- `GET /api/users?email=`
- `POST /api/tokens/introspect`

## Migration Strategy

1. Introduce canonical routes as aliases.
2. Keep legacy routes for backward compatibility.
3. Move UI clients to canonical routes first.
4. Emit `Deprecation` and `Sunset` headers on legacy routes.
5. Remove aliases after telemetry confirms no legacy traffic.

## What Has Already Been Applied

- Canonical order routes were added in parallel with legacy aliases.
- UI order client was migrated to canonical order endpoints.
- Existing legacy routes remain available to avoid breaking integrations.

