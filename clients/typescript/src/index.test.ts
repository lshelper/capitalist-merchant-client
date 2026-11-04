import assert from 'node:assert/strict';
import test from 'node:test';
import { MerchantApiError, MerchantClient, parsePaymentsInfo, sign, verifyNotification, verifyReturn } from './index.js';

const secret = 'mysecret123';
const payment = {
  merchantid: '85123', number: 'testorder123', amount: '12.34',
  currency: 'RUR', description: 'My demo payment.',
};

test('official signature vector and form', () => {
  const client = new MerchantClient(secret);
  assert.deepEqual(client.buildPaymentForm(payment), {
    action: 'https://api.capitalist.net/merchant/pay?lang=en',
    method: 'POST',
    fields: { ...payment, sign: '8bb5946e96999f81f8958a8c79d50d4b' },
  });
  const optional = { ...payment, description: 'Оплата', opt_email: '' };
  assert.notEqual(sign(optional, secret), sign({ ...payment, description: 'Оплата' }, secret));
});

test('notification verification and already decoded payment data', () => {
  const fields = {
    merchant_id: '70', order_number: 'example', payment_amount: '2.00000000',
    payment_currency: 'USD', payment_state: 'success',
  };
  const signed = { ...fields, sign: sign(fields, secret) };
  assert.ok(verifyNotification(signed, secret));
  const extended = { ...fields, client_id: '', payments_info: '[]', sign: 'b970a2012002e3ab3058a9462a3c34d5' };
  assert.ok(verifyNotification(extended, secret));
  // Simulate a query parser returning an array despite the caller's TypeScript type.
  const malformed = { ...extended, payments_info: ['[]'] } as unknown as typeof extended;
  assert.ok(!verifyNotification(malformed, secret));
  assert.ok(!verifyNotification({ ...extended, payments_info: '[{}]' }, secret));
  assert.ok(!verifyNotification({ ...extended, sign: signed.sign }, secret));
  assert.ok(verifyReturn({ ...extended, sign: signed.sign }, secret));
  assert.ok(!verifyNotification({ ...signed, payment_amount: '2' }, secret));
  assert.ok(!verifyNotification({ sign: signed.sign }, secret));
  assert.deepEqual(parsePaymentsInfo('[{"amount":2.00000000,"txid":"a%20b"}]'), [
    { amount: '2.00000000', txid: 'a%20b' },
  ]);
  assert.throws(() => parsePaymentsInfo('[null]'));
});

test('documented API routes, errors and decimal precision', { skip: !process.env.MERCHANT_TEST_URL }, async () => {
  const client = new MerchantClient(secret, process.env.MERCHANT_TEST_URL);
  const created = await client.createOrder(payment);
  assert.equal(created.order?.amount, '12.3400');
  assert.equal(created.order?.id, '9007199254740993');
  assert.equal(created.order?.label, 'quoted "x12"');
  const state = { merchant_id: '85123', order_number: 'testorder123', amount: '12.3400', currency: 'RUR' };
  assert.deepEqual((await client.checkState(state)).data?.order, { paid: true });
  await assert.rejects(client.checkState({ ...state, order_number: 'missing' }),
    (error: MerchantApiError) => error.status === 200 && error.response.error === 'Unable to verify the order.');
  assert.equal((await client.getPaywayInfo({
    merchant_id: '85123', amount: '12.34', currency: 'RUR', opt_payway: 'visa', opt_pc: 'USD',
  })).data?.order, null);
  await assert.rejects(client.cancelOrder('85123', 'testorder123'),
    (error: MerchantApiError) => error.status === 422 && error.response.order?.paid === true);
  assert.equal((await client.getOrderAddress({
    merchantid: '85123', amount: '12.34', number: 'testorder123', cryptoCurrency: 'USDTt',
  })).address, 'mock-address');
  assert.deepEqual((await client.listPayways()).data?.payways, []);
  assert.deepEqual((await client.listPayways({
    merchantid: '85123', amount: '12.34', currency: 'RUR',
  })).data?.payways, []);
  const wrongSecret = new MerchantClient('wrong-secret', process.env.MERCHANT_TEST_URL);
  await assert.rejects(wrongSecret.checkState(state),
    (error: MerchantApiError) => error.status === 200 && error.response.error === 'Unable to verify the signature.');
  assert.deepEqual((await client.getCurrencyRate({
    merchantid: '85123', amount: '12.34', currency: 'RUR', opt_pc: 'BTC',
  })).data?.buy, []);
});
