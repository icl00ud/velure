// Velure - Complete User Journey Load Test
// Tests the complete user flow: Register -> Login -> Browse Products -> Purchase

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// ============================================================================
// CONFIGURATION
// ============================================================================

const BASE_URL = __ENV.BASE_URL || 'https://velure.local';
const AUTH_URL = `${BASE_URL}/api/auth`;
const PRODUCT_URL = `${BASE_URL}/api/product`;
const ORDER_URL = `${BASE_URL}/api/order`;

const WARMUP_DURATION = __ENV.WARMUP_DURATION || '10s';
const TEST_DURATION = __ENV.TEST_DURATION || '1m';
const COOLDOWN_DURATION = __ENV.COOLDOWN_DURATION || '10s';
const TARGET_VUS = __ENV.TARGET_VUS || 20;

// ============================================================================
// CUSTOM METRICS
// ============================================================================

const journeySuccess = new Rate('journey_success_rate');
const journeyDuration = new Trend('journey_duration');
const stepsCompleted = new Counter('journey_steps_completed');

// ============================================================================
// TEST OPTIONS
// ============================================================================

export const options = {
  scenarios: {
    complete_user_journey: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: WARMUP_DURATION, target: parseInt(TARGET_VUS) / 2 },  // Warmup
        { duration: TEST_DURATION, target: parseInt(TARGET_VUS) },         // Test load
        { duration: COOLDOWN_DURATION, target: 0 },                        // Cool down
      ],
      gracefulRampDown: '5s',
    },
  },

  thresholds: {
    // Journey-level thresholds
    'journey_success_rate': ['rate>0.95'],        // 95% of journeys should succeed
    'journey_duration': ['p(95)<10200'],          // 95% of journeys should complete in 10.2s

    // HTTP-level thresholds
    'http_req_duration': ['p(95)<2500'],          // 95% of requests under 2.5s
    'http_req_failed': ['rate<0.05'],             // Less than 5% errors

    // Individual steps
    'checks': ['rate>0.9'],                       // 90% of checks should pass
  },

  // Disable SSL verification for local development
  insecureSkipTLSVerify: true,
};

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

function generateUsername() {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 10000);
  return `user_${timestamp}_${random}`;
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

function getRandomProductIds(availableProductIds, count = 2) {
  if (!availableProductIds || availableProductIds.length === 0) {
    // Fallback to placeholder IDs if no products available
    const productIds = [
      '67a2c1e5564dfbe318544ca7',
      '67a2c1e5564dfbe318544ca8',
      '67a2c1e5564dfbe318544ca9',
      '67a2c1e5564dfbe318544caa',
    ];
    const selected = [];
    for (let i = 0; i < Math.min(count, productIds.length); i++) {
      selected.push(productIds[Math.floor(Math.random() * productIds.length)]);
    }
    return selected;
  }

  const selected = [];
  for (let i = 0; i < Math.min(count, availableProductIds.length); i++) {
    selected.push(availableProductIds[Math.floor(Math.random() * availableProductIds.length)]);
  }
  return selected;
}

// ============================================================================
// MAIN TEST SCENARIO
// ============================================================================

