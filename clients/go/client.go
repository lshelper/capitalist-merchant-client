package merchant

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.capitalist.net"

type Fields map[string]string
type Response map[string]any

type APIError struct {
	Status   int
	Response Response
}

func (e *APIError) Error() string {
	return fmt.Sprintf("merchant API rejected the request (HTTP %d)", e.Status)
}

type Client struct {
	secret  string
	baseURL string
	http    *http.Client
}

func NewClient(secret, baseURL string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{secret: secret, baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

func Sign(fields Fields, secret, signatureField string) string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		if key != signatureField {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, fields[key])
	}
	mac := hmac.New(md5.New, []byte(secret))
	mac.Write([]byte(strings.Join(values, ":")))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyResult(fields Fields, secret string, notification bool) bool {
	signed := Fields{}
	for _, key := range []string{"merchant_id", "order_number", "payment_amount", "payment_currency", "payment_state"} {
		if fields[key] == "" {
			return false
		}
		signed[key] = fields[key]
	}
	if notification {
		signed = fields
	}
	return hmac.Equal([]byte(fields["sign"]), []byte(Sign(signed, secret, "sign")))
}

func VerifyReturn(fields Fields, secret string) bool {
	return verifyResult(fields, secret, false)
}

func VerifyNotification(fields Fields, secret string) bool {
	return verifyResult(fields, secret, true)
}

// ParsePaymentsInfo accepts the already URL-decoded query or form value.
func ParsePaymentsInfo(value string) ([]map[string]any, error) {
	var payments []map[string]any
	if err := decode(strings.NewReader(value), &payments); err != nil {
		return nil, err
	}
	if payments == nil {
		return nil, fmt.Errorf("payments_info must be an array of objects")
	}
	for _, payment := range payments {
		if payment == nil {
			return nil, fmt.Errorf("payments_info must be an array of objects")
		}
	}
	return payments, nil
}

func (c *Client) signed(fields Fields, key string) Fields {
	result := make(Fields, len(fields)+1)
	for name, value := range fields {
		result[name] = value
	}
	result[key] = Sign(result, c.secret, key)
	return result
}

type PaymentForm struct {
	Action string
	Method string
	Fields Fields
}

func (c *Client) BuildPaymentForm(fields Fields, lang string) PaymentForm {
	if lang == "" {
		lang = "en"
	}
	query := url.Values{"lang": {lang}}
	return PaymentForm{Action: c.baseURL + "/merchant/pay?" + query.Encode(), Method: http.MethodPost, Fields: c.signed(fields, "sign")}
}

func (c *Client) CreateOrder(ctx context.Context, payment Fields) (Response, error) {
	return c.request(ctx, http.MethodPost, "createorder", c.signed(payment, "sign"))
}

func (c *Client) CheckState(ctx context.Context, fields Fields) (Response, error) {
	return c.request(ctx, http.MethodGet, "checkstate", c.signed(fields, "sign"))
}

func (c *Client) GetPaywayInfo(ctx context.Context, fields Fields) (Response, error) {
	signed := c.signed(Fields{
		"amount": fields["amount"], "currency": fields["currency"],
		"merchant_id": fields["merchant_id"], "opt_payway": fields["opt_payway"],
	}, "sign")
	if preferredCurrency, ok := fields["opt_pc"]; ok {
		signed["opt_pc"] = preferredCurrency
	}
	return c.request(ctx, http.MethodGet, "paywayinfo", signed)
}

func (c *Client) CancelOrder(ctx context.Context, merchantID, orderNumber string) (Response, error) {
	fields := Fields{"merchantId": merchantID, "orderNumber": orderNumber}
	return c.request(ctx, http.MethodPost, "cancel", c.signed(fields, "signature"))
}

func (c *Client) GetOrderAddress(ctx context.Context, fields Fields) (Response, error) {
	return c.request(ctx, http.MethodPost, "getorderadress", fields)
}

func (c *Client) ListPayways(ctx context.Context, filter Fields) (Response, error) {
	if len(filter) == 0 {
		filter = Fields{"all": "1"}
	}
	return c.request(ctx, http.MethodGet, "payways", filter)
}

func (c *Client) GetCurrencyRate(ctx context.Context, fields Fields) (Response, error) {
	return c.request(ctx, http.MethodGet, "currencyrate", fields)
}

func (c *Client) request(ctx context.Context, method, endpoint string, fields Fields) (Response, error) {
	target := c.baseURL + "/merchant/payGate/" + endpoint
	var body *bytes.Reader
	if method == http.MethodGet {
		query := url.Values{}
		for key, value := range fields {
			query.Set(key, value)
		}
		target += "?" + query.Encode()
		body = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result Response
	err = decode(resp.Body, &result)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, &APIError{Status: resp.StatusCode, Response: result}
	}
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("merchant API response must be an object")
	}
	if result["success"] == false {
		return result, &APIError{Status: resp.StatusCode, Response: result}
	}
	return result, nil
}

func decode(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("unexpected data after JSON response")
	}
	return nil
}
