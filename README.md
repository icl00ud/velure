# Velure E-Commerce Microservices Project

Main objective of this project is: learn :)

This repository contains a basic e-commerce system built using microservices architecture. The project provides essential functionalities, including user registration, login, and checkout processes. Below is an overview of the services and tools used in the project.

## Table of Contents
- [Services Overview](#services-overview)
- [Technologies Used](#technologies-used)
- [Setup Instructions](#setup-instructions)
- [Future Improvements](#future-improvements)

## Services Overview

### 1. **Auth Service**
- **Purpose:** Handles user authentication and authorization.
- **Stack:** NestJS, Prisma, PostgreSQL.
- **Features:**
  - User registration.
  - User login and JWT token generation.

### 2. **Product Service**
- **Purpose:** Manages product information and provides APIs for product data.
- **Stack:** NestJS, MongoDB, Redis.
- **Features:**
  - CRUD operations for products.

### 3. **Order Service**
- **Purpose:** Handles the checkout process and publishes user purchase orders to a RabbitMQ queue.
- **Stack:** GoLang, RabbitMQ.
- **Features:**
  - Receives purchase orders.
  - Publishes orders to RabbitMQ for further processing.

## Technologies Used
- **Programming Languages:** TypeScript, Go.
- **Frameworks:** NestJS, GoFr.
- **Databases:** PostgreSQL, MongoDB, Redis.
- **Queue:** RabbitMQ.
- **Containerization:** Docker, Kubernetes (Helm Charts).
- **Monitoring:** Prometheus, Grafana.
- **Automation:** Ansible, Vagrant.

## Setup Instructions

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/icl00ud/velure.git
   cd velure
   ```

2. Start services dependencies by using Docker Compose:
   ```bash
   docker-compose up
   ```

3. Start the services:
   - **Auth Service:**
     ```bash
     cd auth-service
     npm install
     npm run start
     ```
   - **Product Service:**
     ```bash
     cd product-service
     npm install
     npm run start
     ```
   - **Order Service:**
     ```bash
     cd order-service
     go run main.go
     ```

## Future Improvements
- Develop a Payment Service to consume orders from RabbitMQ and process payments.
- Add comprehensive test coverage for all services.
- Enhance observability with distributed tracing.
