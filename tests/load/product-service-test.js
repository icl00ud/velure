import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Environment configuration
// Can be set via: k6 run -e PRODUCT_URL=https://velure.local/api/product product-service-test.js
const PRODUCT_URL = __ENV.PRODUCT_URL || __ENV.BASE_URL || 'http://localhost:3010';
const WARMUP_DURATION = __ENV.WARMUP_DURATION || '30s';
const TEST_DURATION = __ENV.TEST_DURATION || '15s';

console.log(`🎯 Target URL: ${PRODUCT_URL}`);
console.log(`⏱️  Warmup: ${WARMUP_DURATION}, Test stages: ${TEST_DURATION} each`);

export const options = {
  stages: [
    { duration: WARMUP_DURATION, target: 10 },   // Warmup phase
    { duration: TEST_DURATION, target: 20 },     // Ramp up to 20 users
    { duration: TEST_DURATION, target: 50 },     // Ramp up to 50 users
    { duration: TEST_DURATION, target: 100 },    // Ramp up to 100 users
    { duration: TEST_DURATION, target: 200 },    // Ramp up to 200 users
    { duration: TEST_DURATION, target: 300 },    // Ramp up to 300 users
    { duration: TEST_DURATION, target: 400 },    // Peak load
    { duration: TEST_DURATION, target: 200 },    // Ramp down
    { duration: TEST_DURATION, target: 100 },    // Ramp down
    { duration: TEST_DURATION, target: 0 },      // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1000ms
    errors: ['rate<0.05'],             // Error rate must be less than 5%
  },
};

const BASE_URL = PRODUCT_URL;

export function setup() {
  // Create some test products first
  const sampleProducts = [
    {
      name: 'Test Product 1',
      description: 'Test product for load testing',
      price: 29.99,
      category: 'Electronics',
      stock: 100
    },
    {
      name: 'Test Product 2',
      description: 'Another test product',
      price: 49.99,
      category: 'Books',
      stock: 50
    }
  ];

  const createdProducts = [];
  
  for (const product of sampleProducts) {
    const res = http.post(`${BASE_URL}/product/`, JSON.stringify(product), {
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

  return { products: createdProducts };
}

export default function (data) {
  const scenarios = [
    () => testGetAllProducts(),
    () => testGetProductsByName(),
    () => testGetProductsByPage(),
    () => testGetProductsByCategory(),
    () => testGetProductsCount(),
    () => testCreateProduct(),
    () => testHealthCheck(),
  ];

  // Randomly select a scenario to execute
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();

  sleep(0.5);
}

function testGetAllProducts() {
  const res = http.get(`${BASE_URL}/product/`);

  const success = check(res, {
    'get all products status is 200': (r) => r.status === 200,
    'get all products response time < 1000ms': (r) => r.timings.duration < 1000,
    'get all products returns array': (r) => {
      try {
        const data = JSON.parse(r.body);
        return Array.isArray(data) || Array.isArray(data.products);
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
}

function testGetProductsByName() {
  const productNames = ['Test Product 1', 'Laptop', 'Phone', 'Book'];
  const randomName = productNames[Math.floor(Math.random() * productNames.length)];
  
  const res = http.get(`${BASE_URL}/product/getProductsByName/${encodeURIComponent(randomName)}`);

  const success = check(res, {
    'get products by name response received': (r) => r.status !== 0,
    'get products by name response time < 800ms': (r) => r.timings.duration < 800,
  });

  errorRate.add(!success);
}

function testGetProductsByPage() {
  const page = Math.floor(Math.random() * 5) + 1;
  const limit = Math.floor(Math.random() * 20) + 10;
  
  const res = http.get(`${BASE_URL}/product/getProductsByPage?page=${page}&limit=${limit}`);

  const success = check(res, {
    'get products by page response received': (r) => r.status !== 0,
    'get products by page response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testGetProductsByCategory() {
  const categories = ['Electronics', 'Books', 'Clothing', 'Home'];
  const randomCategory = categories[Math.floor(Math.random() * categories.length)];
  const page = Math.floor(Math.random() * 3) + 1;
  const limit = Math.floor(Math.random() * 15) + 5;
  
  const res = http.get(`${BASE_URL}/product/getProductsByPageAndCategory?page=${page}&limit=${limit}&category=${encodeURIComponent(randomCategory)}`);

  const success = check(res, {
    'get products by category response received': (r) => r.status !== 0,
    'get products by category response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testGetProductsCount() {
  const res = http.get(`${BASE_URL}/product/getProductsCount`);

  const success = check(res, {
    'get products count response received': (r) => r.status !== 0,
    'get products count response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}

function testCreateProduct() {
  const product = {
    name: `Load Test Product ${Math.random().toString(36).substring(7)}`,
    description: 'Product created during load testing',
    price: Math.round((Math.random() * 100 + 10) * 100) / 100,
    category: ['Electronics', 'Books', 'Clothing', 'Home'][Math.floor(Math.random() * 4)],
    stock: Math.floor(Math.random() * 100) + 1
  };

  const res = http.post(`${BASE_URL}/product/`, JSON.stringify(product), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'create product response received': (r) => r.status !== 0,
    'create product response time < 1500ms': (r) => r.timings.duration < 1500,
  });

  errorRate.add(!success);
}

function testHealthCheck() {
  const res = http.get(`${BASE_URL}/health`);

  const success = check(res, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 200ms': (r) => r.timings.duration < 200,
  });

  errorRate.add(!success);
}