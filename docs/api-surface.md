# Merchant API contract

Source: [Capitalist Merchant Integration Guide](https://docs.capitalist.net/api/merchants/merchant-doc.html). Snapshot verified on 2026-11-04.

| Operation | Route and client transport | Fields | Response |
| --- | --- | --- | --- |
| Build payment form | Browser form POST to `/merchant/pay?lang=en` | `amount`, `currency`, `description`, `merchantid`, `number`, optional form parameters, `sign` | Hosted checkout |
| Create order | JSON POST to `/merchant/payGate/createorder` | Same as the payment form | Top-level `order` with `paymentUrl`, or validation errors |
| Check state | GET `/merchant/payGate/checkstate` | `amount`, `currency`, `merchant_id`, `order_number`, `sign` | `data.order` on success, including `paid`; error payload on failure |
| Quote payway | GET `/merchant/payGate/paywayinfo` | `amount`, `currency`, `merchant_id`, `opt_payway`, `sign`; see `opt_pc` in the guide | `data.order` may be null; `data.availablePaywayCurrencies` lists available currencies |
| Cancel | JSON POST to `/merchant/payGate/cancel` | `merchantId`, `orderNumber`, `signature` | Top-level `order` in the documented response; HTTP 422 when cancellation fails |
| Get crypto address | JSON POST to `/merchant/payGate/getorderadress` | `merchantid`, `amount`, `number`, `cryptoCurrency` | Provider response preserved as returned; guide supplies a request example |
| List payways | GET `/merchant/payGate/payways` | `all=1`, or `merchantid`, `amount`, `currency` | `data.payways` contains method codes and names |
| Get currency rate | GET `/merchant/payGate/currencyrate` | `merchantid`, `amount`, `currency`, `opt_pc` | `data.buy` and `data.sell` |
| List merchant orders | [Integration API Merchant Orders](https://docs.capitalist.net/api/integration-api.html#_merchant_orders) | Integration API credentials and signature | Use the adjacent Capitalist-Api2 project |

The clients choose one transport per operation. The guide also describes GET for creation and form or query variants for several methods. The mutation helpers use POST. Address, payways and rate requests follow the guide's unsigned examples.

## Signing and precision

- HMAC-MD5 uses the merchant secret and returns lowercase hexadecimal. The message is the colon-separated sequence of field **values** in lexicographic field-name order. Exclude the signature itself. Sign before JSON or query encoding.
- For create and form requests, this implementation follows the guide's PHP signing algorithm and includes every supplied field, including supplied `opt_*` fields. An omitted field is excluded and an explicitly empty string contributes an empty value. Supported form options are `opt_email`, `opt_payway`, `opt_pc`, `opt_lock_payway`, `opt_cb_method`, and `opt_client_id`.
- `checkstate` explicitly signs `amount:currency:merchant_id:order_number`. `cancel` explicitly signs `merchantId:orderNumber`. The quote signs `amount:currency:merchant_id:opt_payway`. Send `opt_pc` separately when supplied. The PHP backend’s `PaywayInfoForm::verifySign` selects those four fields explicitly. On 2026-11-04, the beta sample merchant returned HTTP 200 with a quote when `opt_pc` was sent unsigned. A control request that included it in the signature was rejected.
- Send decimal amounts and identifiers as strings. In responses TypeScript and PHP expose JSON numbers as their original text, Go uses `json.Number`, and Bash emits raw JSON. The SDKs leave endpoint-specific fields in the response envelope.
- The guide lists order currencies and precision rules in §3.2. Address requests use a separate `cryptoCurrency` vocabulary: `BTC`, `USDT`, `USDTt`, `USDTb`, `USDC`, `USDCb`, `ETH`. Availability and current limits are controlled by Capitalist. A quote is approximate.

## Payment result and notifications

Browser returns sign the five documented fields `merchant_id`, `order_number`, `payment_amount`, `payment_currency` and `payment_state`. Interaction URL notifications from Invest sign all transmitted callback fields except `sign`, in field-name order. The Invest `MerchantCallbackService` adds configured metadata, including `payments_info`, before signing. A notification without additional fields uses the same signature as a browser return.

Pass the decoded callback fields as strings and preserve empty values and the original decimal and JSON text. Reject duplicate keys and separate any known static Interaction URL query parameters from the callback fields. Parse `payments_info` only after verifying the signature. The field rules were checked against the local PHP and Invest backend sources and covered by local tests. Live Interaction URL verification with `payments_info` is pending.

Use the server notification and status reconciliation to update the local order. Check the expected merchant, order number, currency, exact decimal amount and success state before fulfillment. If a create request times out, reconcile using the original order number before deciding what to do next.
