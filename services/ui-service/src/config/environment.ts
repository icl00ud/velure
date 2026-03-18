export const environment = {
  production: true,
  PRODUCT_SERVICE_URL: import.meta.env.VITE_PRODUCT_SERVICE_URL || "/api/products",
  AUTHENTICATION_SERVICE_URL: import.meta.env.VITE_AUTHENTICATION_SERVICE_URL || "/api",
  ORDER_SERVICE_URL: import.meta.env.VITE_ORDER_SERVICE_URL || "/api/orders",
};
