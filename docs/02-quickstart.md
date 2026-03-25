---
sidebar_position: 2
---

# Quickstart

Get Velure up and running locally.

## Prerequisites

- Docker and Docker Compose
- Make
- Go 1.25+ (for local development)
- Node.js & npm (for UI development)

## Local Setup

1. Clone the repository and navigate to the root directory.
2. Run the following command to start all infrastructure, services, and monitoring:

```bash
make local-up
```

To stop and clean up, run:

```bash
make local-down
```

## ⚠️ Mandatory Configuration

**CRITICAL:** You must access the application via `https://velure.local`. Accessing containers directly via localhost will result in 405 errors because it bypasses the Caddy reverse proxy routing.

Add the following entry to your `/etc/hosts` file:

```bash
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

Accept any browser security warnings on first access (due to self-signed certificates).
