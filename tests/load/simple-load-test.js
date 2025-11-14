import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const ORDER_URL = __ENV.ORDER_URL || 'http://localhost:8081';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up
    { duration: '1m', target: 50 },   // Stay at 50 VUs
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.01'],
    'errors': ['rate<0.05'],
  },
};

export default function () {
  const healthRes = http.get(`${ORDER_URL}/health`);
  check(healthRes, {
    'health is 200': (r) => r.status === 200,
  }) || errorRate.add(1);
  
  sleep(Math.random() + 0.5); // 0.5-1.5s
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, opts = {}) {
  const indent = opts.indent || '';
  const enableColors = opts.enableColors !== false;
  
  const colorize = (str, color) => {
    if (!enableColors) return str;
    const colors = {
      green: '\x1b[32m',
      red: '\x1b[31m',
      yellow: '\x1b[33m',
      cyan: '\x1b[36m',
      reset: '\x1b[0m',
    };
    return `${colors[color] || ''}${str}${colors.reset}`;
  };
  
  let summary = '\n' + colorize('████ LOAD TEST RESULTS ████\n', 'cyan');
  
  // HTTP metrics
  const httpMetrics = data.metrics.http_reqs;
  const duration = data.metrics.http_req_duration;
  const failed = data.metrics.http_req_failed;
  
  summary += `\n${indent}${colorize('HTTP Requests:', 'yellow')}\n`;
  summary += `${indent}  Total: ${httpMetrics.values.count}\n`;
  summary += `${indent}  Rate: ${httpMetrics.values.rate.toFixed(2)}/s\n`;
  summary += `${indent}  Failed: ${(failed.values.rate * 100).toFixed(2)}%\n`;
  
  summary += `\n${indent}${colorize('Response Times:', 'yellow')}\n`;
  summary += `${indent}  Avg: ${duration.values.avg.toFixed(2)}ms\n`;
  summary += `${indent}  P95: ${duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `${indent}  P99: ${duration.values['p(99)'].toFixed(2)}ms\n`;
  summary += `${indent}  Max: ${duration.values.max.toFixed(2)}ms\n`;
  
  // Checks
  const checks = data.metrics.checks;
  const checksPassed = checks.values.passes / checks.values.count * 100;
  summary += `\n${indent}${colorize('Checks:', 'yellow')}\n`;
  summary += `${indent}  Passed: ${colorize(`${checksPassed.toFixed(2)}%`, checksPassed >= 95 ? 'green' : 'red')}\n`;
  
  // VUs
  summary += `\n${indent}${colorize('Virtual Users:', 'yellow')}\n`;
  summary += `${indent}  Max: ${data.metrics.vus_max.values.max}\n`;
  
  return summary;
}
