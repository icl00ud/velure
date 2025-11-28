import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const registrationFailureRate = new Rate('registration_failures');
const loginFailureRate = new Rate('login_failures');
const productListFailureRate = new Rate('product_list_failures');
const orderCreationFailureRate = new Rate('order_creation_failures');
const endToEndDuration = new Trend('end_to_end_duration');
const orderTotal = new Trend('order_total_value');
const ordersCreated = new Counter('orders_created_total');

// Load test configuration
// Progressive ramp: 10 → 150 VUs over 7 minutes, then sustained load
export const options = {
  stages: [
    { duration: '1m', target: 10 },   // Warmup: Start with 10 users
    { duration: '1m', target: 30 },   // Ramp: 10 → 30 users
    { duration: '1m', target: 50 },   // Ramp: 30 → 50 users
    { duration: '1m', target: 70 },   // Ramp: 50 → 70 users
    { duration: '1m', target: 90 },   // Ramp: 70 → 90 users
    { duration: '1m', target: 110 },  // Ramp: 90 → 110 users
    { duration: '1m', target: 130 },  // Ramp: 110 → 130 users
    { duration: '1m', target: 150 },  // Ramp: 130 → 150 users (peak)
    { duration: '2m', target: 150 },  // Sustained: Hold at 150 users
    { duration: '1m', target: 0 },    // Ramp down: 150 → 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% of requests should be below 3s (relaxed for high load)
    http_req_failed: ['rate<0.15'],    // Error rate should be below 15% (relaxed for high load)
    registration_failures: ['rate<0.15'],
    login_failures: ['rate<0.15'],
    product_list_failures: ['rate<0.15'],
    order_creation_failures: ['rate<0.25'], // Allow higher failure rate at peak load
    end_to_end_duration: ['p(95)<15000'], // 95% of full journeys should complete in 15s
  },
};

// Get base URL from environment variable or use default
const BASE_URL = __ENV.BASE_URL || 'https://your-k8s-ingress-url.com';

// Helper function to generate unique user data
function generateUserData() {
  const timestamp = Date.now();
  const vu = __VU;
  const iter = __ITER;
  const uniqueId = `${timestamp}_${vu}_${iter}`;

  return {
    name: `User ${uniqueId}`,
    email: `user_${uniqueId}@test.com`,
    password: 'Test@123456',
  };
}

// Helper function to select random products
function selectRandomProducts(products, minItems = 1, maxItems = 3) {
  const numItems = Math.floor(Math.random() * (maxItems - minItems + 1)) + minItems;
  const selectedProducts = [];

  // Create a copy of products array to avoid modifying original
  const availableProducts = [...products];

  for (let i = 0; i < numItems && availableProducts.length > 0; i++) {
    const randomIndex = Math.floor(Math.random() * availableProducts.length);
    const product = availableProducts.splice(randomIndex, 1)[0];

    selectedProducts.push({
      product_id: product._id,
      name: product.name,
      quantity: Math.floor(Math.random() * 3) + 1, // 1-3 items
      price: product.price,
    });
  }

  return selectedProducts;
}

