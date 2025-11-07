export const environment = {
  production: true,
  PRODUCT_SERVICE_URL: import.meta.env.VITE_PRODUCT_SERVICE_URL || "/api/product",
  AUTHENTICATION_SERVICE_URL:
    import.meta.env.VITE_AUTHENTICATION_SERVICE_URL || "/api/auth",
  ORDER_SERVICE_URL: import.meta.env.VITE_ORDER_SERVICE_URL || "/api/order",
};
