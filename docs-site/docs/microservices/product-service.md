# Product Service

The **Product Service** manages the product catalog for the Velure platform. It provides endpoints for retrieving product details and listing categories.

## Tech Stack

- **Language:** Go 1.25+
- **Framework:** Fiber
- **Database:** MongoDB (using the native Go driver)
- **Port:** 3010

## Core Responsibilities

1. **Catalog Management:** Storing and managing product information, including titles, descriptions, pricing, and inventory details.
2. **Product Retrieval:** Providing fast and efficient APIs for the frontend to list products and retrieve single-product details.
3. **Inventory Checks:** Providing an internal HTTP API endpoint used synchronously by the **Process Order Service** to verify product availability before processing orders.

## Architecture & Conventions

The service follows a Clean Architecture approach:
- `handler/`: HTTP routing and processing incoming requests (via Fiber).
- `service/`: Core logic for fetching product details from the datastore.
- `repository/`: MongoDB data access logic.

The service connects to MongoDB and utilizes the `velure-shared` module for logging and shared data constructs.