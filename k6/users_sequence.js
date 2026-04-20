import http from 'k6/http';
import { check, group, fail } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3002';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1.0'],
  },
};

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

const headers = { 'Content-Type': 'application/json' };

function authHeaders(token) {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` };
}

export default function () {
  const email = `user-${uuidv4()}@example.com`;
  const password = 'password123';
  let token = '';
  let userID = '';

  group('health', () => {
    const res = http.get(`${BASE_URL}/health`);
    check(res, { 'GET /health returns 200': (r) => r.status === 200 });
  });

  group('register', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'Test User', email, password }),
      { headers }
    );
    check(res, {
      'POST /users returns 201': (r) => r.status === 201,
      'user has id': (r) => r.json('id') !== '',
      'user has email': (r) => r.json('email') === email,
    });
    userID = res.json('id');
  });

  group('register duplicate returns 409', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'Test User', email, password }),
      { headers }
    );
    check(res, { 'duplicate register returns 409': (r) => r.status === 409 });
  });

  group('register validation', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'X', email: 'not-an-email', password: 'short' }),
      { headers }
    );
    check(res, { 'invalid register returns 400': (r) => r.status === 400 });
  });

  group('login', () => {
    const res = http.post(
      `${BASE_URL}/sessions`,
      JSON.stringify({ email, password }),
      { headers }
    );
    check(res, {
      'POST /sessions returns 200': (r) => r.status === 200,
      'response has token': (r) => r.json('token') !== '',
    });
    token = res.json('token');
    if (!token) fail('no token returned from login');
  });

  group('login wrong password returns 401', () => {
    const res = http.post(
      `${BASE_URL}/sessions`,
      JSON.stringify({ email, password: 'wrongpass' }),
      { headers }
    );
    check(res, { 'wrong password returns 401': (r) => r.status === 401 });
  });

  group('get me', () => {
    const res = http.get(`${BASE_URL}/users/me`, { headers: authHeaders(token) });
    check(res, {
      'GET /users/me returns 200': (r) => r.status === 200,
      'me.id matches': (r) => r.json('id') === userID,
    });
  });

  group('get me without token returns 401', () => {
    const res = http.get(`${BASE_URL}/users/me`, { headers });
    check(res, { 'no-token returns 401': (r) => r.status === 401 });
  });

  group('update me', () => {
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ name: 'Updated Name' }),
      { headers: authHeaders(token) }
    );
    check(res, {
      'PUT /users/me returns 200': (r) => r.status === 200,
      'name updated': (r) => r.json('name') === 'Updated Name',
    });
  });

  group('update me with empty body returns 400', () => {
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({}),
      { headers: authHeaders(token) }
    );
    check(res, { 'empty update returns 400': (r) => r.status === 400 });
  });
}
