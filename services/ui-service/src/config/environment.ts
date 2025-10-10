export const environment = {
  production: false,
  PRODUCT_SERVICE_URL: import.meta.env.VITE_PRODUCT_SERVICE_URL || "http://localhost:3010",
  AUTHENTICATION_SERVICE_URL:
    import.meta.env.VITE_AUTHENTICATION_SERVICE_URL || "http://localhost:3020",
  ORDER_SERVICE_URL: import.meta.env.VITE_ORDER_SERVICE_URL || "http://localhost:3030",
};
