import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '15s', target: 10 },   // Ramp up to 10 users
    { duration: '15s', target: 30 },   // Ramp up to 30 users
    { duration: '15s', target: 60 },   // Ramp up to 60 users
    { duration: '15s', target: 100 },  // Ramp up to 100 users
    { duration: '15s', target: 150 },  // Ramp up to 150 users
    { duration: '15s', target: 200 },  // Peak load
    { duration: '15s', target: 100 },  // Ramp down
    { duration: '15s', target: 50 },   // Ramp down
    { duration: '15s', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2000ms
    errors: ['rate<0.1'],              // Error rate must be less than 10%
  },
};

const BASE_URL = 'http://localhost:3030'; // Adjust port based on your config
let orderIds = [];

export function setup() {
  // Create some initial orders for status updates
  const sampleOrders = [
    {
      customer_id: "customer-1",
      items: [
        { product_id: "product-1", quantity: 2, price: 29.99 },
        { product_id: "product-2", quantity: 1, price: 49.99 }
      ],
      shipping_address: {
        street: "123 Test St",
        city: "Test City",
        state: "TS",
        zip: "12345",
        country: "US"
      }
    },
    {
      customer_id: "customer-2",
      items: [
        { product_id: "product-3", quantity: 1, price: 99.99 }
      ],
      shipping_address: {
        street: "456 Load Ave",
        city: "Load City",
        state: "LC",
        zip: "67890",
        country: "US"
      }
    }
  ];

  const createdOrderIds = [];
  
  for (const order of sampleOrders) {
    const res = http.post(`${BASE_URL}/create-order`, JSON.stringify(order), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (res.status === 201 || res.status === 200) {
      try {
        const orderData = JSON.parse(res.body);
        if (orderData.id || orderData.order_id) {
          createdOrderIds.push(orderData.id || orderData.order_id);
        }
      } catch (e) {
        // Continue if parsing fails
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

  // Weight scenarios - more creates than updates
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
  const customerIds = ['customer-1', 'customer-2', 'customer-3', 'customer-4', 'customer-5'];
  const productIds = ['product-1', 'product-2', 'product-3', 'product-4', 'product-5'];
  
  const order = {
    customer_id: customerIds[Math.floor(Math.random() * customerIds.length)],
    items: generateRandomItems(productIds),
    shipping_address: generateRandomAddress()
  };

  const res = http.post(`${BASE_URL}/create-order`, JSON.stringify(order), {
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
      if (orderData.id || orderData.order_id) {
        orderIds.push(orderData.id || orderData.order_id);
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
    // Create a dummy order ID for testing
    const dummyOrderId = `order-${Math.random().toString(36).substring(7)}`;
    allOrderIds.push(dummyOrderId);
  }

  const orderId = allOrderIds[Math.floor(Math.random() * allOrderIds.length)];
  const statuses = ['processing', 'shipped', 'delivered', 'cancelled'];
  const randomStatus = statuses[Math.floor(Math.random() * statuses.length)];

  const statusUpdate = {
    order_id: orderId,
    status: randomStatus,
    updated_by: 'load-test',
    notes: `Status updated to ${randomStatus} during load test`
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
  
  for (let i = 0; i < numItems; i++) {
    items.push({
      product_id: productIds[Math.floor(Math.random() * productIds.length)],
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