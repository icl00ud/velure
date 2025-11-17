// Velure - All Microservices Load Test
// Tests all backend services simultaneously (excludes UI service)
// Services: auth-service, product-service, publish-order-service, process-order-service

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { Counter, Rate, Trend } from 'k6/metrics';

// ============================================================================
// CONFIGURATION
// ============================================================================

const BASE_URL = __ENV.BASE_URL || 'https://velure.local';
const AUTH_URL = __ENV.AUTH_URL || `${BASE_URL}/api/auth`;
const PRODUCT_URL = __ENV.PRODUCT_URL || `${BASE_URL}/api/product`;
const ORDER_URL = __ENV.ORDER_URL || `${BASE_URL}/api/order`;

const WARMUP_DURATION = __ENV.WARMUP_DURATION || '30s';
const TEST_DURATION = __ENV.TEST_DURATION || '2m';
const COOLDOWN_DURATION = __ENV.COOLDOWN_DURATION || '30s';

// ============================================================================
// CUSTOM METRICS
// ============================================================================

const authRequests = new Counter('auth_requests_total');
const authErrors = new Rate('auth_error_rate');
const authDuration = new Trend('auth_request_duration');

const productRequests = new Counter('product_requests_total');
const productErrors = new Rate('product_error_rate');
const productDuration = new Trend('product_request_duration');

const orderRequests = new Counter('order_requests_total');
const orderErrors = new Rate('order_error_rate');
const orderDuration = new Trend('order_request_duration');

// ============================================================================
// TEST OPTIONS
// ============================================================================

export const options = {
  scenarios: {
    // Scenario 1: Auth Service Load
    auth_load: {
      executor: 'ramping-vus',
      exec: 'authScenario',
      startVUs: 0,
      stages: [
        { duration: WARMUP_DURATION, target: 20 },   // Warmup
        { duration: TEST_DURATION, target: 100 },     // Ramp up
        { duration: TEST_DURATION, target: 200 },     // Peak load
        { duration: COOLDOWN_DURATION, target: 0 },   // Cool down
      ],
      gracefulRampDown: '10s',
      tags: { service: 'auth' },
    },

    // Scenario 2: Product Service Load
    product_load: {
      executor: 'ramping-vus',
      exec: 'productScenario',
      startVUs: 0,
      stages: [
        { duration: WARMUP_DURATION, target: 30 },   // Warmup
        { duration: TEST_DURATION, target: 200 },     // Ramp up
        { duration: TEST_DURATION, target: 400 },     // Peak load
        { duration: COOLDOWN_DURATION, target: 0 },   // Cool down
      ],
      gracefulRampDown: '10s',
      tags: { service: 'product' },
    },

    // Scenario 3: Order Creation Load (Publish Order Service)
    order_publish_load: {
      executor: 'ramping-vus',
      exec: 'orderPublishScenario',
      startVUs: 0,
      stages: [
        { duration: WARMUP_DURATION, target: 50 },   // Warmup
        { duration: TEST_DURATION, target: 300 },     // Ramp up
        { duration: TEST_DURATION, target: 500 },     // Peak load
        { duration: COOLDOWN_DURATION, target: 0 },   // Cool down
      ],
      gracefulRampDown: '10s',
      tags: { service: 'publish-order' },
    },

    // Scenario 4: Integrated User Journey
    user_journey: {
      executor: 'ramping-vus',
      exec: 'userJourneyScenario',
      startVUs: 0,
      stages: [
        { duration: WARMUP_DURATION, target: 10 },   // Warmup
        { duration: TEST_DURATION, target: 50 },      // Ramp up
        { duration: TEST_DURATION, target: 100 },     // Peak load
        { duration: COOLDOWN_DURATION, target: 0 },   // Cool down
      ],
      gracefulRampDown: '10s',
      tags: { service: 'integrated' },
    },
  },

  thresholds: {
    // Global thresholds
    'http_req_duration': ['p(95)<2000', 'p(99)<5000'],
    'http_req_failed': ['rate<0.1'], // Less than 10% errors

    // Auth service thresholds
    'http_req_duration{service:auth}': ['p(95)<500'],
    'auth_error_rate': ['rate<0.1'],

    // Product service thresholds
    'http_req_duration{service:product}': ['p(95)<1000'],
    'product_error_rate': ['rate<0.05'],

    // Order service thresholds
    'http_req_duration{service:publish-order}': ['p(95)<2000'],
    'order_error_rate': ['rate<0.1'],
  },

  // Disable SSL verification for local development
  insecureSkipTLSVerify: true,
};

// ============================================================================
// TEST DATA
// ============================================================================

const usernames = new SharedArray('usernames', function () {
  const users = [];
  for (let i = 1; i <= 1000; i++) {
    users.push(`loadtest_user_${i}_${Date.now()}`);
  }
  return users;
});

