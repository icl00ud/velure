import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '15s', target: 15 },   // Ramp up to 15 users
    { duration: '15s', target: 40 },   // Ramp up to 40 users
    { duration: '15s', target: 80 },   // Ramp up to 80 users
    { duration: '15s', target: 120 },  // Ramp up to 120 users
    { duration: '15s', target: 180 },  // Ramp up to 180 users
    { duration: '15s', target: 250 },  // Peak load
    { duration: '15s', target: 150 },  // Ramp down
    { duration: '15s', target: 75 },   // Ramp down
    { duration: '15s', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95% of requests must complete below 3000ms
    errors: ['rate<0.15'],             // Error rate must be less than 15%
  },
};

const BASE_URL = 'http://localhost:80'; // Adjust port based on your UI service config

export default function () {
  const scenarios = [
    () => testHomePage(),
    () => testStaticAssets(),
    () => testNavigationPages(),
    () => testAPIRoutes(),
  ];

  // Randomly select a scenario to execute
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();

  sleep(Math.random() * 2 + 1); // Random sleep between 1-3 seconds
}

function testHomePage() {
  const res = http.get(`${BASE_URL}/`);

  const success = check(res, {
    'home page status is 200': (r) => r.status === 200,
    'home page response time < 2000ms': (r) => r.timings.duration < 2000,
    'home page contains HTML': (r) => r.body.includes('<html') || r.body.includes('<!DOCTYPE'),
  });

  errorRate.add(!success);
}

function testStaticAssets() {
  const assets = [
    '/favicon.ico',
    '/robots.txt',
    '/assets/index.css',
    '/assets/index.js',
    '/placeholder.svg'
  ];

  const asset = assets[Math.floor(Math.random() * assets.length)];
  const res = http.get(`${BASE_URL}${asset}`);

  const success = check(res, {
    'static asset response received': (r) => r.status !== 0,
    'static asset response time < 1000ms': (r) => r.timings.duration < 1000,
    'static asset status ok': (r) => r.status === 200 || r.status === 404, // 404 is acceptable for some assets
  });

  errorRate.add(!success);
}

function testNavigationPages() {
  const pages = [
    '/products',
    '/orders',
    '/profile',
    '/login',
    '/register',
    '/about',
    '/contact'
  ];

  const page = pages[Math.floor(Math.random() * pages.length)];
  const res = http.get(`${BASE_URL}${page}`);

  const success = check(res, {
    'navigation page response received': (r) => r.status !== 0,
    'navigation page response time < 2500ms': (r) => r.timings.duration < 2500,
    'navigation page returns content': (r) => r.body && r.body.length > 0,
  });

  errorRate.add(!success);
}

function testAPIRoutes() {
  // Test potential API routes that the UI might expose
  const apiRoutes = [
    '/api/health',
    '/api/status',
    '/api/config',
    '/health',
    '/status'
  ];

  const route = apiRoutes[Math.floor(Math.random() * apiRoutes.length)];
  const res = http.get(`${BASE_URL}${route}`);

  const success = check(res, {
    'api route response received': (r) => r.status !== 0,
    'api route response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

export function handleSummary(data) {
  return {
    'ui-service-summary.html': htmlReport(data),
    'ui-service-summary.json': JSON.stringify(data),
  };
}

function htmlReport(data) {
  return `
<!DOCTYPE html>
<html>
<head>
    <title>UI Service Load Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { margin: 10px 0; padding: 10px; background-color: #f5f5f5; }
        .error { background-color: #ffebee; }
        .success { background-color: #e8f5e8; }
    </style>
</head>
<body>
    <h1>UI Service Load Test Results</h1>
    <div class="metric">
        <h3>Test Duration</h3>
        <p>${Math.round(data.state.testRunDurationMs / 1000)}s</p>
    </div>
    <div class="metric">
        <h3>Total Requests</h3>
        <p>${data.metrics.http_reqs.values.count}</p>
    </div>
    <div class="metric">
        <h3>Request Rate</h3>
        <p>${Math.round(data.metrics.http_reqs.values.rate * 100) / 100} req/s</p>
    </div>
    <div class="metric ${data.metrics.http_req_duration.values.p95 < 3000 ? 'success' : 'error'}">
        <h3>95th Percentile Response Time</h3>
        <p>${Math.round(data.metrics.http_req_duration.values.p95 * 100) / 100}ms</p>
    </div>
    <div class="metric ${data.metrics.errors ? (data.metrics.errors.values.rate < 0.15 ? 'success' : 'error') : 'success'}">
        <h3>Error Rate</h3>
        <p>${data.metrics.errors ? Math.round(data.metrics.errors.values.rate * 10000) / 100 : 0}%</p>
    </div>
</body>
</html>
  `;
}