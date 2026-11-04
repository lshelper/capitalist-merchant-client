import { createHmac, timingSafeEqual } from 'node:crypto';

type Fields = Readonly<Record<string, string>>;
type Amount = { amount: string; currency: string };
export type Payment = Amount & {
  merchantid: string;
  number: string;
  description: string;
  opt_email?: string;
  opt_payway?: string;
  opt_pc?: string;
  opt_lock_payway?: '0' | '1';
  opt_cb_method?: 'get' | 'post';
  opt_client_id?: string;
};
export type StateQuery = Amount & { merchant_id: string; order_number: string };
export type PaywayQuote = Amount & { merchant_id: string; opt_payway: string; opt_pc?: string };
export type OrderAddress = {
  merchantid: string;
  amount: string;
  number: string;
  cryptoCurrency: 'BTC' | 'USDT' | 'USDTt' | 'USDTb' | 'USDC' | 'USDCb' | 'ETH';
};
export type PaymentResult = {
  merchant_id: string;
  order_number: string;
  payment_amount: string;
  payment_currency: string;
  payment_state: string;
  sign: string;
};
export type ApiResponse = {
  success?: boolean;
  error?: string;
  data?: Record<string, unknown>;
  order?: Record<string, unknown>;
  [key: string]: unknown;
};

export function sign(fields: Fields, secret: string, signatureField = 'sign'): string {
  const message = Object.keys(fields)
    .filter(key => key !== signatureField)
    .sort()
    .map(key => fields[key])
    .join(':');
  return createHmac('md5', secret).update(message).digest('hex');
}

function verifyResult(result: Partial<PaymentResult> & Fields, secret: string, notification: boolean): boolean {
  const { merchant_id, order_number, payment_amount, payment_currency, payment_state, sign: digest } = result;
  const fields = { merchant_id, order_number, payment_amount, payment_currency, payment_state };
  if (typeof digest !== 'string' || !/^[0-9a-f]{32}$/.test(digest) ||
      Object.values(fields).some(value => typeof value !== 'string' || value === '')) {
    return false;
  }
  if (notification && Object.values(result).some(value => typeof value !== 'string')) {
    return false;
  }
  return timingSafeEqual(Buffer.from(digest), Buffer.from(sign(notification ? result : fields as Fields, secret)));
}

export function verifyReturn(result: Partial<PaymentResult> & Fields, secret: string): boolean {
  return verifyResult(result, secret, false);
}

export function verifyNotification(result: Partial<PaymentResult> & Fields, secret: string): boolean {
  return verifyResult(result, secret, true);
}

function parseJson(text: string): unknown {
  return JSON.parse(text, (_key: string, value: unknown, context?: { source: string }) => {
    if (typeof value === 'number') {
      if (!context) throw new Error('Lossless JSON parsing requires Node.js 22 or newer');
      return context.source;
    }
    return value;
  });
}

/** Receives a value already decoded by the query or form parser. */
export function parsePaymentsInfo(value: string): Array<Record<string, unknown>> {
  const payments = parseJson(value);
  if (!Array.isArray(payments) || payments.some(item => !item || typeof item !== 'object' || Array.isArray(item))) {
    throw new TypeError('payments_info must be an array of objects');
  }
  return payments;
}

export class MerchantApiError extends Error {
  constructor(readonly status: number, readonly response: ApiResponse) {
    super(status >= 400 ? `Merchant API HTTP ${status}` : 'Merchant API rejected the request');
  }
}

export class MerchantClient {
  constructor(
    private readonly secret: string,
    private readonly baseUrl = 'https://api.capitalist.net',
    private readonly timeoutMs = 15000,
  ) {}

  private signed(params: Record<string, string | undefined>, key = 'sign'): Fields {
    const fields: Record<string, string> = {};
    for (const [name, value] of Object.entries(params)) {
      if (value !== undefined && name !== key) fields[name] = value;
    }
    return { ...fields, [key]: sign(fields, this.secret, key) };
  }

  buildPaymentForm(payment: Payment, lang = 'en') {
    const action = new URL('/merchant/pay', this.baseUrl);
    action.searchParams.set('lang', lang);
    return { action: action.href, method: 'POST' as const, fields: this.signed(payment) };
  }

  createOrder(payment: Payment) {
    return this.request('POST', 'createorder', this.signed(payment));
  }

  checkState(params: StateQuery) {
    return this.request('GET', 'checkstate', this.signed(params));
  }

  getPaywayInfo(params: PaywayQuote) {
    const { amount, currency, merchant_id, opt_payway, opt_pc } = params;
    const signed = this.signed({ amount, currency, merchant_id, opt_payway });
    return this.request('GET', 'paywayinfo', opt_pc === undefined ? signed : { ...signed, opt_pc });
  }

  cancelOrder(merchantId: string, orderNumber: string) {
    return this.request('POST', 'cancel', this.signed({ merchantId, orderNumber }, 'signature'));
  }

  getOrderAddress(params: OrderAddress) {
    return this.request('POST', 'getorderadress', params);
  }

  listPayways(params?: Amount & { merchantid: string }) {
    return this.request('GET', 'payways', params ?? { all: '1' });
  }

  getCurrencyRate(params: Amount & { merchantid: string; opt_pc: string }) {
    return this.request('GET', 'currencyrate', params);
  }

  private async request(method: 'GET' | 'POST', endpoint: string, fields: Fields): Promise<ApiResponse> {
    const url = new URL(`/merchant/payGate/${endpoint}`, this.baseUrl);
    const headers = new Headers({ Accept: 'application/json' });
    let body: string | undefined;
    if (method === 'GET') {
      url.search = new URLSearchParams(fields).toString();
    } else {
      headers.set('Content-Type', 'application/json');
      body = JSON.stringify(fields);
    }
    const response = await fetch(url, {
      method, headers, body, redirect: 'error', signal: AbortSignal.timeout(this.timeoutMs),
    });
    const raw = await response.text();
    let payload: unknown;
    try {
      payload = parseJson(raw);
    } catch (error) {
      if (!response.ok) throw new MerchantApiError(response.status, {});
      throw error;
    }
    if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
      if (!response.ok) throw new MerchantApiError(response.status, {});
      throw new TypeError('Merchant API response must be an object');
    }
    const result = payload as ApiResponse;
    if (!response.ok || result.success === false) throw new MerchantApiError(response.status, result);
    return result;
  }
}
