<?php

declare(strict_types=1);

use Capitalist\Merchant\ApiError;
use Capitalist\Merchant\Client;

require __DIR__ . '/Client.php';

function check(bool $condition, string $message): void
{
    if (!$condition) {
        throw new RuntimeException($message);
    }
}

$secret = 'mysecret123';
$base = getenv('MERCHANT_TEST_URL');
$client = $base === false ? new Client($secret) : new Client($secret, $base);
$payment = [
    'merchantid' => '85123', 'number' => 'testorder123', 'amount' => '12.34',
    'currency' => 'RUR', 'description' => 'My demo payment.',
];
$form = $client->buildPaymentForm($payment);
check($form['fields']['sign'] === '8bb5946e96999f81f8958a8c79d50d4b', 'Invalid form signature');
$notification = [
    'merchant_id' => '70', 'order_number' => 'example', 'payment_amount' => '2.00000000',
    'payment_currency' => 'USD', 'payment_state' => 'success',
];
$notification['sign'] = Client::sign($notification, $secret);
check(Client::verifyNotification($notification, $secret), 'Valid notification rejected');
$extended = array_merge($notification, [
    'client_id' => '', 'payments_info' => '[]', 'sign' => 'b970a2012002e3ab3058a9462a3c34d5',
]);
check(Client::verifyNotification($extended, $secret), 'Extended notification rejected');
$extended['payments_info'] = '[{}]';
check(!Client::verifyNotification($extended, $secret), 'Changed payments_info accepted');
$extended['sign'] = $notification['sign'];
check(Client::verifyReturn($extended, $secret), 'Browser return fields rejected');
check(!Client::verifyNotification($extended, $secret), 'Browser return signature accepted for extended notification');
$notification['payment_amount'] = '2';
check(!Client::verifyNotification($notification, $secret), 'Changed notification accepted');
$payments = Client::parsePaymentsInfo('[{"amount":2.00000000,"txid":"a%20b"}]');
check($payments[0]->amount === '2.00000000' && $payments[0]->txid === 'a%20b', 'Invalid payments_info');

if ($base !== false) {
    $order = $client->createOrder($payment)['order'];
    check($order->amount === '12.3400' && $order->id === '9007199254740993', 'Lost numeric precision');
    check($order->label === 'quoted "x12"', 'JSON string changed');
    $state = ['merchant_id' => '85123', 'order_number' => 'testorder123', 'amount' => '12.3400', 'currency' => 'RUR'];
    check($client->checkState($state)['data']->order->paid === true, 'Expected paid state');
    $state['order_number'] = 'missing';
    try {
        $client->checkState($state);
        throw new RuntimeException('Expected missing order error');
    } catch (ApiError $error) {
        check($error->status === 200 && $error->response['error'] === 'Unable to verify the order.', 'Invalid API error');
    }
    $quote = $client->getPaywayInfo([
        'merchant_id' => '85123', 'amount' => '12.34', 'currency' => 'RUR', 'opt_payway' => 'visa', 'opt_pc' => 'USD',
    ]);
    check($quote['data']->order === null, 'Expected unavailable quote');
    try {
        $client->cancelOrder('85123', 'testorder123');
        throw new RuntimeException('Expected cancellation error');
    } catch (ApiError $error) {
        check($error->status === 422 && $error->response['order']->paid === true, 'Invalid cancellation error');
    }
    check($client->getOrderAddress([
        'merchantid' => '85123', 'amount' => '12.34', 'number' => 'testorder123', 'cryptoCurrency' => 'USDTt',
    ])['address'] === 'mock-address', 'Expected wallet address');
    check($client->listPayways()['data']->payways === [], 'Expected payways');
    check($client->getCurrencyRate([
        'merchantid' => '85123', 'amount' => '12.34', 'currency' => 'RUR', 'opt_pc' => 'BTC',
    ])['data']->buy === [], 'Expected currency rates');
}

echo "PHP tests passed\n";