const productIds = new SharedArray('productIds', function () {
  // These should be actual product IDs from your database
  // For now, using placeholder IDs
  return [
    '67a2c1e5564dfbe318544ca7',
    '67a2c1e5564dfbe318544ca8',
    '67a2c1e5564dfbe318544ca9',
    '67a2c1e5564dfbe318544caa',
  ];
});

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

function getRandomUser() {
  return usernames[Math.floor(Math.random() * usernames.length)];
}

function getRandomProductId() {
  return productIds[Math.floor(Math.random() * productIds.length)];
}

function getHeaders(token = null) {
  const headers = {
    'Content-Type': 'application/json',
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

// ============================================================================
// SCENARIO 1: AUTH SERVICE
// ============================================================================

export function authScenario() {
  const username = getRandomUser();
  const email = `${username}@loadtest.com`;
  const password = 'LoadTest123!';

  group('Auth Service - User Registration & Login', function () {
    // Register new user
    group('Register User', function () {
      const startTime = Date.now();
      const registerPayload = JSON.stringify({
        username: username,
        email: email,
        password: password,
      });

      const registerRes = http.post(
        `${AUTH_URL}/register`,
        registerPayload,
        { headers: getHeaders(), tags: { service: 'auth', endpoint: 'register' } }
      );

      authRequests.add(1);
      authDuration.add(Date.now() - startTime);

      const registerSuccess = check(registerRes, {
        'register: status 201': (r) => r.status === 201,
        'register: has token': (r) => r.json('token') !== undefined,
      });

      authErrors.add(!registerSuccess);
    });

    sleep(0.5);

    // Login with created user
    group('Login User', function () {
      const startTime = Date.now();
      const loginPayload = JSON.stringify({
        email: email,
        password: password,
      });

      const loginRes = http.post(
        `${AUTH_URL}/login`,
        loginPayload,
        { headers: getHeaders(), tags: { service: 'auth', endpoint: 'login' } }
      );

      authRequests.add(1);
      authDuration.add(Date.now() - startTime);

      const loginSuccess = check(loginRes, {
        'login: status 200': (r) => r.status === 200,
        'login: has token': (r) => r.json('token') !== undefined,
      });

      authErrors.add(!loginSuccess);

      if (loginSuccess && loginRes.json('token')) {
        // Validate token
        const token = loginRes.json('token');
        const validateRes = http.get(
          `${AUTH_URL}/validate`,
          { headers: getHeaders(token), tags: { service: 'auth', endpoint: 'validate' } }
        );

        authRequests.add(1);
        const validateSuccess = check(validateRes, {
          'validate: status 200': (r) => r.status === 200,
        });
        authErrors.add(!validateSuccess);
      }
    });
  });

  sleep(1);
}

// ============================================================================
// SCENARIO 2: PRODUCT SERVICE
// ============================================================================

export function productScenario() {
  group('Product Service - Browse & Search', function () {
    // List products
    group('List Products', function () {
      const startTime = Date.now();
      const listRes = http.get(
        `${PRODUCT_URL}/products?page=1&limit=20`,
        { headers: getHeaders(), tags: { service: 'product', endpoint: 'list' } }
      );

      productRequests.add(1);
      productDuration.add(Date.now() - startTime);

      const listSuccess = check(listRes, {
        'list products: status 200': (r) => r.status === 200,
        'list products: has data': (r) => r.json('products') !== undefined,
      });

      productErrors.add(!listSuccess);
    });

    sleep(0.3);

    // Search products
    group('Search Products', function () {
      const searchTerms = ['product', 'test', 'item', 'goods'];
      const term = searchTerms[Math.floor(Math.random() * searchTerms.length)];

      const startTime = Date.now();
      const searchRes = http.get(
        `${PRODUCT_URL}/products/search?q=${term}`,
        { headers: getHeaders(), tags: { service: 'product', endpoint: 'search' } }
      );

      productRequests.add(1);
      productDuration.add(Date.now() - startTime);

      const searchSuccess = check(searchRes, {
        'search products: status 200': (r) => r.status === 200,
      });

      productErrors.add(!searchSuccess);
    });

    sleep(0.3);

    // Get product by ID
    group('Get Product Details', function () {
      const productId = getRandomProductId();
      const startTime = Date.now();

      const getRes = http.get(
        `${PRODUCT_URL}/products/${productId}`,
        { headers: getHeaders(), tags: { service: 'product', endpoint: 'get' } }
      );

      productRequests.add(1);
      productDuration.add(Date.now() - startTime);

      const getSuccess = check(getRes, {
        'get product: status is 200 or 404': (r) => r.status === 200 || r.status === 404,
      });

      productErrors.add(!getSuccess);
    });
  });

  sleep(1);
}

// ============================================================================
// SCENARIO 3: ORDER PUBLISH SERVICE
// ============================================================================

export function orderPublishScenario() {
  group('Order Service - Create Orders', function () {
    // 70% create orders, 30% list orders
    const action = Math.random();

    if (action < 0.7) {
      // Create order
      group('Create Order', function () {
        const orderPayload = JSON.stringify({
          items: [
            {
              product_id: getRandomProductId(),
              name: `Product ${Math.floor(Math.random() * 100)}`,
              quantity: Math.floor(Math.random() * 5) + 1,
              price: Math.floor(Math.random() * 100) + 10.99,
            },
            {
              product_id: getRandomProductId(),
              name: `Product ${Math.floor(Math.random() * 100)}`,
              quantity: Math.floor(Math.random() * 3) + 1,
              price: Math.floor(Math.random() * 50) + 5.99,
            },
          ],
        });

        const startTime = Date.now();
        const createRes = http.post(
          `${ORDER_URL}/create-order`,
          orderPayload,
          { headers: getHeaders(), tags: { service: 'publish-order', endpoint: 'create' } }
        );

        orderRequests.add(1);
        orderDuration.add(Date.now() - startTime);

        const createSuccess = check(createRes, {
          'create order: status 201': (r) => r.status === 201,
          'create order: has order_id': (r) => r.json('order_id') !== undefined,
        });

        orderErrors.add(!createSuccess);
      });
    } else {
      // List orders
      group('List Orders', function () {
        const startTime = Date.now();
        const listRes = http.get(
          `${ORDER_URL}/orders?page=1&limit=10`,
          { headers: getHeaders(), tags: { service: 'publish-order', endpoint: 'list' } }
        );

        orderRequests.add(1);
        orderDuration.add(Date.now() - startTime);

        const listSuccess = check(listRes, {
          'list orders: status 200': (r) => r.status === 200,
        });

        orderErrors.add(!listSuccess);
      });
    }
  });

  sleep(0.5);
}

// ============================================================================
// SCENARIO 4: INTEGRATED USER JOURNEY
// ============================================================================

export function userJourneyScenario() {
  const username = getRandomUser();
  const email = `${username}@loadtest.com`;
  const password = 'LoadTest123!';
  let authToken = null;

  group('Complete User Journey', function () {
    // Step 1: Register
    group('1. User Registration', function () {
      const registerRes = http.post(
        `${AUTH_URL}/register`,
        JSON.stringify({ username, email, password }),
        { headers: getHeaders(), tags: { journey: 'step1' } }
      );

      check(registerRes, {
        'journey: registration successful': (r) => r.status === 201,
      });

      if (registerRes.status === 201) {
        authToken = registerRes.json('token');
      }
    });

    sleep(1);

    // Step 2: Browse products
    group('2. Browse Products', function () {
      const listRes = http.get(
        `${PRODUCT_URL}/products?page=1&limit=20`,
        { headers: getHeaders(), tags: { journey: 'step2' } }
      );

      check(listRes, {
        'journey: product list loaded': (r) => r.status === 200,
      });
    });

    sleep(2);

    // Step 3: Search for specific product
    group('3. Search Products', function () {
      const searchRes = http.get(
        `${PRODUCT_URL}/products/search?q=product`,
        { headers: getHeaders(), tags: { journey: 'step3' } }
      );

      check(searchRes, {
        'journey: search completed': (r) => r.status === 200,
      });
    });

    sleep(1);

    // Step 4: Create order
    group('4. Create Order', function () {
      const orderRes = http.post(
        `${ORDER_URL}/create-order`,
        JSON.stringify({
          items: [
            {
              product_id: getRandomProductId(),
              name: 'Selected Product',
              quantity: 2,
              price: 29.99,
            },
          ],
        }),
        { headers: getHeaders(), tags: { journey: 'step4' } }
      );

      check(orderRes, {
        'journey: order created': (r) => r.status === 201,
      });
    });

    sleep(2);

    // Step 5: View orders
    group('5. View Orders', function () {
      const ordersRes = http.get(
        `${ORDER_URL}/orders`,
        { headers: getHeaders(), tags: { journey: 'step5' } }
      );

      check(ordersRes, {
        'journey: orders retrieved': (r) => r.status === 200,
      });
    });
  });

  sleep(3);
}

// ============================================================================
// TEST LIFECYCLE
// ============================================================================

export function setup() {
  console.log('🚀 Starting Velure Load Test - All Microservices');
  console.log(`📍 Base URL: ${BASE_URL}`);
  console.log(`⏱️  Warmup: ${WARMUP_DURATION} | Test: ${TEST_DURATION} | Cooldown: ${COOLDOWN_DURATION}`);
  console.log('');
  console.log('📊 Testing services:');
  console.log('   ✓ Auth Service');
  console.log('   ✓ Product Service');
  console.log('   ✓ Publish Order Service');
  console.log('   ✓ Process Order Service (indirectly via message queue)');
  console.log('');
}

export function teardown(data) {
  console.log('');
  console.log('✅ Load test completed');
  console.log('📈 Check Grafana dashboards for detailed metrics');
  console.log('   → http://localhost:3000');
}
