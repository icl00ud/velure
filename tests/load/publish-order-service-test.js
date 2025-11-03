import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Environment configuration
// Can be set via: k6 run -e ORDER_URL=https://velure.local/api/order publish-order-service-test.js
const ORDER_URL = __ENV.ORDER_URL || __ENV.BASE_URL || 'http://localhost:3030';
const WARMUP_DURATION = __ENV.WARMUP_DURATION || '30s';
const TEST_DURATION = __ENV.TEST_DURATION || '5s';

console.log(`🎯 Target URL: ${ORDER_URL}`);
console.log(`⏱️  Warmup: ${WARMUP_DURATION}, Test stages: ${TEST_DURATION} each`);

export const options = {
  stages: [
    { duration: WARMUP_DURATION, target: 5 },    // Warmup phase
    { duration: TEST_DURATION, target: 10 },     // Ramp up to 10 users
    { duration: TEST_DURATION, target: 50 },     // Ramp up to 50 users
    { duration: TEST_DURATION, target: 100 },    // Ramp up to 100 users
    { duration: TEST_DURATION, target: 200 },    // Ramp up to 200 users
    { duration: TEST_DURATION, target: 500 },    // Ramp up to 500 users
    { duration: TEST_DURATION, target: 1000 },   // Peak load
    { duration: TEST_DURATION, target: 500 },    // Ramp down
    { duration: TEST_DURATION, target: 100 },    // Ramp down
    { duration: TEST_DURATION, target: 0 },      // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2000ms
    errors: ['rate<0.1'],              // Error rate must be less than 10%
  },
};

const BASE_URL = ORDER_URL;

export default function () {
  const scenarios = [
    () => testCreateOrder(),
    () => testGetOrders(),
  ];

  const weights = [0.7, 0.3];
  const random = Math.random();
  let selectedScenario = 0;
  
  let cumulativeWeight = 0;
  for (let i = 0; i < weights.length; i++) {
    cumulativeWeight += weights[i];
    if (random <= cumulativeWeight) {
      selectedScenario = i;
      break;
    }
  }

  scenarios[selectedScenario]();
  sleep(1);
}

function testCreateOrder() {
  const productIds = [
    '68c8522460d8eb66fa2b925c',
    '68c8522460d8eb66fa2b925d',
    '68c8522460d8eb66fa2b925e',
    '68c8522460d8eb66fa2b925f',
    '68c8522460d8eb66fa2b9260'
  ];
  
  const items = generateRandomItems(productIds);

  const res = http.post(`${BASE_URL}/create-order`, JSON.stringify(items), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'create order response received': (r) => r.status !== 0,
    'create order response time < 2000ms': (r) => r.timings.duration < 2000,
    'create order status is success': (r) => r.status === 200 || r.status === 201,
    'order has id': (r) => {
      try {
        const data = JSON.parse(r.body);
        return data.order_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!success);
}

function testGetOrders() {
  const page = Math.floor(Math.random() * 5) + 1; // Pages 1-5
  const pageSize = 10;

  const res = http.get(`${BASE_URL}/orders?page=${page}&page_size=${pageSize}`, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'get orders response received': (r) => r.status !== 0,
    'get orders response time < 1500ms': (r) => r.timings.duration < 1500,
    'get orders status is success': (r) => r.status === 200,
    'orders have data': (r) => {
      try {
        const data = JSON.parse(r.body);
        return data.orders !== undefined && Array.isArray(data.orders);
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!success);
}

function generateRandomItems(productIds) {
  const numItems = Math.floor(Math.random() * 3) + 1; // 1-3 items
  const items = [];
  const productNames = ['Product A', 'Product B', 'Product C', 'Product D', 'Product E'];
  
  for (let i = 0; i < numItems; i++) {
    items.push({
      product_id: productIds[Math.floor(Math.random() * productIds.length)],
      name: productNames[Math.floor(Math.random() * productNames.length)],
      quantity: Math.floor(Math.random() * 5) + 1,
      price: Math.round((Math.random() * 100 + 10) * 100) / 100
    });
  }
  
  return items;
}