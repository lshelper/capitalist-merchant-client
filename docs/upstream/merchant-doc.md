<div id="header">

# Capitalist Merchant API: Merchant Integration Guide

<div id="toc" class="toc2">

<div id="toctitle">

Contents

</div>

- [1. Requirements](#_requirements)
- [2. Creating a Merchant](#_creating_a_merchant)
  - [2.1. Step 1: Enter Name and Domain](#_step_1_enter_name_and_domain)
  - [2.2. Step 2: Confirm Domain Ownership](#_step_2_confirm_domain_ownership)
  - [2.3. Step 3: Detailed Merchant Configuration](#_step_3_detailed_merchant_configuration)
- [3. Integration](#_integration)
  - [3.1. Payment Request Form](#payment_request_form)
  - [3.2. Payment Request Parameters](#_payment_request_parameters)
  - [3.3. Digital Signature Generation](#digital_signature_generation)
- [4. Order Creation API](#_order_creation_api)
  - [4.1. Request Example](#_request_example_2)
  - [4.2. Error Response](#_error_response)
  - [4.3. Successful Response](#_successful_response)
- [5. Customer Return Pages](#_customer_return_pages)
  - [5.1. Parameters Passed to Result Pages](#_parameters_passed_to_result_pages)
- [6. Payment Notification (`Interaction URL`)](#_payment_notification_interaction_url)
  - [6.1. Verifying Payment Information](#_verifying_payment_information)
- [7. Verifying Data via API](#_verifying_data_via_api)
  - [7.1. Request Parameters](#_request_parameters)
  - [7.2. Result](#_result)
- [8. Merchant Orders List](#_merchant_orders_list)
- [9. Calculating the Approximate Order Amount](#_calculating_the_approximate_order_amount)
  - [9.1. Request Parameters](#_request_parameters_2)
  - [9.2. Response Parameters](#_response_parameters)
- [10. Canceling an Order](#_canceling_an_order)
  - [10.1. Request Parameters](#_request_parameters_3)
- [11. Getting an Order Crypto Wallet Address](#_getting_an_order_crypto_wallet_address)
  - [11.1. Request Parameters](#_request_parameters_4)
- [12. Payment System Code List](#_payment_system_code_list)
  - [12.1. Response Example](#_response_example_3)
- [13. Currency Rate Request](#_currency_rate_request)
  - [13.1. Request Parameters](#_request_parameters_5)
- [14. Change History](#_change_history)

</div>

</div>

<div id="content">

<div id="preamble">

<div class="sectionbody">

<div class="paragraph">

*v.1.7.2 dated 15.05.2026*

</div>

</div>

</div>

<div class="sect1">

## 1. Requirements

<div class="sectionbody">

<div class="paragraph">

To connect the payment system to your website, you need a Capitalist.net account. If you do not have one yet, register at <a href="https://capitalist.net/reg" class="bare">https://capitalist.net/reg</a>.

</div>

</div>

</div>

<div class="sect1">

## 2. Creating a Merchant

<div class="sectionbody">

<div class="paragraph">

Before filling out the form, be ready to upload a verification file to your website (see step 2).

</div>

<div class="paragraph">

In the main menu, select **Add merchant** or open <a href="https://capitalist.net/merchants/add" class="bare">https://capitalist.net/merchants/add</a>.

</div>

<div class="sect2">

### 2.1. Step 1: Enter Name and Domain

<div class="ulist">

- The merchant name can contain up to 60 characters. It will be visible to users.

- Enter your website domain name.

</div>

</div>

<div class="sect2">

### 2.2. Step 2: Confirm Domain Ownership

<div class="paragraph">

Follow the instructions on the page, upload the verification file to your website, and click **Confirm**.

</div>

</div>

<div class="sect2">

### 2.3. Step 3: Detailed Merchant Configuration

<div class="ulist">

- Fill in the description, select a category, and upload a logo.

- In the settings section, fill in `Success URL`, `Fail URL`, `Status URL`, and `Interaction URL`.

- Automatic notifications about paid orders will be sent to `Interaction URL`.

- Save the merchant and wait for moderator approval.

</div>

</div>

</div>

</div>

<div class="sect1">

## 3. Integration

<div class="sectionbody">

<div class="paragraph">

You have now registered and created a merchant. The next important step is to prepare your website to work with the Capitalist API and configure the payment flow.

</div>

<div class="paragraph">

You can also open the sample payment page to see the payment flow in action: <a href="https://capitalist.net/merchant2/sample/payment" class="bare">https://capitalist.net/merchant2/sample/payment</a>

</div>

<div class="sect2">

### 3.1. Payment Request Form

<div class="paragraph">

This HTML form is the key integration point. It is used to send a customer to the payment page. The required parameters are passed in this form. As shown in the example, you can pass the `lang=en` or `ru` query parameter to choose the interface language.

</div>

<div class="paragraph">

`method="get"` is also supported, so parameters can be passed in the query string.

</div>

<div class="paragraph">

There is also a separate order creation API (see below) that lets you inspect validation errors.

</div>

<div class="sect3">

#### Request Example

<div class="listingblock">

<div class="content">

``` highlightjs
<form name="payment" method="post" action="https://api.capitalist.net/merchant/pay?lang=en"
      accept-charset="UTF-8">
  <input type="hidden" name="merchantid"   value="70" />
  <input type="hidden" name="number"       value="123981" />
  <input type="hidden" name="amount"       value="15.23" />
  <input type="hidden" name="currency"     value="USD" />
  <input type="hidden" name="description"  value="My description" />
  <input type="hidden" name="sign"         value="5ae3dbfdbac751f44355dedf5db59d02" />
</form>
```

</div>

</div>

</div>

</div>

<div class="sect2">

### 3.2. Payment Request Parameters

<div class="sect3">

#### Required Parameters

<table class="tableblock frame-all grid-all stretch">
<colgroup>
<col style="width: 14%" />
<col style="width: 14%" />
<col style="width: 21%" />
<col style="width: 14%" />
<col style="width: 35%" />
</colgroup>
<thead>
<tr>
<th class="tableblock halign-left valign-top">Key</th>
<th class="tableblock halign-left valign-top">Name</th>
<th class="tableblock halign-left valign-top">Format</th>
<th class="tableblock halign-left valign-top">Example</th>
<th class="tableblock halign-left valign-top">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td class="tableblock halign-left valign-top"><p><code>merchantid</code></p></td>
<td class="tableblock halign-left valign-top"><p>Merchant ID</p></td>
<td class="tableblock halign-left valign-top"><p><code>/^\d+$/</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>70</code></p></td>
<td class="tableblock halign-left valign-top"><p>Merchant identifier.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>number</code></p></td>
<td class="tableblock halign-left valign-top"><p>Payment No.</p></td>
<td class="tableblock halign-left valign-top"><p><code>/^[A-Z\d\-]{1,42}$/i</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>123981</code></p></td>
<td class="tableblock halign-left valign-top"><p>Order number. It is stored in the Capitalist billing system and allows you to identify the order and link it to orders in your own billing system. It must be unique per Merchant ID. Can contain digits, Latin letters, and a hyphen.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>amount</code></p></td>
<td class="tableblock halign-left valign-top"><p>Amount</p></td>
<td class="tableblock halign-left valign-top"><p><code>/^[\d\.]+$/</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>15.23</code></p></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>Order amount.</p>
</div>
<div class="ulist">
<ul>
<li><p>No more than two decimal places for USD, EUR, RUR, USDT, USDTt, USDTb, USDC, and USDCb.</p></li>
<li><p>No more than 8 decimal places for BTC.</p></li>
<li><p>Minimum: 1 RUR, 0.1 USD/EUR, 0.00001 BTC.</p></li>
<li><p>Minimum for USDT (Trc20): 5 USD/USDT.</p></li>
<li><p>Minimum for USDT (Erc20): 5 USD/USDT.</p></li>
<li><p>Minimum for USDT (Bep20): 5 USD/USDT.</p></li>
<li><p>Minimum for USDC (Erc20/Bep20): 5 USD/USDC.</p></li>
</ul>
</div>
</div></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>currency</code></p></td>
<td class="tableblock halign-left valign-top"><p>Currency</p></td>
<td class="tableblock halign-left valign-top"><p><code>/^(USD|EUR|RUR|BTC|USDT|USDTt|USDTb|USDT-TRC20|USDT-ERC20|USDT-BEP20|USDC|USDCb|USDC-BEP20)$/</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>USD</code></p></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>Order currency. Depending on it, the corresponding account configured in the merchant settings will be credited.</p>
</div>
<div class="ulist">
<ul>
<li><p>To enable Bitcoin payments, contact support. The order currency must be <code>BTC</code>.</p></li>
<li><p>For USDT (Erc20), allowed constants are <code>USDT</code> or <code>USDT-ERC20</code>.</p></li>
<li><p>For USDT (Trc20), allowed constants are <code>USDTt</code> or <code>USDT-TRC20</code>.</p></li>
<li><p>For USDT (Bep20), allowed constants are <code>USDTb</code> or <code>USDT-BEP20</code>.</p></li>
<li><p>For USDC (Erc20), use <code>USDC</code>.</p></li>
<li><p>For USDC (Bep20), allowed constants are <code>USDCb</code> or <code>USDC-BEP20</code>.</p></li>
</ul>
</div>
<div class="admonitionblock note">
<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<tbody>
<tr>
<td class="icon"><em></em></td>
<td class="content"><div class="paragraph">
<p>The <strong>USDT (Erc20)</strong> payment method is available only for orders with amount &gt;= 5 USDT/USDTt/USD.<br />
The <strong>USDT (Trc20)</strong> payment method is available only for orders with amount &gt;= 5 USDT/USDTt/USD.<br />
The <strong>USDT (Bep20)</strong> payment method is available only for orders with amount &gt;= 5 USDT/USDTb/USD.<br />
The <strong>USDC (Erc20/Bep20)</strong> payment methods are available only for orders in USD, USDC, or USDCb with amount &gt;= 5 USD/USDC.</p>
</div></td>
</tr>
</tbody>
</table>
</div>
</div></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>description</code></p></td>
<td class="tableblock halign-left valign-top"><p>Description</p></td>
<td class="tableblock halign-left valign-top"><p><code>/^.{1,150}$/</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>Hello.</code></p></td>
<td class="tableblock halign-left valign-top"><p>Payment description.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>sign</code></p></td>
<td class="tableblock halign-left valign-top"><p>Sign</p></td>
<td class="tableblock halign-left valign-top"></td>
<td class="tableblock halign-left valign-top"></td>
<td class="tableblock halign-left valign-top"><p>Digital signature. See <a href="#digital_signature_generation">Digital Signature Generation</a>.</p></td>
</tr>
</tbody>
</table>

</div>

<div class="sect3">

#### Optional `opt_*` Parameters

<table class="tableblock frame-all grid-all stretch">
<colgroup>
<col style="width: 22%" />
<col style="width: 22%" />
<col style="width: 55%" />
</colgroup>
<thead>
<tr>
<th class="tableblock halign-left valign-top">Key</th>
<th class="tableblock halign-left valign-top">Name</th>
<th class="tableblock halign-left valign-top">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_email</code></p></td>
<td class="tableblock halign-left valign-top"><p>Preselected email</p></td>
<td class="tableblock halign-left valign-top"><p>User email required on the second step.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_payway</code></p></td>
<td class="tableblock halign-left valign-top"><p>Preselected payway</p></td>
<td class="tableblock halign-left valign-top"><p>Payment method code. The full list of codes is provided at the end of the document.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_pc</code></p></td>
<td class="tableblock halign-left valign-top"><p>Preferred payment currency</p></td>
<td class="tableblock halign-left valign-top"><p>Preferred currency on the second payment step. If this currency is available and <code>opt_payway</code> is passed, the order amount on the second step will be shown in that currency. This is a decorative parameter.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_lock_payway</code></p></td>
<td class="tableblock halign-left valign-top"><p>Lock payway</p></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>Prevents the user from changing the payment method.</p>
</div>
<div class="ulist">
<ul>
<li><p><code>"1"</code> means changing the payment method is disabled.</p></li>
<li><p><code>"0"</code> (default) means changing the payment method is allowed.</p></li>
</ul>
</div>
</div></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_cb_method</code></p></td>
<td class="tableblock halign-left valign-top"><p>Callback method</p></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>The user will be returned to the success/fail/status URL with parameters sent using the selected method.</p>
</div>
<div class="ulist">
<ul>
<li><p><code>"get"</code></p></li>
<li><p><code>"post"</code> (default)</p></li>
</ul>
</div>
</div></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>opt_client_id</code></p></td>
<td class="tableblock halign-left valign-top"><p>Client ID</p></td>
<td class="tableblock halign-left valign-top"><p>String matching <code>/^.{0,64}$/</code>. Some identifier of your client.</p></td>
</tr>
</tbody>
</table>

</div>

</div>

<div class="sect2">

### 3.3. Digital Signature Generation

<div class="paragraph">

The digital signature is generated by concatenating the values of all required form parameters. Parameters are concatenated alphabetically using the `":"` character.

</div>

<div class="paragraph">

For the payment request form, the order is: `amount`, `currency`, `description`, `merchantid`, `number`, `opt_email`.

</div>

<div class="admonitionblock note">

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<tbody>
<tr>
<td class="icon"><em></em></td>
<td class="content"><div class="paragraph">
<p>If optional <code>opt_*</code> parameters are not passed, they are not included in the signature.</p>
</div></td>
</tr>
</tbody>
</table>

</div>

<div class="paragraph">

The resulting string is hashed using **HMAC-MD5**, where the merchant secret key is used as the secret. (HMAC description: <a href="https://en.wikipedia.org/wiki/HMAC" class="bare">https://en.wikipedia.org/wiki/HMAC</a>)

</div>

<div class="paragraph">

The `sign` parameter must contain the lowercase hexadecimal hash value.

</div>

<div class="sect3">

#### PHP Signature Example

<div class="listingblock">

<div class="content">

``` highlightjs
/**
 * Data signing method
 *
 * @param array  $data   Hash map to sign (for example, $_POST)
 * @param string $secret Merchant secret key
 *
 * @return string
 */
function signData($data, $secret)
{
    unset($data['sign']);
    ksort($data, SORT_STRING);
    $str = implode(':', $data);
    return hash_hmac('md5', $str, $secret);
}

$order = [
    'merchantid'  => 85123,
    'number'      => 'testorder123',
    'amount'      => '12.34',
    'currency'    => 'RUR',
    'description' => 'My demo payment.',
];
$order['sign'] = signData($order, 'mysecret123');

// var_dump($order):
//
// array (size=6)
//   'merchantid'  => int 85123
//   'number'      => string 'testorder123' (length=12)
//   'amount'      => string '12.34' (length=5)
//   'currency'    => string 'RUR' (length=3)
//   'description' => string 'My demo payment.' (length=16)
//   'sign'        => string '8bb5946e96999f81f8958a8c79d50d4b' (length=32)
```

</div>

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 4. Order Creation API

<div class="sectionbody">

<div class="paragraph">

To create an order or inspect order creation errors, send a GET or POST request to:

</div>

<div class="literalblock">

<div class="content">

    POST https://api.capitalist.net/merchant/payGate/createorder

</div>

</div>

<div class="paragraph">

The request parameters are the same as in [Payment Request Form](#payment_request_form).

</div>

<div class="paragraph">

The example below shows a POST request with a JSON body (`Content-Type: application/json`). Passing parameters in the query string via GET is also supported.

</div>

<div class="sect2">

### 4.1. Request Example

<div class="listingblock">

<div class="content">

``` highlightjs
POST https://api.capitalist.net/merchant/payGate/createorder
Accept: application/json
Content-Type: application/json

{
  "merchantid":  163,
  "amount":      50,
  "number":      "1657014969",
  "description": "test order creation",
  "currency":    "USD",
  "sign":        "mytestsign"
}
```

</div>

</div>

</div>

<div class="sect2">

### 4.2. Error Response

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "errors": {
    "sign": "Invalid signature."
  }
}
```

</div>

</div>

</div>

<div class="sect2">

### 4.3. Successful Response

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "order": {
    "description":       "My demo payment.",
    "amount":            50,
    "currency":          "USD",
    "number":            "1677063939",
    "merchant":          163,
    "status":            "NEW",
    "statusLocalized":   "Pending payment",
    "date":              1677063945,
    "email":             null,
    "paid":              false,
    "refunded":          false,
    "chargedBack":       false,
    "paymentUrl":        "https://capitalist.net/merchant/order/163/1677063939",
    "convertedAmount":   null,
    "convertedCurrency": null
  }
}
```

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 5. Customer Return Pages

<div class="sectionbody">

<div class="paragraph">

After payment, the payment gateway redirects the customer to a return page. On this page, Capitalist determines the payment state and redirects the customer to the corresponding result page (`Success URL`, `Fail URL`, `Status URL`, `Interaction URL`).

</div>

<div class="sect2">

### 5.1. Parameters Passed to Result Pages

| Key | Name | Example | Description |
|----|----|----|----|
| `merchant_id` | Merchant ID | `70` | Merchant identifier. |
| `order_number` | Payment No. | `123981` | Order number. |
| `payment_state` | Payment state | `success` | Payment status. Possible values: `success`, `fail`, `inprocess`. |
| `payment_amount` | Payment amount | `15.23` | Payment amount. |
| `payment_currency` | Payment currency | `USD` | Payment currency. (`USD`, `EUR`, `RUR`) |
| `sign` | Sign |  | Digital signature. Parameters are concatenated alphabetically in this order: `merchant_id`, `order_number`, `payment_amount`, `payment_currency`, `payment_state`. |

</div>

</div>

</div>

<div class="sect1">

## 6. Payment Notification (`Interaction URL`)

<div class="sectionbody">

<div class="paragraph">

When a payment is made, Capitalist sends a request with payment data to the interaction page (`Interaction URL`). This request contains all the information required to complete the order payment flow on your website.

</div>

<div class="sect2">

### 6.1. Verifying Payment Information

<div class="paragraph">

To avoid processing a compromised notification, implement verification on the interaction page.

</div>

<div class="sect3">

#### Merchant Callback Example

<div class="paragraph">

Below is an example of the payload that can be sent to the merchant callback (`Interaction URL`) via `GET` query parameters:

</div>

<div class="listingblock">

<div class="content">

``` highlightjs
order_number=1773983260437-9aspsb&payments_info=%5B%7B%22txid%22%3A%220x105324e4697645f61eff505785f586d8c2b0488feaa169367cf69f12a191db12%22%2C%22amount%22%3A2.00000000%2C%22currency%22%3A%22USDT%22%2C%22id%22%3A1051270200%7D%5D&sign=1369554cee8e1e25b9c6715da01c69c0&payment_state=success&payment_currency=USD&payment_amount=2.00000000&merchant_id=104991
```

</div>

</div>

<div class="paragraph">

Decoded, this callback means:

</div>

<div class="ulist">

- `merchant_id=104991` identifies the merchant that received the payment notification.

- `order_number=1773983260437-9aspsb` is the merchant order number that should be matched with the order in your system.

- `payment_state=success` means the payment completed successfully.

- `payment_amount=2.00000000` and `payment_currency=USD` describe the order amount and order currency recorded for the merchant.

- `sign=1369554cee8e1e25b9c6715da01c69c0` is the signature that must be verified before you trust the callback.

- `payments_info` is a URL-encoded JSON array with payment execution details.

</div>

<div class="paragraph">

In this example, `payments_info` decodes to the following JSON:

</div>

<div class="listingblock">

<div class="content">

``` highlightjs
[
  {
    "txid": "0x105324e4697645f61eff505785f586d8c2b0488feaa169367cf69f12a191db12",
    "amount": 2.00000000,
    "currency": "USDT",
    "id": 1051270200
  }
]
```

</div>

</div>

<div class="paragraph">

This means the order was paid by a single transaction for `2.00000000 USDT`. The `txid` field contains the blockchain transaction hash, and `id` is the internal payment identifier on the Capitalist side.

</div>

</div>

<div class="sect3">

#### Verify the Data Source

<div class="paragraph">

Make sure that the payment notification was sent directly by Capitalist and has not been compromised.

</div>

</div>

<div class="sect3">

#### Verify the Data

<div class="paragraph">

Always validate the following parameters in the payment notification:

</div>

<div class="ulist">

- **Merchant ID** (`merchant_id`)

- **Order number** (`order_number`)

- **Payment amount** (`payment_amount`)

- **Payment currency** (`payment_currency`)

- **Payment state** (`payment_state`) must be `success`

- **Digital signature** (`sign`) see [Digital Signature Generation](#digital_signature_generation)

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 7. Verifying Data via API

<div class="sectionbody">

<div class="paragraph">

To check the payment status of an order, send a GET or POST request to:

</div>

<div class="literalblock">

<div class="content">

    GET/POST https://api.capitalist.net/merchant/payGate/checkstate

</div>

</div>

<div class="sect2">

### 7.1. Request Parameters

| Parameter | Name | Example | Description |
|----|----|----|----|
| `merchant_id` | Merchant ID | `70` | Merchant identifier. |
| `order_number` | Payment No. | `123981` | Order number. |
| `amount` | Amount | `2.32` | Amount to be paid. |
| `currency` | Currency | `USD` | Order currency. |
| `sign` | Sign |  | Digital signature. Parameters are concatenated alphabetically in this order: `amount`, `currency`, `merchant_id`, `order_number`. |

</div>

<div class="sect2">

### 7.2. Result

<div class="paragraph">

The result is returned as JSON:

</div>

<div class="ulist">

- `success != false` means the order status could not be checked.

- `data.order.paid = true` means the order has been paid.

- `convertedAmount` and `convertedCurrency` are filled if auto-conversion to another currency is used during payment.

</div>

<div class="sect3">

#### Successful Response Example (Order Paid)

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "success": true,
  "data": {
    "order": {
      "merchant_id":         70,
      "order_number":        "1446466734",
      "description":         "My payment purpose.",
      "state":               "paid",
      "paid":                true,
      "payUrl":              "https://capitalist.net/merchant/order/70/1446466734",
      "convertedAmount":     314.06,
      "convertedCurrency":   "RUR"
    }
  }
}
```

</div>

</div>

</div>

<div class="sect3">

#### Invalid Signature Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "success": false,
  "error":   "Unable to verify the signature.",
  "data": {
    "errors": {
      "sign": ["Unable to verify the signature."]
    }
  }
}
```

</div>

</div>

</div>

<div class="sect3">

#### Order Not Found Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "success": false,
  "error":   "Unable to verify the order.",
  "data": {
    "errors": {
      "sign": ["Unable to verify the order."]
    }
  }
}
```

</div>

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 8. Merchant Orders List

<div class="sectionbody">

<div class="paragraph">

To retrieve the list of merchant orders, use the endpoint described in our payment API documentation: [Merchant Orders](https://docs.capitalist.net/api/integration-api.html#_merchant_orders).

</div>

</div>

</div>

<div class="sect1">

## 9. Calculating the Approximate Order Amount

<div class="sectionbody">

<div class="paragraph">

To get an approximate payment amount in a currency different from the order currency, send a request to:

</div>

<div class="literalblock">

<div class="content">

    GET/POST https://api.capitalist.net/merchant/payGate/paywayinfo

</div>

</div>

<div class="admonitionblock note">

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<tbody>
<tr>
<td class="icon"><em></em></td>
<td class="content"><div class="paragraph">
<p>When processing the response, check that <code>data.order != null</code>. See also the <code>opt_pc</code> parameter in <a href="#optional_opt_parameters">Optional <code>opt_*</code> Parameters</a>.</p>
</div></td>
</tr>
</tbody>
</table>

</div>

<div class="sect2">

### 9.1. Request Parameters

| Parameter | Name | Example | Description |
|----|----|----|----|
| `merchant_id` | Merchant ID | `70` | Merchant identifier. |
| `amount` | Amount | `71.94` | Amount to be paid. |
| `currency` | Currency | `EUR` | Order currency. |
| `opt_payway` | Selected payway | `mastercard` | Payment method code. The full list of codes is provided at the end of the document. |
| `sign` | Sign |  | Digital signature. |

<div class="sect3">

#### Request Example

<div class="listingblock">

<div class="content">

``` highlightjs
GET https://api.capitalist.net/merchant/payGate/paywayinfo?merchant_id=70&amount=71.94&currency=EUR&sign=b2053cea0a768666e7aa887e4dba9b44&opt_payway=mastercard
```

</div>

</div>

</div>

<div class="sect3">

#### Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "success": true,
  "data": {
    "target": {
      "amount":   "71.94",
      "currency": "EUR"
    },
    "availablePaywayCurrencies": ["RUR", "EUR", "UAH", "USD"],
    "order": {
      "merchantid":  "70",
      "opt_payway":  "mastercard",
      "amount":      "5197.98",
      "currency":    "RUR"
    }
  }
}
```

</div>

</div>

</div>

</div>

<div class="sect2">

### 9.2. Response Parameters

| Parameter | Description |
|----|----|
| `availablePaywayCurrencies` | List of currencies available on the second payment step. |
| `order { }` | If the amount cannot be calculated or the currency is unavailable, `data.order` may be `null`. |
| `order.amount` / `order.currency` | Amount and currency to invoice. |

</div>

</div>

</div>

<div class="sect1">

## 10. Canceling an Order

<div class="sectionbody">

<div class="paragraph">

An issued order can be canceled at any time. Only an **unpaid** order can be canceled.

</div>

<div class="literalblock">

<div class="content">

    POST https://api.capitalist.net/merchant/payGate/cancel

</div>

</div>

<div class="paragraph">

If the cancellation succeeds, the response code is `200`; otherwise it is `422`. The response body contains JSON with the order details.

</div>

<div class="paragraph">

Both GET with query parameters and POST (`application/x-www-form-urlencoded`, `application/json`) are supported.

</div>

<div class="sect2">

### 10.1. Request Parameters

<table class="tableblock frame-all grid-all stretch">
<colgroup>
<col style="width: 22%" />
<col style="width: 22%" />
<col style="width: 55%" />
</colgroup>
<thead>
<tr>
<th class="tableblock halign-left valign-top">Parameter</th>
<th class="tableblock halign-left valign-top">Example</th>
<th class="tableblock halign-left valign-top">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td class="tableblock halign-left valign-top"><p><code>merchantId</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>70</code></p></td>
<td class="tableblock halign-left valign-top"><p>Merchant identifier.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>orderNumber</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>1234589</code></p></td>
<td class="tableblock halign-left valign-top"><p>Order number.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>signature</code></p></td>
<td class="tableblock halign-left valign-top"></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>Digital signature. All parameters except <code>signature</code> are concatenated alphabetically with <code>:</code>, then:</p>
</div>
<div class="literalblock">
<div class="content">
<pre><code>Hash_HMAC-MD5(&quot;&lt;merchantId&gt;:&lt;orderNumber&gt;&quot;, &lt;merchantSecret&gt;)</code></pre>
</div>
</div>
</div></td>
</tr>
</tbody>
</table>

<div class="sect3">

#### Request Example

<div class="listingblock">

<div class="content">

``` highlightjs
POST https://api.capitalist.net/merchant/payGate/cancel
Accept: application/json
Content-Type: application/json

{
  "merchantId":  163,
  "orderNumber": 1665130908,
  "signature":   "b2053cea0a768666e7aa887e4dba9b44"
}
```

</div>

</div>

</div>

<div class="sect3">

#### Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "order": {
    "description":   "My demo payment.",
    "amount":        20,
    "currency":      "RUR",
    "number":        "1665130908",
    "merchant":      163,
    "status":        "CANCELLED",
    "statusLocalized": "CANCELLED",
    "date":          1665130909,
    "created":       true,
    "email":         null,
    "paid":          false,
    "paymentUrl":    "https://capitalist.net/merchant/order/163/1665130908"
  }
}
```

</div>

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 11. Getting an Order Crypto Wallet Address

<div class="sectionbody">

<div class="paragraph">

To get the crypto wallet address for paying an order, send a request to:

</div>

<div class="literalblock">

<div class="content">

    POST https://api.capitalist.net/merchant/payGate/getorderadress

</div>

</div>

<div class="paragraph">

Both GET with query parameters and POST (`application/x-www-form-urlencoded`, `application/json`) are supported.

</div>

<div class="sect2">

### 11.1. Request Parameters

<table class="tableblock frame-all grid-all stretch">
<colgroup>
<col style="width: 22%" />
<col style="width: 22%" />
<col style="width: 55%" />
</colgroup>
<thead>
<tr>
<th class="tableblock halign-left valign-top">Parameter</th>
<th class="tableblock halign-left valign-top">Example</th>
<th class="tableblock halign-left valign-top">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td class="tableblock halign-left valign-top"><p><code>merchantid</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>163</code></p></td>
<td class="tableblock halign-left valign-top"><p>Merchant identifier.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>amount</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>50</code></p></td>
<td class="tableblock halign-left valign-top"><p>Order amount.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>number</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>1657014970</code></p></td>
<td class="tableblock halign-left valign-top"><p>Order number.</p></td>
</tr>
<tr>
<td class="tableblock halign-left valign-top"><p><code>cryptoCurrency</code></p></td>
<td class="tableblock halign-left valign-top"><p><code>USDTt</code></p></td>
<td class="tableblock halign-left valign-top"><div class="content">
<div class="paragraph">
<p>Required cryptocurrency. Possible values:</p>
</div>
<div class="ulist">
<ul>
<li><p><code>BTC</code> for Bitcoin</p></li>
<li><p><code>USDT</code> for USDT (ERC-20)</p></li>
<li><p><code>USDTt</code> for USDT (TRC-20)</p></li>
<li><p><code>USDTb</code> for USDT (BEP-20)</p></li>
<li><p><code>USDC</code> for USDC (ERC-20)</p></li>
<li><p><code>USDCb</code> for USDC (BEP-20)</p></li>
<li><p><code>ETH</code> for Ethereum</p></li>
</ul>
</div>
</div></td>
</tr>
</tbody>
</table>

<div class="sect3">

#### Request Example

<div class="listingblock">

<div class="content">

``` highlightjs
POST https://api.capitalist.net/merchant/payGate/getorderadress
Accept: application/json
Content-Type: application/json

{
  "merchantid": 163,
  "amount": 50,
  "number": "1657014970",
  "cryptoCurrency": "USDTt"
}
```

</div>

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 12. Payment System Code List

<div class="sectionbody">

<div class="paragraph">

In the optional `opt_payway` parameter, you can pass the code of the selected payment system. The user will be redirected immediately to the next payment step for that system.

</div>

<div class="paragraph">

List of all codes (JSON):

</div>

<div class="literalblock">

<div class="content">

    https://api.capitalist.net/merchant/payGate/payways?all=1

</div>

</div>

<div class="paragraph">

List of codes filtered for a specific merchant, amount, and currency:

</div>

<div class="literalblock">

<div class="content">

    https://api.capitalist.net/merchant/payGate/payways?merchantid=70&amount=2.32&currency=USD

</div>

</div>

<div class="sect2">

### 12.1. Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "success": true,
  "data": {
    "payways": [
      { "code": "capitalist",   "name": "Capitalist.net"   },
      { "code": "webmoney",     "name": "WebMoney"         },
      { "code": "yandexmoney",  "name": "Yandex.Money"     },
      { "code": "epese",        "name": "EPESE"            }
    ]
  }
}
```

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 13. Currency Rate Request

<div class="sectionbody">

<div class="paragraph">

Some payment methods are available only for orders in a specific currency. For example, to accept Bitcoin via transfer, the order must be issued in `BTC`. To calculate the equivalent payment amount, use:

</div>

<div class="literalblock">

<div class="content">

    GET https://api.capitalist.net/merchant/payGate/currencyrate

</div>

</div>

<div class="admonitionblock note">

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<tbody>
<tr>
<td class="icon"><em></em></td>
<td class="content"><div class="paragraph">
<p>No signature is required for this request.</p>
</div></td>
</tr>
</tbody>
</table>

</div>

<div class="sect2">

### 13.1. Request Parameters

| Parameter    | Name                 | Example | Description          |
|--------------|----------------------|---------|----------------------|
| `merchantid` | Merchant ID          | `70`    | Merchant identifier. |
| `amount`     | Amount               | `71.94` | Amount.              |
| `currency`   | Currency             | `USD`   | Source currency.     |
| `opt_pc`     | Destination currency | `BTC`   | Target currency.     |

<div class="sect3">

#### Response Example

<div class="listingblock">

<div class="content">

``` highlightjs
{
  "data": {
    "sell": [
      {
        "comment":      "sell USD for BTC to capitalist",
        "pair":         "USD/BTC",
        "sellAmount":   "500",
        "sellCurrency": "USD",
        "amount":       0.051175,
        "currency":     "BTC"
      }
    ],
    "buy": [
      {
        "comment":     "buy USD for BTC from capitalist",
        "pair":        "USD/BTC",
        "buyAmount":   "500",
        "buyCurrency": "USD",
        "amount":      0.0544,
        "currency":    "BTC"
      }
    ]
  }
}
```

</div>

</div>

</div>

</div>

</div>

</div>

<div class="sect1">

## 14. Change History

<div class="sectionbody">

| Version | Date | Changes |
|----|----|----|
| 1.7.4 | 01.06.2026 | Changed the minimum limit for USDT Erc20 to 5 USD/USDT. Synchronized general RUR and BTC minimums with API validation. |
| 1.7.3 | 19.05.2026 | Added USDT and USDC BEP-20 merchant payment support details. |
| 1.7.2 | 15.05.2026 | Added USDC (ERC-20) support details. |
| 1.7.1 | 16.06.2023 | Added a merchant callback example with `GET` query parameters and decoded `payments_info` description. |

</div>

</div>

</div>

<div id="footer">

<div id="footer-text">

Last updated 2026-06-01 12:52:37 UTC

</div>

</div>
