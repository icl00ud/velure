---
sidebar_position: 4
---

# REST Route Consistency

This document defines the canonical REST route contract across Velure Microservices.

## Status

Route standardization is complete. Canonical resource-oriented REST endpoints are now the source of truth for all documented service contracts.

## External References Used

- Microsoft API design guidance: use nouns, plural collections, stable URI structure.
- Google API Improvement Proposals (AIP-121/AIP-122): resource-oriented naming and consistent hierarchical paths.
- RESTful API naming conventions: avoid verbs in URI paths, use query params for filtering/pagination.

## Canonical Endpoint Contract

### Orders

- `POST /api/orders`
- `GET /api/orders`
- `GET /api/me/orders`
- `GET /api/me/orders/{id}`
- `GET /api/me/orders/{id}/events` (SSE)
- `PATCH /api/orders/{id}/status`

### Products (target)

- `GET /api/products?page=&limit=&category=&q=`
- `GET /api/products/{id}`
- `GET /api/products/categories`
- `GET /api/products/count`
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

## Notes

- Service documentation should reference only the canonical routes listed above.
- New endpoint design must remain resource-oriented (nouns in paths, HTTP methods for actions).
