import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 20 }, // Ramp up to 20 users
    { duration: '1m', target: 20 },  // Stay at 20 users for 1 minute
    { duration: '30s', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
  },
};

let accountId = null;
let creditCardId = null;
let pixKey = null;

export function setup() {
  // Create a checking account
  const createAccountResponse = http.post('http://localhost:8080/accounts', JSON.stringify({
    type: 'CHECKING'
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(createAccountResponse, {
    'account created successfully': (r) => r.status === 201,
  });

  const account = JSON.parse(createAccountResponse.body);
  accountId = account.id;

  // Create a credit card
  const createCardResponse = http.post('http://localhost:8080/accounts/credit-cards', JSON.stringify({
    account_id: accountId,
    limit: 5000.00
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(createCardResponse, {
    'credit card created successfully': (r) => r.status === 201,
  });

  const card = JSON.parse(createCardResponse.body);
  creditCardId = card.id;

  // Register PIX key
  const createPIXResponse = http.post('http://localhost:8080/accounts/pix-keys', JSON.stringify({
    account_id: accountId,
    key_type: 'EMAIL',
    key: `test.${randomString(8)}@example.com`
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(createPIXResponse, {
    'PIX key registered successfully': (r) => r.status === 201,
  });

  const pix = JSON.parse(createPIXResponse.body);
  pixKey = pix.key;

  return { accountId, creditCardId, pixKey };
}

export default function (data) {
  // Make a deposit
  const depositResponse = http.post('http://localhost:8080/accounts/transactions', JSON.stringify({
    account_id: data.accountId,
    type: 'CREDIT',
    amount: 1000.00,
    description: 'Initial deposit'
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(depositResponse, {
    'deposit successful': (r) => r.status === 201,
  });

  // Make a credit card purchase
  const purchaseResponse = http.post('http://localhost:8080/accounts/transactions', JSON.stringify({
    account_id: data.accountId,
    type: 'CARD_PURCHASE',
    amount: 100.00,
    description: 'Online purchase',
    credit_card_id: data.creditCardId
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(purchaseResponse, {
    'credit card purchase successful': (r) => r.status === 201,
  });

  // Make a PIX transfer
  const pixResponse = http.post('http://localhost:8080/accounts/transactions', JSON.stringify({
    account_id: data.accountId,
    type: 'PIX_SENT',
    amount: 50.00,
    description: 'PIX transfer',
    destination_key: data.pixKey
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(pixResponse, {
    'PIX transfer successful': (r) => r.status === 201,
  });

  sleep(1);
}
