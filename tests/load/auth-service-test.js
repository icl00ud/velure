import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '15s', target: 10 },   // Ramp up to 10 users
    { duration: '15s', target: 25 },   // Ramp up to 25 users
    { duration: '15s', target: 50 },   // Ramp up to 50 users
    { duration: '15s', target: 100 },  // Ramp up to 100 users
    { duration: '15s', target: 150 },  // Ramp up to 150 users
    { duration: '15s', target: 200 },  // Peak load
    { duration: '15s', target: 100 },  // Ramp down
    { duration: '15s', target: 50 },   // Ramp down
    { duration: '15s', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    errors: ['rate<0.1'],             // Error rate must be less than 10%
  },
};

const BASE_URL = 'http://localhost:3020';
let authToken = '';

export function setup() {
  // Create a test user and get auth token
  const registerPayload = {
    email: `test-${Date.now()}@example.com`,
    password: 'Test123!',
    name: 'Test User'
  };

  const registerRes = http.post(`${BASE_URL}/authentication/register`, JSON.stringify(registerPayload), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (registerRes.status === 201) {
    const loginRes = http.post(`${BASE_URL}/authentication/login`, JSON.stringify({
      email: registerPayload.email,
      password: registerPayload.password
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    if (loginRes.status === 200) {
      const loginData = JSON.parse(loginRes.body);
      return { token: loginData.token, email: registerPayload.email };
    }
  }

  return { token: '', email: '' };
}

export default function (data) {
  const scenarios = [
    () => testRegister(),
    () => testLogin(),
    () => testValidateToken(data.token),
    () => testGetUsers(data.token),
    () => testGetUserByEmail(data.email, data.token),
  ];

  // Randomly select a scenario to execute
  const scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  scenario();

  sleep(1);
}

function testRegister() {
  const payload = {
    email: `user-${Math.random().toString(36).substring(7)}@example.com`,
    password: 'Test123!',
    name: 'Load Test User'
  };

  const res = http.post(`${BASE_URL}/authentication/register`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'register status is 201': (r) => r.status === 201,
    'register response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testLogin() {
  const payload = {
    email: 'test@example.com',
    password: 'wrongpassword'
  };

  const res = http.post(`${BASE_URL}/authentication/login`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'login response received': (r) => r.status !== 0,
    'login response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testValidateToken(token) {
  if (!token) return;

  const payload = { token: token };

  const res = http.post(`${BASE_URL}/authentication/validateToken`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'validate token response received': (r) => r.status !== 0,
    'validate token response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}

function testGetUsers(token) {
  if (!token) return;

  const res = http.get(`${BASE_URL}/authentication/users`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const success = check(res, {
    'get users response received': (r) => r.status !== 0,
    'get users response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

function testGetUserByEmail(email, token) {
  if (!email || !token) return;

  const res = http.get(`${BASE_URL}/authentication/user/email/${email}`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });

  const success = check(res, {
    'get user by email response received': (r) => r.status !== 0,
    'get user by email response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}