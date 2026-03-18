# Product Service

The **Product Service** manages the product catalog for the Velure platform. It provides endpoints for retrieving product details and listing categories.

## Key Endpoints

- `GET /api/products`: Lists products with pagination and filtering.
- `GET /api/products/{id}`: Retrieves a specific product by ID.
- `GET /api/products/categories`: Lists available product categories.
- `GET /api/products/count`: Returns product count metadata.
- `POST /api/products`: Creates a product.
- `PUT /api/products/{id}`: Updates a product.
- `DELETE /api/products/{id}`: Deletes a product by ID.
- `PATCH /api/products/{id}/inventory`: Updates inventory quantity for a product.

## Tech Stack

- **Language:** Go 1.25+
- **Framework:** Fiber
- **Database:** MongoDB (using the native Go driver)
- **Port:** 3010

## Core Responsibilities

1. **Catalog Management:** Storing and managing product information, including titles, descriptions, pricing, and inventory details.
2. **Product Retrieval:** Providing fast and efficient APIs for the frontend to list products and retrieve single-product details.
3. **Inventory Checks:** Exposing inventory adjustment endpoints consumed by the **Process Order Service** during order processing.

## Architecture & Conventions

The service follows a Clean Architecture approach:
- `handler/`: HTTP routing and processing incoming requests (via Fiber).
- `service/`: Core logic for fetching product details from the datastore.
- `repository/`: MongoDB data access logic.

The service connects to MongoDB and utilizes the `velure-shared` module for logging and shared data constructs.