export default function(data) {
  const journeyStart = Date.now();
  let journeyFailed = false;
  let authToken = null;
  const availableProductIds = data && data.productIds ? data.productIds : [];

  const username = generateUsername();
  const email = `${username}@velure.test`;
  const password = 'TestPassword123!';

  group('Complete User Journey', function() {

    // =======================================================================
    // STEP 1: User Registration
    // =======================================================================
    group('1. Register New Account', function() {
      const registerPayload = JSON.stringify({
        name: username,
        email: email,
        password: password,
      });

      const registerRes = http.post(
        `${AUTH_URL}/register`,
        registerPayload,
        {
          headers: getHeaders(),
          tags: { step: 'register' }
        }
      );

      const registerOk = check(registerRes, {
        'registration: status is 201': (r) => r.status === 201,
        'registration: returns token': (r) => {
          try {
            return r.json('accessToken') !== undefined;
          } catch (e) {
            return false;
          }
        },
      });

      if (registerOk && registerRes.status === 201) {
        try {
          authToken = registerRes.json('accessToken');
          stepsCompleted.add(1);
        } catch (e) {
          console.error('Failed to parse registration response');
          journeyFailed = true;
        }
      } else {
        const errorReason = registerRes.status === 201 ? 'missing accessToken in response' : `HTTP ${registerRes.status}`;
        console.error(`Registration failed: ${errorReason}`);
        journeyFailed = true;
      }
    });

    if (journeyFailed) {
      journeySuccess.add(false);
      return;
    }

    sleep(0.5);

    // =======================================================================
    // STEP 2: User Login
    // =======================================================================
    group('2. Login to Account', function() {
      const loginPayload = JSON.stringify({
        email: email,
        password: password,
      });

      const loginRes = http.post(
        `${AUTH_URL}/login`,
        loginPayload,
        {
          headers: getHeaders(),
          tags: { step: 'login' }
        }
      );

      const loginOk = check(loginRes, {
        'login: status is 200': (r) => r.status === 200,
        'login: returns token': (r) => {
          try {
            return r.json('accessToken') !== undefined;
          } catch (e) {
            return false;
          }
        },
      });

      if (loginOk && loginRes.status === 200) {
        try {
          authToken = loginRes.json('accessToken');
          stepsCompleted.add(1);
        } catch (e) {
          console.error('Failed to parse login response');
          journeyFailed = true;
        }
      } else {
        const errorReason = loginRes.status === 200 ? 'missing accessToken in response' : `HTTP ${loginRes.status}`;
        console.error(`Login failed: ${errorReason}`);
        journeyFailed = true;
      }
    });

    if (journeyFailed) {
      journeySuccess.add(false);
      return;
    }

    sleep(1);

    // =======================================================================
    // STEP 3: Browse Product Catalog
    // =======================================================================
    group('3. Browse Product Catalog', function() {
      // List all products
      const listRes = http.get(
        `${PRODUCT_URL}/products?page=1&limit=20`,
        {
          headers: getHeaders(authToken),
          tags: { step: 'browse_products' }
        }
      );

      const listOk = check(listRes, {
        'browse: status is 200': (r) => r.status === 200,
        'browse: returns products': (r) => {
          try {
            const data = r.json();
            return data.products !== undefined;
          } catch (e) {
            return false;
          }
        },
      });

      if (listOk) {
        stepsCompleted.add(1);
      } else {
        console.error(`Browse products failed: ${listRes.status} - ${listRes.body}`);
        journeyFailed = true;
      }
    });

    if (journeyFailed) {
      journeySuccess.add(false);
      return;
    }

    sleep(0.5);

    // =======================================================================
    // STEP 4: Search for Specific Products
    // =======================================================================
    group('4. Search Products', function() {
      const searchTerms = ['product', 'test', 'item'];
      const term = searchTerms[Math.floor(Math.random() * searchTerms.length)];

      const searchRes = http.get(
        `${PRODUCT_URL}/products/search?q=${term}`,
        {
          headers: getHeaders(authToken),
          tags: { step: 'search_products' }
        }
      );

      const searchOk = check(searchRes, {
        'search: status is 200': (r) => r.status === 200,
      });

      if (searchOk) {
        stepsCompleted.add(1);
      } else {
        console.error(`Search products failed: ${searchRes.status}`);
        journeyFailed = true;
      }
    });

    if (journeyFailed) {
      journeySuccess.add(false);
      return;
    }

    sleep(1);

    // =======================================================================
    // STEP 5: Create Purchase Order
    // =======================================================================
    group('5. Create Purchase Order', function() {
      const productIds = getRandomProductIds(availableProductIds, 2);

      const orderPayload = JSON.stringify({
        items: [
          {
            product_id: productIds[0],
            name: 'Selected Product 1',
            quantity: Math.floor(Math.random() * 3) + 1,
            price: parseFloat((Math.random() * 100 + 10).toFixed(2)),
          },
          {
            product_id: productIds[1],
            name: 'Selected Product 2',
            quantity: Math.floor(Math.random() * 2) + 1,
            price: parseFloat((Math.random() * 50 + 5).toFixed(2)),
          },
        ],
      });

      const orderRes = http.post(
        `${ORDER_URL}/create-order`,
        orderPayload,
        {
          headers: getHeaders(authToken),
          tags: { step: 'create_order' }
        }
      );

      const orderOk = check(orderRes, {
        'order: status is 201': (r) => r.status === 201,
        'order: returns order_id': (r) => {
          try {
            return r.json('order_id') !== undefined;
          } catch (e) {
            return false;
          }
        },
      });

      if (orderOk) {
        stepsCompleted.add(1);
      } else {
        console.error(`Create order failed: ${orderRes.status} - ${orderRes.body}`);
        journeyFailed = true;
      }
    });

    if (journeyFailed) {
      journeySuccess.add(false);
      return;
    }

    sleep(0.5);

    // =======================================================================
    // STEP 6: View Order History
    // =======================================================================
    group('6. View Order History', function() {
      const ordersRes = http.get(
        `${ORDER_URL}/orders?page=1&limit=10`,
        {
          headers: getHeaders(authToken),
          tags: { step: 'view_orders' }
        }
      );

      const ordersOk = check(ordersRes, {
        'orders: status is 200': (r) => r.status === 200,
      });

      if (ordersOk) {
        stepsCompleted.add(1);
      } else {
        console.error(`View orders failed: ${ordersRes.status}`);
        journeyFailed = true;
      }
    });

  });

  // Record journey metrics
  const journeyEnd = Date.now();
  const duration = journeyEnd - journeyStart;

  journeyDuration.add(duration);
  journeySuccess.add(!journeyFailed);

  // Small delay between iterations
  sleep(2);
}

