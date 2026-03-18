FROM oven/bun:1-alpine AS build

WORKDIR /app

COPY services/ui-service/package.json services/ui-service/bun.lockb ./

RUN bun install

COPY services/ui-service/ .

ARG VITE_PRODUCT_SERVICE_URL=/api/products
ARG VITE_AUTHENTICATION_SERVICE_URL=/api
ARG VITE_ORDER_SERVICE_URL=/api/orders

RUN echo "VITE_PRODUCT_SERVICE_URL=${VITE_PRODUCT_SERVICE_URL}" > .env
RUN echo "VITE_AUTHENTICATION_SERVICE_URL=${VITE_AUTHENTICATION_SERVICE_URL}" >> .env
RUN echo "VITE_ORDER_SERVICE_URL=${VITE_ORDER_SERVICE_URL}" >> .env

RUN bun run build

FROM nginx:alpine

RUN addgroup -g 1001 -S appuser && \
    adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G appuser -g appuser appuser

RUN rm /etc/nginx/conf.d/default.conf

COPY services/ui-service/nginx.conf /etc/nginx/conf.d/

COPY --from=build /app/dist /usr/share/nginx/html

RUN chown -R appuser:appuser /var/cache/nginx && \
    chown -R appuser:appuser /var/log/nginx && \
    chown -R appuser:appuser /etc/nginx/conf.d && \
    chown -R appuser:appuser /usr/share/nginx/html && \
    touch /var/run/nginx.pid && \
    chown -R appuser:appuser /var/run/nginx.pid

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
