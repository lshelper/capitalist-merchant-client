<?php

declare(strict_types=1);

namespace Capitalist\Merchant;

use InvalidArgumentException;
use JsonException;
use RuntimeException;

final class ApiError extends RuntimeException
{
    public function __construct(public readonly int $status, public readonly array $response)
    {
        parent::__construct("Merchant API rejected the request (HTTP $status)");
    }
}

final class Client
{
    public function __construct(
        private readonly string $secret,
        private readonly string $baseUrl = 'https://api.capitalist.net',
        private readonly int $timeoutMs = 15000,
    ) {
        if ($timeoutMs <= 0) {
            throw new InvalidArgumentException('Timeout must be positive');
        }
    }

    /** @param array<string, string> $fields */
    public static function sign(array $fields, string $secret, string $signatureField = 'sign'): string
    {
        unset($fields[$signatureField]);
        ksort($fields, SORT_STRING);

        return hash_hmac('md5', implode(':', $fields), $secret);
    }

    private static function verifyResult(array $fields, string $secret, bool $notification): bool
    {
        $signed = [];
        foreach (['merchant_id', 'order_number', 'payment_amount', 'payment_currency', 'payment_state'] as $key) {
            $value = $fields[$key] ?? null;
            if (!is_string($value) || $value === '') {
                return false;
            }
            $signed[$key] = $value;
        }

        if ($notification) {
            foreach ($fields as $value) {
                if (!is_string($value)) {
                    return false;
                }
            }
            $signed = $fields;
        }
        $signature = $fields['sign'] ?? null;
        return is_string($signature) && hash_equals(self::sign($signed, $secret), $signature);
    }

    public static function verifyReturn(array $fields, string $secret): bool
    {
        return self::verifyResult($fields, $secret, false);
    }

    public static function verifyNotification(array $fields, string $secret): bool
    {
        return self::verifyResult($fields, $secret, true);
    }

    /** Receives a value already decoded by the query or form parser. */
    public static function parsePaymentsInfo(string $value): array
    {
        $payments = self::decode($value);
        if (!is_array($payments)) {
            throw new InvalidArgumentException('payments_info must be an array of objects');
        }
        foreach ($payments as $payment) {
            if (!$payment instanceof \stdClass) {
                throw new InvalidArgumentException('payments_info must be an array of objects');
            }
        }

        return $payments;
    }

    public function buildPaymentForm(array $payment, string $lang = 'en'): array
    {
        return [
            'action' => rtrim($this->baseUrl, '/') . '/merchant/pay?lang=' . rawurlencode($lang),
            'method' => 'POST',
            'fields' => $this->signed($payment),
        ];
    }

    public function createOrder(array $payment): array
    {
        return $this->request('POST', 'createorder', $this->signed($payment));
    }

    public function checkState(array $fields): array
    {
        return $this->request('GET', 'checkstate', $this->signed($fields));
    }

    public function getPaywayInfo(array $fields): array
    {
        $signed = $this->signed(array_intersect_key($fields, array_flip([
            'amount', 'currency', 'merchant_id', 'opt_payway',
        ])));
        if (array_key_exists('opt_pc', $fields)) {
            $signed['opt_pc'] = $fields['opt_pc'];
        }

        return $this->request('GET', 'paywayinfo', $signed);
    }

    public function cancelOrder(string $merchantId, string $orderNumber): array
    {
        $fields = $this->signed(compact('merchantId', 'orderNumber'), 'signature');
        return $this->request('POST', 'cancel', $fields);
    }

    public function getOrderAddress(array $fields): array
    {
        return $this->request('POST', 'getorderadress', $fields);
    }

    public function listPayways(array $filter = []): array
    {
        return $this->request('GET', 'payways', $filter ?: ['all' => '1']);
    }

    public function getCurrencyRate(array $fields): array
    {
        return $this->request('GET', 'currencyrate', $fields);
    }

    private function signed(array $fields, string $key = 'sign'): array
    {
        $fields[$key] = self::sign($fields, $this->secret, $key);

        return $fields;
    }

    private function request(string $method, string $endpoint, array $fields): array
    {
        $url = rtrim($this->baseUrl, '/') . '/merchant/payGate/' . $endpoint;
        $headers = ['Accept: application/json'];
        $options = [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT_MS => $this->timeoutMs,
            CURLOPT_FOLLOWLOCATION => false,
        ];
        if ($method === 'GET') {
            $url .= '?' . http_build_query($fields, '', '&', PHP_QUERY_RFC3986);
        } else {
            $headers[] = 'Content-Type: application/json';
            $options[CURLOPT_POST] = true;
            $options[CURLOPT_POSTFIELDS] = json_encode((object) $fields, JSON_THROW_ON_ERROR);
        }
        $options[CURLOPT_HTTPHEADER] = $headers;
        $handle = curl_init($url);
        curl_setopt_array($handle, $options);
        $raw = curl_exec($handle);
        if ($raw === false) {
            throw new RuntimeException('Merchant API transport failed (cURL ' . curl_errno($handle) . ')');
        }
        $status = curl_getinfo($handle, CURLINFO_RESPONSE_CODE);
        $ok = $status >= 200 && $status < 300;
        try {
            $decoded = self::decode($raw);
        } catch (JsonException $error) {
            if (!$ok) {
                throw new ApiError($status, []);
            }
            throw $error;
        }
        if (!$ok) {
            throw new ApiError($status, $decoded instanceof \stdClass ? (array) $decoded : []);
        }
        if (!$decoded instanceof \stdClass) {
            throw new RuntimeException('Merchant API response must be an object');
        }
        $result = (array) $decoded;
        if (($result['success'] ?? null) === false) {
            throw new ApiError($status, $result);
        }

        return $result;
    }

    private static function decode(string $json): mixed
    {
        if (!json_validate($json)) {
            throw new JsonException('Invalid Merchant API JSON');
        }

        // Skip complete JSON strings and quote numeric tokens before decoding.
        $lossless = preg_replace_callback(
            '~"(?:[^"\\\\]++|\\\\.)*"(*SKIP)(*F)|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?~s',
            static fn (array $match): string => '"' . $match[0] . '"',
            $json,
        );
        if ($lossless === null) {
            throw new JsonException('Could not preserve Merchant API numbers');
        }

        return json_decode($lossless, false, 512, JSON_THROW_ON_ERROR);
    }
}
