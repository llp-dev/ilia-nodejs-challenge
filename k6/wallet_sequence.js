import http from 'k6/http';
import { check, group, fail } from 'k6';
import crypto from 'k6/crypto';
import encoding from 'k6/encoding';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const JWT_SECRET = __ENV.WALLET_JWT_SECRET || 'ILIACHALLENGE';
const USER_ID = '550e8400-e29b-41d4-a716-446655440000';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    checks: ['rate==1.0'],
  },
};

function generateJWT(secret, userID) {
  const header = encoding.b64encode(JSON.stringify({ alg: 'HS256', typ: 'JWT' }), 'rawurl');
  const payload = encoding.b64encode(
    JSON.stringify({ sub: userID, iat: Math.floor(Date.now() / 1000) }),
    'rawurl'
  );
  const data = `${header}.${payload}`;
  const sig = encoding.b64encode(crypto.hmac('sha256', secret, data, 'binary'), 'rawurl');
  return `${data}.${sig}`;
}

function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

export default function () {
  const token = generateJWT(JWT_SECRET, USER_ID);
  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  let walletID;

  // ── health ────────────────────────────────────────────────────────────────

  group('health check', () => {
    const res = http.get(`${BASE_URL}/health`);
    check(res, { 'GET /health → 200': (r) => r.status === 200 });
  });

  // ── auth guard ────────────────────────────────────────────────────────────

  group('auth: missing token returns 401', () => {
    const res = http.get(`${BASE_URL}/wallets`);
    check(res, { 'no token → 401': (r) => r.status === 401 });
  });

  group('auth: invalid token returns 401', () => {
    const res = http.get(`${BASE_URL}/wallets`, {
      headers: { Authorization: 'Bearer invalid.token.here' },
    });
    check(res, { 'bad token → 401': (r) => r.status === 401 });
  });

  // ── wallet CRUD ───────────────────────────────────────────────────────────

  group('create wallet', () => {
    const res = http.post(
      `${BASE_URL}/wallets`,
      JSON.stringify({ user_id: USER_ID, description: 'k6 test wallet' }),
      { headers }
    );
    const ok = check(res, {
      'POST /wallets → 201': (r) => r.status === 201,
      'has id': (r) => !!JSON.parse(r.body).id,
      'balance starts at 0': (r) => Number(JSON.parse(r.body).balance) === 0,
    });
    if (!ok) fail('wallet creation failed — aborting sequence');
    walletID = JSON.parse(res.body).id;
  });

  group('create wallet: missing user_id → 400', () => {
    const res = http.post(
      `${BASE_URL}/wallets`,
      JSON.stringify({ description: 'no user' }),
      { headers }
    );
    check(res, { 'missing user_id → 400': (r) => r.status === 400 });
  });

  group('list wallets', () => {
    const res = http.get(`${BASE_URL}/wallets`, { headers });
    check(res, {
      'GET /wallets → 200': (r) => r.status === 200,
      'list is non-empty': (r) => JSON.parse(r.body).length > 0,
    });
  });

  group('get wallet by id', () => {
    const res = http.get(`${BASE_URL}/wallets/${walletID}`, { headers });
    check(res, {
      'GET /wallets/:id → 200': (r) => r.status === 200,
      'correct id returned': (r) => JSON.parse(r.body).id === walletID,
    });
  });

  group('get wallet: not found → 404', () => {
    const res = http.get(`${BASE_URL}/wallets/00000000-0000-0000-0000-000000000000`, { headers });
    check(res, { 'unknown wallet → 404': (r) => r.status === 404 });
  });

  group('update description', () => {
    const res = http.put(
      `${BASE_URL}/wallets/${walletID}`,
      JSON.stringify({ description: 'updated by k6' }),
      { headers }
    );
    check(res, {
      'PUT /wallets/:id → 200': (r) => r.status === 200,
      'description updated': (r) => JSON.parse(r.body).description === 'updated by k6',
    });
  });

  group('update description: missing field → 400', () => {
    const res = http.put(`${BASE_URL}/wallets/${walletID}`, JSON.stringify({}), { headers });
    check(res, { 'missing description → 400': (r) => r.status === 400 });
  });

  group('update description: unknown fields → 400', () => {
    const res = http.put(
      `${BASE_URL}/wallets/${walletID}`,
      JSON.stringify({ description: 'x', unknown: 'field' }),
      { headers }
    );
    check(res, { 'unknown fields → 400': (r) => r.status === 400 });
  });

  // ── transactions ──────────────────────────────────────────────────────────

  group('credit wallet (+100)', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '100.00', description: 'deposit', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'credit → 201': (r) => r.status === 201 });
  });

  group('balance after credit is 100', () => {
    const res = http.get(`${BASE_URL}/wallets/${walletID}`, { headers });
    check(res, { 'balance === 100': (r) => Number(JSON.parse(r.body).balance) === 100 });
  });

  group('debit wallet (-40)', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '-40.00', description: 'withdrawal', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'debit → 201': (r) => r.status === 201 });
  });

  group('balance after debit is 60', () => {
    const res = http.get(`${BASE_URL}/wallets/${walletID}`, { headers });
    check(res, { 'balance === 60': (r) => Number(JSON.parse(r.body).balance) === 60 });
  });

  group('transaction: insufficient balance → 422', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '-9999.00', description: 'overdraft', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'overdraft → 422': (r) => r.status === 422 });
  });

  group('transaction: duplicate operation_id → 409', () => {
    const opID = uuidv4();
    http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '1.00', description: 'first', operation_id: opID }),
      { headers }
    );
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '1.00', description: 'duplicate', operation_id: opID }),
      { headers }
    );
    check(res, { 'duplicate operation_id → 409': (r) => r.status === 409 });
  });

  group('transaction: zero value → 400', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '0', description: 'no-op', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'zero value → 400': (r) => r.status === 400 });
  });

  group('transaction: missing value → 400', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ description: 'no value', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'missing value → 400': (r) => r.status === 400 });
  });

  group('transaction: missing operation_id → 400', () => {
    const res = http.post(
      `${BASE_URL}/wallets/${walletID}/transactions`,
      JSON.stringify({ value: '10.00', description: 'no op id' }),
      { headers }
    );
    check(res, { 'missing operation_id → 400': (r) => r.status === 400 });
  });

  group('transaction: wallet not found → 404', () => {
    const res = http.post(
      `${BASE_URL}/wallets/00000000-0000-0000-0000-000000000000/transactions`,
      JSON.stringify({ value: '10.00', description: 'ghost', operation_id: uuidv4() }),
      { headers }
    );
    check(res, { 'unknown wallet transaction → 404': (r) => r.status === 404 });
  });
}
