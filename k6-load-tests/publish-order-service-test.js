import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '15s', target: 10 },   // Ramp up to 10 users
    { duration: '15s', target: 100 },   // Ramp up to 100 users
    { duration: '15s', target: 1000 },   // Ramp up to 1000 users
    { duration: '15s', target: 10000 },  // Ramp up to 10000 users
    { duration: '15s', target: 15000 },    // Ramp up to 15000 users
    { duration: '15s', target: 20000 },    // Peak load
    { duration: '15s', target: 10000 },    // Ramp down
    { duration: '15s', target: 5000 },     // Ramp down
    { duration: '15s', target: 0 },      // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2000ms
    errors: ['rate<0.1'],              // Error rate must be less than 10%
  },
};

const BASE_URL = 'http://localhost:3030';
let orderIds = [];

export function setup() {
  const sampleOrders = [
    [
      { product_id: "68c8522460d8eb66fa2b925c", name: "Product 1", quantity: 2, price: 29.99 },
      { product_id: "68c8522460d8eb66fa2b925d", name: "Product 2", quantity: 1, price: 49.99 }
    ],
    [
      { product_id: "68c8522460d8eb66fa2b925e", name: "Product 3", quantity: 1, price: 99.99 }
    ]
  ];

  const createdOrderIds = [];
  
  for (const items of sampleOrders) {
    const res = http.post(`${BASE_URL}/create-order`, JSON.stringify(items), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (res.status === 201 || res.status === 200) {
      try {
        const orderData = JSON.parse(res.body);
        if (orderData.order_id) {
          createdOrderIds.push(orderData.order_id);
        }
      } catch (e) {
      }
    }
  }

  return { orderIds: createdOrderIds };
}

export default function (data) {
  const scenarios = [
    () => testCreateOrder(),
    () => testUpdateOrderStatus(data.orderIds),
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
  });

  errorRate.add(!success);

  // Store order ID for potential status updates
  if (res.status === 200 || res.status === 201) {
    try {
      const orderData = JSON.parse(res.body);
      if (orderData.order_id) {
        orderIds.push(orderData.order_id);
        // Keep only last 50 order IDs to prevent memory issues
        if (orderIds.length > 50) {
          orderIds = orderIds.slice(-50);
        }
      }
    } catch (e) {
      // Continue if parsing fails
    }
  }
}

function testUpdateOrderStatus(setupOrderIds) {
  const allOrderIds = [...(setupOrderIds || []), ...orderIds];
  
  if (allOrderIds.length === 0) {
    // Skip if no orders available
    return;
  }

  const orderId = allOrderIds[Math.floor(Math.random() * allOrderIds.length)];
  const statuses = ['processing', 'shipped', 'delivered', 'cancelled'];
  const randomStatus = statuses[Math.floor(Math.random() * statuses.length)];

  const statusUpdate = {
    order_id: orderId,
    status: randomStatus
  };

  const res = http.post(`${BASE_URL}/update-order-status`, JSON.stringify(statusUpdate), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'update order status response received': (r) => r.status !== 0,
    'update order status response time < 1500ms': (r) => r.timings.duration < 1500,
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

function generateRandomAddress() {
  const cities = ['New York', 'Los Angeles', 'Chicago', 'Houston', 'Phoenix'];
  const states = ['NY', 'CA', 'IL', 'TX', 'AZ'];
  const streets = ['Main St', 'Oak Ave', 'Park Blvd', 'First St', 'Second Ave'];
  
  return {
    street: `${Math.floor(Math.random() * 9999) + 1} ${streets[Math.floor(Math.random() * streets.length)]}`,
    city: cities[Math.floor(Math.random() * cities.length)],
    state: states[Math.floor(Math.random() * states.length)],
    zip: String(Math.floor(Math.random() * 90000) + 10000),
    country: 'US'
  };
}