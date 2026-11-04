# Capitalist Merchant Client

Server-side clients for Capitalist Merchant API in TypeScript, Go, PHP and Bash, with an IntelliJ HTTP Client collection. The documented contract was checked on **2026-11-04**. See the [API surface](docs/api-surface.md) and [official integration guide](https://docs.capitalist.net/api/merchants/merchant-doc.html).

## ⚙️ Setup

If you’re new to Capitalist, [create a Capitalist account →](https://capitalist.net/reg?from=5ab603bd9f90dc8e733965d7a4a1cf0e) before setting up your merchant.

Create a merchant in Capitalist, confirm the shop domain, and configure the credited account, browser return URLs and server-side Interaction URL. Keep the merchant secret in the server environment or secret store. Give each local order a unique number within its merchant. Use the same number for status reconciliation.

## 💳 Create an order in TypeScript

Use Node.js 22 or later. In `clients/typescript` run `npm ci && npm run build`. On your server, import the built client and supply the merchant's identifier and secret through the environment.

```ts
import { MerchantClient } from './clients/typescript/dist/index.js';

const client = new MerchantClient(process.env.CAPITALIST_MERCHANT_SECRET!);
const payment = {
  merchantid: process.env.CAPITALIST_MERCHANT_ID!,
  number: 'demo-order-123',
  amount: '12.34',
  currency: 'USD',
  description: 'Demo purchase',
};

const created = await client.createOrder(payment);
const checkout = created.order?.paymentUrl;
// Validate the URL against the expected Capitalist checkout origin before redirecting.
```

For a form-based flow, use `client.buildPaymentForm(payment)` and HTML-escape the returned action and hidden input names and values when rendering. The secret stays on the server.

### 🧩 Go and PHP

Use Go 1.22+ or PHP 8.4+ with cURL. The Go package name is `merchant`; a consumer can reference the local module through a `replace` directive in its `go.mod`. The PHP client can be loaded with `require`. These examples query the payment methods.

```go
import merchant "capitalist-merchant-client"

// Inside a function with ctx and a configured secret:
client := merchant.NewClient(secret, "", 0)
result, err := client.ListPayways(ctx, nil)
```

```php
require __DIR__ . '/clients/php/Client.php';

$client = new Capitalist\Merchant\Client(getenv('CAPITALIST_MERCHANT_SECRET'));
$payways = $client->listPayways()['data']->payways;
```

### 💻 Bash

Requires Bash, cURL and Python 3. Export `CAPITALIST_MERCHANT_SECRET` through your local secret configuration.

```bash
source clients/bash/merchant.sh
merchant_list_payways
merchant_check_state "merchant_id=$CAPITALIST_MERCHANT_ID" \
    order_number=demo-order-123 amount=12.34 currency=USD
```

The Bash client writes raw JSON to standard output. A non-2xx response returns exit code 22. A JSON payload with `success: false` returns a nonzero exit code. In TypeScript, Go and PHP, API errors carry the HTTP status and decoded response. Inspect the check-state failure payload to distinguish the documented order-not-found and invalid-signature cases. A successful quote may have `data.order = null`.

### 🌐 IntelliJ HTTP Client

Copy `clients/http/http-client.env.json.example` to `http-client.env.json` in the same directory. Keep calculated signatures in `http-client.private.env.json` using the selected environment name. Calculate them from exactly the values in the request. For example, after sourcing the Bash client:

```bash
merchant_python json sign merchant_id=85123 order_number=testorder123 \
    amount=12.3400 currency=RUR
```

Copy the resulting `sign` into `stateSign`. For `createSign`, use the payment form fields and the `sign` key. For `quoteSign`, use `merchant_id`, `amount`, `currency`, `opt_payway` and the `sign` key. For `cancelSignature`, use `merchant_python json signature` with `merchantId` and `orderNumber`. Use the guide's form and URL encoding rules when replacing example values.

The example environment points at localhost. To start the fixture for manual requests, run `python3 tests/mock_api.py /tmp/merchant-fixture-url 8089` from the project root. Use the fixture secret `mysecret123` with those sample values. The fixture covers JSON API requests and returns 422 for cancellation. The checkout form requires the hosted checkout service.

## 🔐 Handle payment notifications

Use your framework's parsed query or form parameters, reject duplicate fields, then verify the notification using the merchant secret. Pass all callback fields, including optional metadata as received after transport decoding. Separate any known static parameters configured in the Interaction URL. The signed payment amount must retain the original decimal text for verification. Compare the amount to the stored amount using exact decimal arithmetic.

```ts
if (!verifyNotification(fields, merchantSecret)) return rejectNotification();

await db.transaction(async tx => {
  const order = await tx.lockOrder(fields.merchant_id, fields.order_number);
  if (!order || order.currency !== fields.payment_currency ||
      !decimalEquals(order.amount, fields.payment_amount) ||
      fields.payment_state !== 'success') {
    throw new Error('Notification does not match the order');
  }
  if (order.paid) return;
  await tx.markPaid(order.id);
  await tx.enqueueFulfillmentOnce(order.id);
});
```

In this example, `lockOrder`, the exact-decimal comparison and the transactional fulfillment queue belong to the shop. The browser return handler can verify the return signature and display the order state recorded by the shop. If `createOrder` times out, use `checkState` with the original order details to reconcile before considering another request.

## 🧪 Verification

Install the TypeScript dependencies once with `(cd clients/typescript && npm ci)`, then run `scripts/test.sh` with Node, Go, PHP with cURL, Bash, Python 3 and cURL installed. The suite runs the four clients against a shared loopback fixture. Run `scripts/update-upstream-docs.sh` with cURL and Pandoc to refresh both snapshots, then review `git diff -- docs/upstream`.