export default function () {
  const startTime = Date.now();
  const userData = generateUserData();
  let accessToken = '';

  // Step 1: Register user
  const registerPayload = JSON.stringify(userData);
  const registerParams = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Register' },
  };

  const registerRes = http.post(
    `${BASE_URL}/api/auth/register`,
    registerPayload,
    registerParams
  );

  const registerSuccess = check(registerRes, {
    'registration status is 201': (r) => r.status === 201,
    'registration returns access token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.accessToken !== undefined;
      } catch {
        return false;
      }
    },
  });

  registrationFailureRate.add(!registerSuccess);

  if (!registerSuccess) {
    console.error(`Registration failed: ${registerRes.status} - ${registerRes.body}`);
    return;
  }

  // Extract access token from registration response
  try {
    const registerBody = JSON.parse(registerRes.body);
    accessToken = registerBody.accessToken;
  } catch (e) {
    console.error(`Failed to parse registration response: ${e}`);
    return;
  }

  sleep(1);

  // Step 2: Login (optional, since registration already returns token)
  // But we'll do it anyway to test the login endpoint
  const loginPayload = JSON.stringify({
    email: userData.email,
    password: userData.password,
  });

  const loginParams = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Login' },
  };

  const loginRes = http.post(
    `${BASE_URL}/api/auth/login`,
    loginPayload,
    loginParams
  );

  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login returns access token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.accessToken !== undefined;
      } catch {
        return false;
      }
    },
  });

  loginFailureRate.add(!loginSuccess);

  if (!loginSuccess) {
    console.error(`Login failed: ${loginRes.status} - ${loginRes.body}`);
    return;
  }

  // Update access token with login response
  try {
    const loginBody = JSON.parse(loginRes.body);
    accessToken = loginBody.accessToken;
  } catch (e) {
    console.error(`Failed to parse login response: ${e}`);
    return;
  }

  sleep(1);

  // Step 3: List products
  // Note: Product API doesn't require authentication for browsing
  // Response format: { products: [...], totalCount, page, pageSize, totalPages }
  const productParams = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'ListProducts' },
  };

  const productRes = http.get(
    `${BASE_URL}/api/product/products?page=1&limit=20`,
    productParams
  );

  const productSuccess = check(productRes, {
    'product list status is 200': (r) => r.status === 200,
    'product list returns items': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.products && body.products.length > 0;
      } catch {
        return false;
      }
    },
  });

  productListFailureRate.add(!productSuccess);

  if (!productSuccess) {
    console.error(`Product list failed: ${productRes.status} - ${productRes.body}`);
    return;
  }

  let products = [];
  try {
    const productBody = JSON.parse(productRes.body);
    products = productBody.products || [];
  } catch (e) {
    console.error(`Failed to parse product response: ${e}`);
    return;
  }

  if (products.length === 0) {
    console.error('No products available to order');
    return;
  }

  sleep(2); // User browsing products

  // Step 4: Create order with random products
  // Endpoint: POST /api/order/create-order
  // Response format: { order_id: "uuid", total: 123.45, status: "CREATED" }
  const orderItems = selectRandomProducts(products, 1, 3);
  const orderPayload = JSON.stringify({ items: orderItems });

  const orderParams = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`,
    },
    tags: { name: 'CreateOrder' },
  };

  const orderRes = http.post(
    `${BASE_URL}/api/order/create-order`,
    orderPayload,
    orderParams
  );

  const orderSuccess = check(orderRes, {
    'order creation status is 201': (r) => r.status === 201,
    'order returns order_id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.order_id !== undefined;
      } catch {
        return false;
      }
    },
    'order returns total': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total !== undefined && body.total > 0;
      } catch {
        return false;
      }
    },
  });

  orderCreationFailureRate.add(!orderSuccess);

  if (orderSuccess) {
    ordersCreated.add(1);
    try {
      const orderBody = JSON.parse(orderRes.body);
      orderTotal.add(orderBody.total);
    } catch (e) {
      console.error(`Failed to parse order response: ${e}`);
    }
  } else {
    console.error(`Order creation failed: ${orderRes.status} - ${orderRes.body}`);
  }

  // Record end-to-end duration
  const duration = Date.now() - startTime;
  endToEndDuration.add(duration);

  sleep(1);
}

// Setup function (runs once per VU at the beginning)
export function setup() {
  console.log('🚀 Starting user journey load test');
  console.log(`📍 Base URL: ${BASE_URL}`);
  console.log('📊 Test will simulate: Register → Login → Browse Products → Create Order');

  // Verify the base URL is accessible
  const healthCheck = http.get(`${BASE_URL}/health`, {
    timeout: '10s',
  });

  if (healthCheck.status !== 200 && healthCheck.status !== 404) {
    console.warn(`⚠️  Health check returned status ${healthCheck.status}`);
  }

  return { timestamp: Date.now() };
}

// Teardown function (runs once at the end)
export function teardown(data) {
  console.log('✅ Load test completed');
  console.log(`⏱️  Test duration: ${(Date.now() - data.timestamp) / 1000}s`);
}
