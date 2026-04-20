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

  // ── health ────────────────────────────────────────────────────────────────

  group('health', () => {
    const res = http.get(`${BASE_URL}/health`);
    check(res, { 'GET /health returns 200': (r) => r.status === 200 });
  });

  // ── registration ──────────────────────────────────────────────────────────

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

  group('register validation: short name + bad email + short password', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'X', email: 'not-an-email', password: 'short' }),
      { headers }
    );
    check(res, { 'invalid register returns 400': (r) => r.status === 400 });
  });

  group('register validation: missing name', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ email: `missing-name-${uuidv4()}@example.com`, password }),
      { headers }
    );
    check(res, { 'missing name returns 400': (r) => r.status === 400 });
  });

  group('register validation: missing email', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'No Email', password }),
      { headers }
    );
    check(res, { 'missing email returns 400': (r) => r.status === 400 });
  });

  group('register validation: missing password', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'No Pass', email: `no-pass-${uuidv4()}@example.com` }),
      { headers }
    );
    check(res, { 'missing password returns 400': (r) => r.status === 400 });
  });

  group('register validation: email format @example.com', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'Bad Email', email: '@example.com', password }),
      { headers }
    );
    check(res, { '@example.com returns 400': (r) => r.status === 400 });
  });

  group('register validation: email format user@', () => {
    const res = http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'Bad Email', email: 'user@', password }),
      { headers }
    );
    check(res, { 'user@ returns 400': (r) => r.status === 400 });
  });

  // ── login ─────────────────────────────────────────────────────────────────

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
      JSON.stringify({ email, password: 'wrongpass123' }),
      { headers }
    );
    check(res, { 'wrong password returns 401': (r) => r.status === 401 });
  });

  group('login non-existent email returns 401', () => {
    const res = http.post(
      `${BASE_URL}/sessions`,
      JSON.stringify({ email: `ghost-${uuidv4()}@example.com`, password }),
      { headers }
    );
    check(res, { 'non-existent email returns 401': (r) => r.status === 401 });
  });

  // ── profile ───────────────────────────────────────────────────────────────

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

  // ── update profile ────────────────────────────────────────────────────────

  group('update me: name', () => {
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

  group('update me: email', () => {
    const newEmail = `updated-${uuidv4()}@example.com`;
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ email: newEmail }),
      { headers: authHeaders(token) }
    );
    check(res, {
      'email update returns 200': (r) => r.status === 200,
      'email updated': (r) => r.json('email') === newEmail,
    });
  });

  group('update me: email to already-taken email → 409', () => {
    // Register a second user
    const takenEmail = `taken-${uuidv4()}@example.com`;
    http.post(
      `${BASE_URL}/users`,
      JSON.stringify({ name: 'Taken', email: takenEmail, password }),
      { headers }
    );
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ email: takenEmail }),
      { headers: authHeaders(token) }
    );
    check(res, { 'taken email update returns 409': (r) => r.status === 409 });
  });

  group('update me: password then re-login', () => {
    const newPassword = 'newPassword456';
    // Get current email from profile
    const me = http.get(`${BASE_URL}/users/me`, { headers: authHeaders(token) });
    const currentEmail = me.json('email');

    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ password: newPassword }),
      { headers: authHeaders(token) }
    );
    check(res, { 'password update returns 200': (r) => r.status === 200 });

    // Login with new password
    const loginRes = http.post(
      `${BASE_URL}/sessions`,
      JSON.stringify({ email: currentEmail, password: newPassword }),
      { headers }
    );
    check(loginRes, { 'login with new password → 200': (r) => r.status === 200 });

    // Old password should fail
    const oldLoginRes = http.post(
      `${BASE_URL}/sessions`,
      JSON.stringify({ email: currentEmail, password }),
      { headers }
    );
    check(oldLoginRes, { 'login with old password → 401': (r) => r.status === 401 });

    // Update token for remaining tests
    token = loginRes.json('token');
  });

  group('update me with empty body returns 400', () => {
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({}),
      { headers: authHeaders(token) }
    );
    check(res, { 'empty update returns 400': (r) => r.status === 400 });
  });

  group('update me: short password returns 400', () => {
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ password: 'short' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'short password update returns 400': (r) => r.status === 400 });
  });

  group('update me: invalid email format returns 400', () => {
    const res = http.put(
      `${BASE_URL}/users/me`,
      JSON.stringify({ email: 'not-an-email' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'invalid email update returns 400': (r) => r.status === 400 });
  });
}
