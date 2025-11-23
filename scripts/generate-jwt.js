const crypto = require('crypto');

// JWT credentials from Kong
const key = '37U4VjFAXAItqfppIrZcsfhMBm5hFoxL';
const secret = 'your-256-bit-secret';

// Create JWT header
const header = {
  typ: 'JWT',
  alg: 'HS256'
};

// Create JWT payload
const payload = {
  iss: key,
  exp: Math.floor(Date.now() / 1000) + (365 * 24 * 60 * 60) // Expires in 1 year
};

// Base64 encode
const base64url = (str) => {
  return Buffer.from(JSON.stringify(str))
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
};

// Create signature
const headerEncoded = base64url(header);
const payloadEncoded = base64url(payload);
const signatureInput = `${headerEncoded}.${payloadEncoded}`;

const signature = crypto
  .createHmac('sha256', secret)
  .update(signatureInput)
  .digest('base64')
  .replace(/\+/g, '-')
  .replace(/\//g, '_')
  .replace(/=/g, '');

const token = `${headerEncoded}.${payloadEncoded}.${signature}`;

console.log('JWT Token generated successfully!');
console.log('\nToken:');
console.log(token);
console.log('\n\nUse this token in your frontend Authorization header as:');
console.log(`Bearer ${token}`);
