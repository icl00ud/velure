import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const loginSuccessRate = new Rate('login_success_rate');
const registerSuccessRate = new Rate('register_success_rate');
const loginDuration = new Trend('login_duration');
const registerDuration = new Trend('register_duration');
const tokenValidationDuration = new Trend('token_validation_duration');

// Test configuration
export const options = {
  scenarios: {
    // Cenário 1: Teste de registro com rampa
    registration_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },  // Rampa até 50 VUs
        { duration: '1m', target: 100 },  // Rampa até 100 VUs
        { duration: '2m', target: 100 },  // Mantém 100 VUs
        { duration: '30s', target: 0 },   // Rampa down
      ],
      gracefulRampDown: '30s',
      exec: 'testRegistration',
    },
    
    // Cenário 2: Teste de login com alta carga
    login_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 100 }, // Rampa até 100 VUs
        { duration: '1m', target: 200 },  // Rampa até 200 VUs
        { duration: '2m', target: 300 },  // Rampa até 300 VUs (pico)
        { duration: '1m', target: 100 },  // Rampa down
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '30s',
      startTime: '30s', // Começa após registration
      exec: 'testLogin',
    },
    
    // Cenário 3: Teste de validação de tokens
    token_validation: {
      executor: 'constant-vus',
      vus: 50,
      duration: '5m',
      startTime: '1m',
      exec: 'testTokenValidation',
    },
    
    // Cenário 4: Spike test
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 500 },  // Spike súbito
        { duration: '30s', target: 500 },  // Mantém pico
        { duration: '10s', target: 0 },    // Queda súbita
      ],
      startTime: '6m',
      exec: 'testLogin',
    },
  },
  
  thresholds: {
    // Thresholds para registro
    'http_req_duration{scenario:registration_load}': ['p(95)<500', 'p(99)<1000'],
    'register_success_rate': ['rate>0.95'],
    
    // Thresholds para login
    'http_req_duration{scenario:login_load}': ['p(95)<300', 'p(99)<500'],
    'login_success_rate': ['rate>0.98'],
    
    // Thresholds para validação
    'http_req_duration{scenario:token_validation}': ['p(95)<100', 'p(99)<200'],
    
    // Thresholds gerais
    'http_req_failed': ['rate<0.05'], // 95% success rate
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3020';

// Estado compartilhado entre VUs
let userEmail = '';
let userPassword = 'TestPassword123!';
let accessToken = '';

// Cenário 1: Test Registration
export function testRegistration() {
  const randomId = Math.floor(Math.random() * 1000000);
  const payload = JSON.stringify({
    name: `Test User ${randomId}`,
    email: `testuser${randomId}@example.com`,
    password: userPassword,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Registration' },
  };

  const startTime = Date.now();
  const response = http.post(`${BASE_URL}/authentication/register`, payload, params);
  const duration = Date.now() - startTime;

  registerDuration.add(duration);

  const success = check(response, {
    'registration status is 201': (r) => r.status === 201,
    'registration has access token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.accessToken !== undefined;
      } catch {
        return false;
      }
    },
    'registration response time < 500ms': () => duration < 500,
  });

  registerSuccessRate.add(success);

  if (success && response.status === 201) {
    const body = JSON.parse(response.body);
    userEmail = body.email;
    accessToken = body.accessToken;
  }

  sleep(0.5);
}

// Cenário 2: Test Login
export function testLogin() {
  // Primeiro, registrar um usuário se não existir
  if (!userEmail) {
    testRegistration();
    sleep(0.5);
  }

  const payload = JSON.stringify({
    email: userEmail,
    password: userPassword,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Login' },
  };

  const startTime = Date.now();
  const response = http.post(`${BASE_URL}/authentication/login`, payload, params);
  const duration = Date.now() - startTime;

  loginDuration.add(duration);

  const success = check(response, {
    'login status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'login has access token': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.accessToken !== undefined;
      } catch {
        return false;
      }
    },
    'login response time < 300ms': () => duration < 300,
  });

  loginSuccessRate.add(success);

  if (success && response.body) {
    try {
      const body = JSON.parse(response.body);
      accessToken = body.accessToken;
    } catch {
      // Ignore parse errors
    }
  }

  sleep(0.3);
}

// Cenário 3: Test Token Validation
export function testTokenValidation() {
  // Criar usuário e fazer login se não tiver token
  if (!accessToken) {
    testRegistration();
    sleep(0.5);
  }

  const payload = JSON.stringify({
    token: accessToken,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'TokenValidation' },
  };

  const startTime = Date.now();
  const response = http.post(`${BASE_URL}/authentication/validateToken`, payload, params);
  const duration = Date.now() - startTime;

  tokenValidationDuration.add(duration);

  check(response, {
    'validation status is 200': (r) => r.status === 200,
    'validation response is valid': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.isValid === true;
      } catch {
        return false;
      }
    },
    'validation response time < 100ms': () => duration < 100,
  });

  sleep(0.2);
}

// Setup: executado uma vez no início
export function setup() {
  console.log('🚀 Starting performance tests...');
  console.log(`Base URL: ${BASE_URL}`);
  
  // Health check
  const healthResponse = http.get(`${BASE_URL}/health`);
  if (healthResponse.status !== 200) {
    throw new Error('Service is not healthy!');
  }
  
  console.log('✅ Service is healthy');
  return { baseUrl: BASE_URL };
}

// Teardown: executado uma vez no final
export function teardown(data) {
  console.log('✅ Performance tests completed!');
}

// Handler de erros personalizados
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'performance-report.json': JSON.stringify(data),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  const colors = options.enableColors || false;
  
  let output = '\n';
  output += '========================================\n';
  output += '   PERFORMANCE TEST SUMMARY\n';
  output += '========================================\n\n';
  
  // Scenarios
  output += `${indent}Scenarios:\n`;
  for (const [name, scenario] of Object.entries(data.metrics)) {
    if (scenario.type === 'counter' && name.includes('_success_rate')) {
      const rate = (scenario.values.rate * 100).toFixed(2);
      output += `${indent}  ${name}: ${rate}%\n`;
    }
  }
  
  output += '\n';
  
  // HTTP metrics
  output += `${indent}HTTP Metrics:\n`;
  if (data.metrics.http_req_duration) {
    const dur = data.metrics.http_req_duration.values;
    output += `${indent}  Response Time (p95): ${dur['p(95)'].toFixed(2)}ms\n`;
    output += `${indent}  Response Time (p99): ${dur['p(99)'].toFixed(2)}ms\n`;
    output += `${indent}  Response Time (avg): ${dur.avg.toFixed(2)}ms\n`;
  }
  
  if (data.metrics.http_reqs) {
    const reqs = data.metrics.http_reqs.values;
    output += `${indent}  Total Requests: ${reqs.count}\n`;
    output += `${indent}  Requests/sec: ${reqs.rate.toFixed(2)}\n`;
  }
  
  output += '\n========================================\n';
  
  return output;
}
