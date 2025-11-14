import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function () {
  // Test auth service
  const authRes = http.get('http://localhost:3020/health');
  check(authRes, { 'auth health ok': (r) => r.status === 200 });

  // Test product service
  const productRes = http.get('http://localhost:3010/health');
  check(productRes, { 'product health ok': (r) => r.status === 200 });

  // Test publish-order service
  const publishOrderRes = http.get('http://localhost:3030/health');
  check(publishOrderRes, { 'publish-order health ok': (r) => r.status === 200 });

  // Test process-order service
  const processOrderRes = http.get('http://localhost:8081/health');
  check(processOrderRes, { 'process-order health ok': (r) => r.status === 200 });

  // Test metrics endpoints
  const authMetricsRes = http.get('http://localhost:3020/metrics');
  check(authMetricsRes, { 'auth metrics available': (r) => r.status === 200 && r.body.includes('http_requests_total') });

  const productMetricsRes = http.get('http://localhost:3010/metrics');
  check(productMetricsRes, { 'product metrics available': (r) => r.status === 200 && r.body.includes('http_requests_total') });

  const publishOrderMetricsRes = http.get('http://localhost:3030/metrics');
  check(publishOrderMetricsRes, { 'publish-order metrics available': (r) => r.status === 200 && r.body.includes('http_requests_total') });

  // Test public endpoints (no auth required)
  const ordersListRes = http.get('http://localhost:3030/orders');
  check(ordersListRes, { 'orders list accessible': (r) => r.status === 200 });

  sleep(1);
}
