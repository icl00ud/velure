import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const authRequests = new Counter('auth_requests');
const productRequests = new Counter('product_requests');
const orderRequests = new Counter('order_requests');
const uiRequests = new Counter('ui_requests');

// Environment configuration
// Can be set via: k6 run -e BASE_URL=https://velure.local integrated-load-test.js
const BASE_URL = __ENV.BASE_URL || 'http://localhost';
const AUTH_URL = __ENV.AUTH_URL || `${BASE_URL}:3020`;
const PRODUCT_URL = __ENV.PRODUCT_URL || `${BASE_URL}:3010`;
const ORDER_URL = __ENV.ORDER_URL || `${BASE_URL}:3030`;
const UI_URL = __ENV.UI_URL || `${BASE_URL}:80`;
const WARMUP_DURATION = __ENV.WARMUP_DURATION || '30s';
const TEST_DURATION = __ENV.TEST_DURATION || '15s';

console.log(`🎯 Target URLs:`);
console.log(`   Auth:    ${AUTH_URL}`);
console.log(`   Product: ${PRODUCT_URL}`);
console.log(`   Order:   ${ORDER_URL}`);
console.log(`   UI:      ${UI_URL}`);
console.log(`⏱️  Warmup: ${WARMUP_DURATION}, Test stages: ${TEST_DURATION} each`);

export const options = {
  stages: [
    { duration: WARMUP_DURATION, target: 10 },   // Warmup phase
    { duration: TEST_DURATION, target: 25 },     // Ramp up to 25 users
    { duration: TEST_DURATION, target: 75 },     // Ramp up to 75 users
    { duration: TEST_DURATION, target: 150 },    // Ramp up to 150 users
    { duration: TEST_DURATION, target: 250 },    // Ramp up to 250 users
    { duration: TEST_DURATION, target: 400 },    // Ramp up to 400 users
    { duration: TEST_DURATION, target: 500 },    // Peak load
    { duration: TEST_DURATION, target: 300 },    // Ramp down
    { duration: TEST_DURATION, target: 150 },    // Ramp down
    { duration: TEST_DURATION, target: 0 },      // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2000ms
    errors: ['rate<0.1'],              // Error rate must be less than 10%
    auth_requests: ['count>100'],      // At least 100 auth requests
    product_requests: ['count>200'],   // At least 200 product requests
    order_requests: ['count>50'],      // At least 50 order requests
    ui_requests: ['count>100'],        // At least 100 UI requests
  },
};

const services = {
  auth: AUTH_URL,
  product: PRODUCT_URL,
  order: ORDER_URL,
  ui: UI_URL
};

let authToken = '';
let productIds = [];
let orderIds = [];