// ============================================================================
// TEST LIFECYCLE
// ============================================================================

export function setup() {
  console.log('');
  console.log('🚀 Starting Velure Complete User Journey Load Test');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log(`📍 Base URL: ${BASE_URL}`);
  console.log(`👥 Target VUs: ${TARGET_VUS}`);
  console.log(`⏱️  Duration: ${WARMUP_DURATION} warmup + ${TEST_DURATION} test + ${COOLDOWN_DURATION} cooldown`);
  console.log('');
  console.log('📝 Journey Steps:');
  console.log('   1. Register New Account');
  console.log('   2. Login to Account');
  console.log('   3. Browse Product Catalog');
  console.log('   4. Search Products');
  console.log('   5. Create Purchase Order');
  console.log('   6. View Order History');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('');

  // Fetch available products to use in tests
  let productIds = [];
  try {
    const res = http.get(`${PRODUCT_URL}/products?page=1&limit=50`, {
      headers: { 'Content-Type': 'application/json' }
    });
    
    if (res.status === 200) {
      const body = res.json();
      if (body.products && body.products.length > 0) {
        productIds = body.products.map(p => p.id || p._id).filter(id => id);
        console.log(`✅ Fetched ${productIds.length} real product IDs for testing`);
      } else {
        console.warn('⚠️ No products found in response, using fallback IDs');
      }
    } else {
      console.warn(`⚠️ Failed to fetch products (Status ${res.status}), using fallback IDs`);
    }
  } catch (e) {
    console.error(`❌ Error fetching products: ${e.message}`);
  }

  return { productIds: productIds };
}

export function teardown(data) {
  console.log('');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('✅ User Journey Load Test Completed');
  console.log('');
  console.log('📊 View detailed metrics in Grafana:');
  console.log('   → http://localhost:3000');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('');
}
