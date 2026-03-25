# UI Service

The **UI Service** is the frontend application for the Velure platform, providing the user interface for browsing products, managing the cart, and tracking orders.

## Tech Stack

- **Framework:** React 18 + TypeScript + Vite
- **Routing:** React Router v6
- **State Management:** React Context (AuthContext) + TanStack React Query v5
- **Form Management:** React Hook Form + Zod schema validation
- **Styling / Components:** Radix UI + shadcn/ui + Tailwind CSS
- **Linting / Formatting:** Biome

## Core Responsibilities

1. **User Interface:** Rendering the e-commerce storefront, product catalogs, shopping cart, and order history.
2. **Authentication Flow:** Providing a seamless login and registration experience using the Auth Service, protecting routes via a `<ProtectedRoute>` wrapper.
3. **Real-time Updates:** Connecting to the **Publish Order Service** via Server-Sent Events (SSE) to display real-time order status updates.
4. **Data Fetching:** Utilizing TanStack React Query for efficient data fetching, caching, and state synchronization with backend services.

## Architecture & Conventions

- **Path Aliases:** The project uses `@` mapped to `./src` for cleaner imports (configured in `vite.config.ts` and `tsconfig.json`).
- **Styling:** Tailwind CSS is used for utility-first styling, combined with accessible, unstyled components from Radix UI and shadcn/ui.
- **Linting:** Biome handles both code formatting and linting (`biome.json`), enforcing rules like double quotes, semicolons, and ES5 trailing commas.

## Key Routes

- `/`: Home page
- `/login`: User authentication
- `/products`: Complete product catalog
- `/products/:category`: Category-specific products
- `/product/:id`: Individual product details
- `/cart`: User's shopping cart
- `/orders`: User's order history
- `/orders/:id`: Detailed view of a specific order (including real-time status)