import http from 'k6/http';
import { check, group, fail } from 'k6';

const WALLET_BASE_URL = __ENV.WALLET_BASE_URL || __ENV.BASE_URL || 'http://localhost:3001';
const USERS_BASE_URL  = __ENV.USERS_BASE_URL  || 'http://localhost:3002';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1.0'],
  },
};

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

const jsonHeaders = { 'Content-Type': 'application/json' };

function authHeaders(token) {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` };
}

export default function () {
  // ── bootstrap: register + login via users service to get a real JWT ────────
  // The wallet validates user_id against the users service, so the caller must
  // be a real registered user.

  const email    = `k6-${uuidv4()}@example.com`;
  const password = 'Password123!';
  let token  = '';
  let userID = '';

  group('users: register', () => {
    const res = http.post(
      `${USERS_BASE_URL}/users`,
      JSON.stringify({ name: 'k6 user', email, password }),
      { headers: jsonHeaders }
    );
    const ok = check(res, { 'POST /users → 201': (r) => r.status === 201 });
    if (!ok) fail(`register failed (${res.status}): ${res.body}`);
    userID = res.json('id');
  });

  group('users: login', () => {
    const res = http.post(
      `${USERS_BASE_URL}/sessions`,
      JSON.stringify({ email, password }),
      { headers: jsonHeaders }
    );
    const ok = check(res, { 'POST /sessions → 200': (r) => r.status === 200 });
    if (!ok) fail(`login failed (${res.status}): ${res.body}`);
    token = res.json('token');
    if (!token) fail('no token returned from login');
  });

  // Register a second user so we can test the mismatched user_id → 403 case.
  let otherUserID = '';
  group('users: register second user', () => {
    const res = http.post(
      `${USERS_BASE_URL}/users`,
      JSON.stringify({ name: 'k6 other', email: `k6-other-${uuidv4()}@example.com`, password }),
      { headers: jsonHeaders }
    );
    const ok = check(res, { 'POST /users (other) → 201': (r) => r.status === 201 });
    if (!ok) fail(`register other user failed (${res.status}): ${res.body}`);
    otherUserID = res.json('id');
  });

  let walletID;

  // ── health ────────────────────────────────────────────────────────────────

  group('health check', () => {
    const res = http.get(`${WALLET_BASE_URL}/health`);
    check(res, { 'GET /health → 200': (r) => r.status === 200 });
  });

  // ── auth guard ────────────────────────────────────────────────────────────

  group('auth: missing token returns 401', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets`);
    check(res, { 'no token → 401': (r) => r.status === 401 });
  });

  group('auth: invalid token returns 401', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets`, {
      headers: { Authorization: 'Bearer invalid.token.here' },
    });
    check(res, { 'bad token → 401': (r) => r.status === 401 });
  });

  // ── wallet CRUD ───────────────────────────────────────────────────────────

  group('create wallet', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets`,
      JSON.stringify({ user_id: userID, description: 'k6 test wallet' }),
      { headers: authHeaders(token) }
    );
    const ok = check(res, {
      'POST /wallets → 201': (r) => r.status === 201,
      'has id': (r) => !!r.json('id'),
      'balance starts at 0': (r) => Number(r.json('balance')) === 0,
    });
    if (!ok) fail(`wallet creation failed (${res.status}): ${res.body}`);
    walletID = res.json('id');
  });

  group('create wallet: mismatched user_id → 403', () => {
    // otherUserID belongs to a different registered user — token email won't match.
    const res = http.post(
      `${WALLET_BASE_URL}/wallets`,
      JSON.stringify({ user_id: otherUserID, description: 'other user' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'mismatched user_id → 403': (r) => r.status === 403 });
  });

  group('create wallet: missing user_id → 400', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets`,
      JSON.stringify({ description: 'no user' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'missing user_id → 400': (r) => r.status === 400 });
  });

  group('list wallets', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets`, { headers: authHeaders(token) });
    check(res, {
      'GET /wallets → 200': (r) => r.status === 200,
      'list is non-empty': (r) => r.json().length > 0,
    });
  });

  group('get wallet by id', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets/${walletID}`, { headers: authHeaders(token) });
    check(res, {
      'GET /wallets/:id → 200': (r) => r.status === 200,
      'correct id returned': (r) => r.json('id') === walletID,
    });
  });

  group('get wallet: not found → 404', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets/00000000-0000-0000-0000-000000000000`, { headers: authHeaders(token) });
    check(res, { 'unknown wallet → 404': (r) => r.status === 404 });
  });

  group('update description', () => {
    const res = http.put(
      `${WALLET_BASE_URL}/wallets/${walletID}`,
      JSON.stringify({ description: 'updated by k6' }),
      { headers: authHeaders(token) }
    );
    check(res, {
      'PUT /wallets/:id → 200': (r) => r.status === 200,
      'description updated': (r) => r.json('description') === 'updated by k6',
    });
  });

  group('update description: missing field → 400', () => {
    const res = http.put(`${WALLET_BASE_URL}/wallets/${walletID}`, JSON.stringify({}), { headers: authHeaders(token) });
    check(res, { 'missing description → 400': (r) => r.status === 400 });
  });

  group('update description: unknown fields → 400', () => {
    const res = http.put(
      `${WALLET_BASE_URL}/wallets/${walletID}`,
      JSON.stringify({ description: 'x', unknown: 'field' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'unknown fields → 400': (r) => r.status === 400 });
  });

  // ── transactions ──────────────────────────────────────────────────────────

  group('credit wallet (+100)', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '100.00', description: 'deposit', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'credit → 201': (r) => r.status === 201 });
  });

  group('balance after credit is 100', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets/${walletID}`, { headers: authHeaders(token) });
    check(res, { 'balance === 100': (r) => Number(r.json('balance')) === 100 });
  });

  group('debit wallet (-40)', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '-40.00', description: 'withdrawal', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'debit → 201': (r) => r.status === 201 });
  });

  group('balance after debit is 60', () => {
    const res = http.get(`${WALLET_BASE_URL}/wallets/${walletID}`, { headers: authHeaders(token) });
    check(res, { 'balance === 60': (r) => Number(r.json('balance')) === 60 });
  });

  group('transaction: insufficient balance → 422', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '-9999.00', description: 'overdraft', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'overdraft → 422': (r) => r.status === 422 });
  });

  group('transaction: duplicate operation_id → 409', () => {
    const opID = uuidv4();
    http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '1.00', description: 'first', operation_id: opID }),
      { headers: authHeaders(token) }
    );
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '1.00', description: 'duplicate', operation_id: opID }),
      { headers: authHeaders(token) }
    );
    check(res, { 'duplicate operation_id → 409': (r) => r.status === 409 });
  });

  group('transaction: zero value → 400', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '0', description: 'no-op', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'zero value → 400': (r) => r.status === 400 });
  });

  group('transaction: missing value → 400', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ description: 'no value', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'missing value → 400': (r) => r.status === 400 });
  });

  group('transaction: missing operation_id → 400', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '10.00', description: 'no op id' }),
      { headers: authHeaders(token) }
    );
    check(res, { 'missing operation_id → 400': (r) => r.status === 400 });
  });

  group('transaction: wallet not found → 404', () => {
    const res = http.post(
      `${WALLET_BASE_URL}/wallets/00000000-0000-0000-0000-000000000000/transactions`,
      JSON.stringify({ value: '10.00', description: 'ghost', operation_id: uuidv4() }),
      { headers: authHeaders(token) }
    );
    check(res, { 'unknown wallet transaction → 404': (r) => r.status === 404 });
  });
}
