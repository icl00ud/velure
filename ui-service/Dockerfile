# Build stage
FROM oven/bun:1-alpine AS build

WORKDIR /app

# Copy package files
COPY package.json bun.lockb ./

# Install dependencies
RUN bun install

# Copy source code
COPY . .

# Create .env file with build-time variables
ARG VITE_PRODUCT_SERVICE_URL=http://localhost:3010
ARG VITE_AUTHENTICATION_SERVICE_URL=http://localhost:3020
ARG VITE_ORDER_SERVICE_URL=http://localhost:3030

RUN echo "VITE_PRODUCT_SERVICE_URL=${VITE_PRODUCT_SERVICE_URL}" > .env
RUN echo "VITE_AUTHENTICATION_SERVICE_URL=${VITE_AUTHENTICATION_SERVICE_URL}" >> .env
RUN echo "VITE_ORDER_SERVICE_URL=${VITE_ORDER_SERVICE_URL}" >> .env

# Build the application
RUN bun run build

# Production stage
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/

# Copy built app from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