export function setup() {
  console.log('Setting up integrated load test...');

  // Setup auth token
  const registerPayload = {
    email: `integration-test-${Date.now()}@example.com`,
    password: 'Test123!',
    name: 'Integration Test User'
  };

  const registerRes = http.post(`${services.auth}/authentication/register`, JSON.stringify(registerPayload), {
    headers: { 'Content-Type': 'application/json' },
  });

  let token = '';
  if (registerRes.status === 201) {
    const loginRes = http.post(`${services.auth}/authentication/login`, JSON.stringify({
      email: registerPayload.email,
      password: registerPayload.password
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    if (loginRes.status === 200) {
      const loginData = JSON.parse(loginRes.body);
      token = loginData.token;
    }
  }

  // Setup some products
  const sampleProducts = [
    { name: 'Integration Product 1', description: 'Test product', price: 29.99, category: 'Electronics', stock: 100 },
    { name: 'Integration Product 2', description: 'Another test product', price: 49.99, category: 'Books', stock: 50 }
  ];

  const createdProducts = [];
  for (const product of sampleProducts) {
    const res = http.post(`${services.product}/product/`, JSON.stringify(product), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (res.status === 201 || res.status === 200) {
      try {
        const productData = JSON.parse(res.body);
        createdProducts.push(productData);
      } catch (e) {
        // Continue if parsing fails
      }
    }
  }

  return { 
    token: token, 
    email: registerPayload.email,
    products: createdProducts 
  };
}

export default function (data) {
  // Distribute load across services with different weights
  const serviceWeights = {
    auth: 0.2,      // 20% auth requests
    product: 0.4,   // 40% product requests
    order: 0.2,     // 20% order requests
    ui: 0.2         // 20% UI requests
  };

  const random = Math.random();
  let selectedService = 'product'; // default
  
  let cumulativeWeight = 0;
  for (const [service, weight] of Object.entries(serviceWeights)) {
    cumulativeWeight += weight;
    if (random <= cumulativeWeight) {
      selectedService = service;
      break;
    }
  }

  switch (selectedService) {
    case 'auth':
      testAuthService(data);
      break;
    case 'product':
      testProductService(data);
      break;
    case 'order':
      testOrderService(data);
      break;
    case 'ui':
      testUIService(data);
      break;
  }

  sleep(Math.random() * 1 + 0.5); // Random sleep between 0.5-1.5 seconds
}

function testAuthService(data) {
  const scenarios = [
    () => testLogin(),
    () => testValidateToken(data.token),
    () => testGetUsers(data.token),
    () => testRegister(),
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();
  authRequests.add(1);
}

function testProductService(data) {
  const scenarios = [
    () => testGetAllProducts(),
    () => testGetProductsByName(),
    () => testCreateProduct(),
    () => testGetProductsCount(),
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();
  productRequests.add(1);
}

function testOrderService(data) {
  const scenarios = [
    () => testCreateOrder(),
    () => testUpdateOrderStatus(),
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();
  orderRequests.add(1);
}

function testUIService(data) {
  const scenarios = [
    () => testHomePage(),
    () => testStaticAssets(),
    () => testNavigationPages(),
  ];

  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();
  uiRequests.add(1);
}

// Auth service test functions
function testLogin() {
  const payload = { email: 'test@example.com', password: 'wrongpassword' };
  const res = http.post(`${services.auth}/authentication/login`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'auth login response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testValidateToken(token) {
  if (!token) return;
  const res = http.post(`${services.auth}/authentication/validateToken`, JSON.stringify({ token }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'auth validate response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testGetUsers(token) {
  if (!token) return;
  const res = http.get(`${services.auth}/authentication/users`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  
  const success = check(res, { 'auth get users response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testRegister() {
  const payload = {
    email: `user-${Math.random().toString(36).substring(7)}@example.com`,
    password: 'Test123!',
    name: 'Load Test User'
  };
  const res = http.post(`${services.auth}/authentication/register`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'auth register response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

// Product service test functions
function testGetAllProducts() {
  const res = http.get(`${services.product}/product/`);
  const success = check(res, { 'product get all response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testGetProductsByName() {
  const names = ['Integration Product 1', 'Laptop', 'Phone'];
  const name = names[Math.floor(Math.random() * names.length)];
  const res = http.get(`${services.product}/product/getProductsByName/${encodeURIComponent(name)}`);
  const success = check(res, { 'product get by name response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testCreateProduct() {
  const product = {
    name: `Load Test Product ${Math.random().toString(36).substring(7)}`,
    description: 'Product created during load testing',
    price: Math.round((Math.random() * 100 + 10) * 100) / 100,
    category: ['Electronics', 'Books', 'Clothing'][Math.floor(Math.random() * 3)],
    stock: Math.floor(Math.random() * 100) + 1
  };
  
  const res = http.post(`${services.product}/product/`, JSON.stringify(product), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'product create response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testGetProductsCount() {
  const res = http.get(`${services.product}/product/getProductsCount`);
  const success = check(res, { 'product count response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

// Order service test functions
function testCreateOrder() {
  const order = {
    customer_id: `customer-${Math.floor(Math.random() * 100)}`,
    items: [
      { product_id: 'product-1', quantity: 2, price: 29.99 }
    ],
    shipping_address: {
      street: '123 Test St',
      city: 'Test City',
      state: 'TS',
      zip: '12345',
      country: 'US'
    }
  };
  
  const res = http.post(`${services.order}/create-order`, JSON.stringify(order), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'order create response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testUpdateOrderStatus() {
  const orderId = `order-${Math.random().toString(36).substring(7)}`;
  const statusUpdate = {
    order_id: orderId,
    status: 'processing',
    updated_by: 'load-test'
  };
  
  const res = http.post(`${services.order}/update-order-status`, JSON.stringify(statusUpdate), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const success = check(res, { 'order update response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

// UI service test functions
function testHomePage() {
  const res = http.get(`${services.ui}/`);
  const success = check(res, { 'ui home response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testStaticAssets() {
  const assets = ['/favicon.ico', '/robots.txt'];
  const asset = assets[Math.floor(Math.random() * assets.length)];
  const res = http.get(`${services.ui}${asset}`);
  const success = check(res, { 'ui asset response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

function testNavigationPages() {
  const pages = ['/products', '/orders', '/login'];
  const page = pages[Math.floor(Math.random() * pages.length)];
  const res = http.get(`${services.ui}${page}`);
  const success = check(res, { 'ui navigation response': (r) => r.status !== 0 });
  errorRate.add(!success);
}

export function handleSummary(data) {
  return {
    'integrated-load-test-summary.html': htmlReport(data),
    'integrated-load-test-summary.json': JSON.stringify(data),
  };
}

function htmlReport(data) {
  return `
<!DOCTYPE html>
<html>
<head>
    <title>Integrated Load Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { margin: 10px 0; padding: 10px; background-color: #f5f5f5; border-radius: 5px; }
        .error { background-color: #ffebee; }
        .success { background-color: #e8f5e8; }
        .service-stats { display: flex; gap: 20px; flex-wrap: wrap; }
        .service { border: 1px solid #ddd; padding: 15px; border-radius: 5px; min-width: 200px; }
    </style>
</head>
<body>
    <h1>Velure Microservices - Integrated Load Test Results</h1>
    
    <div class="metric">
        <h3>Overall Test Summary</h3>
        <p><strong>Duration:</strong> ${Math.round(data.state.testRunDurationMs / 1000)}s</p>
        <p><strong>Total Requests:</strong> ${data.metrics.http_reqs.values.count}</p>
        <p><strong>Request Rate:</strong> ${Math.round(data.metrics.http_reqs.values.rate * 100) / 100} req/s</p>
    </div>

    <div class="metric ${data.metrics.http_req_duration.values.p95 < 2000 ? 'success' : 'error'}">
        <h3>Performance</h3>
        <p><strong>95th Percentile Response Time:</strong> ${Math.round(data.metrics.http_req_duration.values.p95 * 100) / 100}ms</p>
        <p><strong>Average Response Time:</strong> ${Math.round(data.metrics.http_req_duration.values.avg * 100) / 100}ms</p>
    </div>

    <div class="metric ${data.metrics.errors ? (data.metrics.errors.values.rate < 0.1 ? 'success' : 'error') : 'success'}">
        <h3>Reliability</h3>
        <p><strong>Error Rate:</strong> ${data.metrics.errors ? Math.round(data.metrics.errors.values.rate * 10000) / 100 : 0}%</p>
        <p><strong>Successful Requests:</strong> ${data.metrics.http_reqs.values.count - (data.metrics.errors ? data.metrics.errors.values.count : 0)}</p>
    </div>

    <h2>Service Distribution</h2>
    <div class="service-stats">
        <div class="service">
            <h4>Auth Service</h4>
            <p>Requests: ${data.metrics.auth_requests ? data.metrics.auth_requests.values.count : 0}</p>
        </div>
        <div class="service">
            <h4>Product Service</h4>
            <p>Requests: ${data.metrics.product_requests ? data.metrics.product_requests.values.count : 0}</p>
        </div>
        <div class="service">
            <h4>Order Service</h4>
            <p>Requests: ${data.metrics.order_requests ? data.metrics.order_requests.values.count : 0}</p>
        </div>
        <div class="service">
            <h4>UI Service</h4>
            <p>Requests: ${data.metrics.ui_requests ? data.metrics.ui_requests.values.count : 0}</p>
        </div>
    </div>

    <div class="metric">
        <h3>Thresholds Status</h3>
        <p>All thresholds: ${Object.values(data.thresholds).every(t => !t.violated) ? '✅ PASSED' : '❌ FAILED'}</p>
    </div>
</body>
</html>
  `;
}