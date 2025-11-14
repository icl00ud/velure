import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// Custom metrics
const authLoginRate = new Rate('auth_login_success_rate');
const authRegisterRate = new Rate('auth_register_success_rate');
const productQueryRate = new Rate('product_query_success_rate');
const orderCreationRate = new Rate('order_creation_success_rate');
const orderProcessingRate = new Rate('order_processing_success_rate');

const authLatency = new Trend('auth_latency_ms');
const productLatency = new Trend('product_latency_ms');
const orderLatency = new Trend('order_latency_ms');

const activeUsers = new Gauge('active_users');
const ordersCreated = new Counter('orders_created_total');
const ordersFailed = new Counter('orders_failed_total');

// Service endpoints
const AUTH_URL = __ENV.AUTH_URL || 'http://localhost:3020';
const PRODUCT_URL = __ENV.PRODUCT_URL || 'http://localhost:3010';
const PUBLISH_ORDER_URL = __ENV.PUBLISH_ORDER_URL || 'http://localhost:3030';

// Load test configuration
export const options = {
  stages: [
    { duration: '1m', target: 10 },   // Ramp up to 10 users
    { duration: '3m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 50 },   // Ramp down to 50 users
    { duration: '1m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.05'],
    'auth_login_success_rate': ['rate>0.95'],
    'product_query_success_rate': ['rate>0.98'],
    'order_creation_success_rate': ['rate>0.90'],
    'order_processing_success_rate': ['rate>0.85'],
  },
};

// Generate unique test data
const generateEmail = () => `user_${Date.now()}_${Math.random().toString(36).substring(7)}@loadtest.com`;
const generatePassword = () => Math.random().toString(36).substring(2, 15);

// Shared products data (simulated - will be fetched dynamically)
let productCache = [];
let authTokens = new Map();

export function setup() {
  console.log('🚀 Starting full system load test...');
  console.log(`📍 Auth Service: ${AUTH_URL}`);
  console.log(`📍 Product Service: ${PRODUCT_URL}`);
  console.log(`📍 Publish Order Service: ${PUBLISH_ORDER_URL}`);
  
  // Health check
  const healthChecks = [
    http.get(`${AUTH_URL}/health`),
    http.get(`${PRODUCT_URL}/health`),
    http.get(`${PUBLISH_ORDER_URL}/health`),
  ];
  
  const allHealthy = healthChecks.every(r => r.status === 200);
  if (!allHealthy) {
    console.error('❌ Health check failed!');
    throw new Error('Services not healthy');
  }
  
  console.log('✅ All services healthy');
  
  // Fetch initial product data
  const productsRes = http.get(`${PRODUCT_URL}/products?page=1&limit=50`);
  if (productsRes.status === 200) {
    const data = JSON.parse(productsRes.body);
    productCache = data.products || [];
    console.log(`✅ Loaded ${productCache.length} products`);
  }
  
  return { products: productCache };
}

export default function(data) {
  const userId = `user_${__VU}_${__ITER}`;
  activeUsers.add(1);
  
  // Simulate realistic user journey
  const scenario = Math.random();
  
  if (scenario < 0.3) {
    // 30% - Browse products only
    browseProductsScenario(data);
  } else if (scenario < 0.6) {
    // 30% - Register new user
    registerAndBrowseScenario(data);
  } else {
    // 40% - Full purchase flow
    fullPurchaseScenario(data);
  }
  
  sleep(Math.random() * 3 + 1); // 1-4 seconds between iterations
  activeUsers.add(-1);
}

function browseProductsScenario(data) {
  group('Browse Products', () => {
    // List all products
    let res = http.get(`${PRODUCT_URL}/products?page=1&limit=20`);
    const success = check(res, {
      'products listed': (r) => r.status === 200,
    });
    productQueryRate.add(success);
    productLatency.add(res.timings.duration);
    
    if (res.status === 200) {
      const products = JSON.parse(res.body).products || [];
      
      // View random product details
      if (products.length > 0) {
        const product = products[Math.floor(Math.random() * products.length)];
        res = http.get(`${PRODUCT_URL}/products?name=${encodeURIComponent(product.name)}`);
        check(res, { 'product details loaded': (r) => r.status === 200 });
      }
    }
    
    sleep(0.5);
    
    // Search by category
    const categories = ['dogs', 'cats', 'birds', 'fish'];
    const category = categories[Math.floor(Math.random() * categories.length)];
    res = http.get(`${PRODUCT_URL}/products/category/${category}`);
    check(res, { 'category search': (r) => r.status === 200 });
    
    sleep(0.3);
  });
}

function registerAndBrowseScenario(data) {
  let token = null;
  
  group('Register New User', () => {
    const email = generateEmail();
    const password = generatePassword();
    
    const payload = JSON.stringify({
      email: email,
      password: password,
      name: `LoadTest User ${__VU}`,
    });
    
    const params = {
      headers: { 'Content-Type': 'application/json' },
    };
    
    const res = http.post(`${AUTH_URL}/auth/register`, payload, params);
    const success = check(res, {
      'registration successful': (r) => r.status === 201,
    });
    authRegisterRate.add(success);
    authLatency.add(res.timings.duration);
    
    if (success) {
      const body = JSON.parse(res.body);
      token = body.token;
      authTokens.set(__VU, token);
    }
    
    sleep(0.5);
  });
  
  if (token) {
    group('Browse Products (Authenticated)', () => {
      const params = {
        headers: { 'Authorization': `Bearer ${token}` },
      };
      
      const res = http.get(`${PRODUCT_URL}/products?page=1&limit=20`, params);
      check(res, { 'authenticated browse': (r) => r.status === 200 });
      
      sleep(0.5);
    });
  }
}

function fullPurchaseScenario(data) {
  let token = null;
  
  // Login or register
  group('Authenticate', () => {
    // Always register new user for this scenario
    const email = generateEmail();
    const password = generatePassword();
      
    const registerPayload = JSON.stringify({
      email: email,
      password: password,
      name: `LoadTest User ${__VU}`,
    });
    
    const params = {
      headers: { 'Content-Type': 'application/json' },
    };
    
    let res = http.post(`${AUTH_URL}/authentication/register`, registerPayload, params);
    const registerSuccess = check(res, {
      'registration ok': (r) => r.status === 201,
    });
    authRegisterRate.add(registerSuccess);
    authLatency.add(res.timings.duration);
    
    if (registerSuccess) {
      const body = JSON.parse(res.body);
      token = body.token;
      authTokens.set(__VU, token);
    }
    
    sleep(0.3);
  });
  
  if (!token) {
    console.log('❌ No token available, skipping purchase');
    return;
  }
  
  // Browse and select products
  let selectedProducts = [];
  group('Select Products', () => {
    const params = {
      headers: { 'Authorization': `Bearer ${token}` },
    };
    
    const res = http.get(`${PRODUCT_URL}/products?page=1&limit=50`, params);
    const success = check(res, {
      'products loaded': (r) => r.status === 200,
    });
    productQueryRate.add(success);
    productLatency.add(res.timings.duration);
    
    if (res.status === 200) {
      const products = JSON.parse(res.body).products || [];
      
      // Select 1-5 random products
      const numItems = Math.floor(Math.random() * 4) + 1;
      for (let i = 0; i < numItems && products.length > 0; i++) {
        const product = products[Math.floor(Math.random() * products.length)];
        selectedProducts.push({
          product_id: product.id,
          quantity: Math.floor(Math.random() * 3) + 1,
        });
      }
    }
    
    sleep(0.5);
  });
  
  if (selectedProducts.length === 0) {
    console.log('⚠️  No products selected, skipping order creation');
    return;
  }
  
  // Create order
  group('Create Order', () => {
    const orderPayload = JSON.stringify({
      items: selectedProducts,
      customer_email: `customer_${__VU}@loadtest.com`,
    });
    
    const params = {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    };
    
    const res = http.post(`${PUBLISH_ORDER_URL}/orders`, orderPayload, params);
    const success = check(res, {
      'order created': (r) => r.status === 201 || r.status === 200,
      'order has id': (r) => {
        if (r.status === 201 || r.status === 200) {
          const body = JSON.parse(r.body);
          return body.order_id !== undefined && body.order_id !== '';
        }
        return false;
      },
    });
    
    orderCreationRate.add(success);
    orderLatency.add(res.timings.duration);
    
    if (success) {
      ordersCreated.add(1);
      const body = JSON.parse(res.body);
      console.log(`✅ Order created: ${body.order_id}`);
      
      // Check order status after a delay
      sleep(2);
      
      const statusRes = http.get(`${PUBLISH_ORDER_URL}/orders/${body.order_id}`, params);
      const statusCheck = check(statusRes, {
        'order status retrieved': (r) => r.status === 200,
      });
      
      if (statusCheck) {
        const orderStatus = JSON.parse(statusRes.body);
        console.log(`📦 Order ${body.order_id} status: ${orderStatus.status}`);
        
        // Track processing success
        if (orderStatus.status === 'completed') {
          orderProcessingRate.add(true);
        } else if (orderStatus.status === 'failed') {
          orderProcessingRate.add(false);
          ordersFailed.add(1);
        }
      }
    } else {
      orderCreationRate.add(false);
      ordersFailed.add(1);
      console.log(`❌ Order creation failed: ${res.status} - ${res.body}`);
    }
    
    sleep(0.5);
  });
}

export function teardown(data) {
  console.log('🏁 Load test completed');
  console.log(`📊 Total products in cache: ${data.products.length}`);
}

export function handleSummary(data) {
  console.log('🏁 Load test completed');
  console.log(`📊 Checks: ${data.metrics.checks ? (data.metrics.checks.values.passes / data.metrics.checks.values.count * 100).toFixed(2) : 0}%`);
  console.log(`📊 HTTP Requests: ${data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0}`);
  console.log(`📊 HTTP Request Duration (avg): ${data.metrics.http_req_duration ? data.metrics.http_req_duration.values.avg.toFixed(2) : 0}ms`);
  
  return {
    'summary.json': JSON.stringify(data),
  };
}
