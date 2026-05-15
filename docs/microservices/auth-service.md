# Auth Service

The **Auth Service** is responsible for user authentication and authorization within the Velure platform. It handles user registration, login, and token generation/validation.

## Tech Stack

- **Language:** Go 1.25+
- **Framework:** Gin
- **Database:** PostgreSQL (via GORM)
- **Cache / Store:** Redis
- **Port:** 3020

## Core Responsibilities

1. **User Management:** Securely storing user credentials and managing user profiles using PostgreSQL.
2. **Authentication:** Providing endpoints for user registration and login.
3. **Session Management:** Issuing JWT (JSON Web Tokens) for authenticated sessions and managing token caching/invalidation via Redis.

## Key Endpoints

- `POST /api/users`: Registers a new user.
- `GET /api/users`: Lists users (paginated).
- `GET /api/users/:id`: Retrieves a user by ID.
- `GET /api/users/email/:email`: Retrieves a user by email.
- `POST /api/sessions`: Authenticates a user and returns a JWT.
- `DELETE /api/sessions/current`: Logs out the current session (invalidates token).
- `POST /api/tokens/introspect`: Validates a JWT and returns its claims.

## Architecture & Conventions

The service follows a Clean Architecture layered design:
- `handler/`: HTTP request handling and validation.
- `service/`: Core business logic for authentication and token generation.
- `repository/`: Data access layer for PostgreSQL and Redis interactions.

It utilizes the shared internal module (`velure-shared`) for structured logging and shared data models.