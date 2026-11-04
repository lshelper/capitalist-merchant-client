package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"
)

const secret = "mysecret123"

var payment = Fields{
	"merchantid": "85123", "number": "testorder123", "amount": "12.34",
	"currency": "RUR", "description": "My demo payment.",
}

func TestSignAndVerify(t *testing.T) {
	form := NewClient(secret, "", 0).BuildPaymentForm(payment, "")
	if form.Action != "https://api.capitalist.net/merchant/pay?lang=en" ||
		form.Fields["sign"] != "8bb5946e96999f81f8958a8c79d50d4b" || payment["sign"] != "" {
		t.Fatal("invalid payment form or input mutated")
	}
	fields := Fields{"merchant_id": "70", "order_number": "example", "payment_amount": "2.00000000",
		"payment_currency": "USD", "payment_state": "success"}
	fields["sign"] = Sign(fields, secret, "sign")
	returnDigest := fields["sign"]
	fields["payments_info"] = "[]"
	fields["client_id"] = ""
	fields["sign"] = "b970a2012002e3ab3058a9462a3c34d5"
	if !VerifyNotification(fields, secret) {
		t.Fatal("valid callback rejected")
	}
	fields["payments_info"] = "[{}]"
	if VerifyNotification(fields, secret) {
		t.Fatal("changed payments_info accepted")
	}
	fields["sign"] = returnDigest
	if !VerifyReturn(fields, secret) || VerifyNotification(fields, secret) {
		t.Fatal("return signature accepted for a notification with extra fields")
	}
	fields["payment_amount"] = "2"
	if VerifyNotification(fields, secret) {
		t.Fatal("changed callback accepted")
	}
	payments, err := ParsePaymentsInfo(`[{"amount":2.00000000,"txid":"a%20b"}]`)
	if err != nil || len(payments) != 1 || payments[0]["amount"] != json.Number("2.00000000") ||
		payments[0]["txid"] != "a%20b" {
		t.Fatalf("unexpected payments: %v, %v", payments, err)
	}
}

func TestRoutes(t *testing.T) {
	base := os.Getenv("MERCHANT_TEST_URL")
	if base == "" {
		t.Skip("run scripts/test.sh to exercise API routes")
	}
	client := NewClient(secret, base, time.Second)
	ctx := context.Background()
	result, err := client.CreateOrder(ctx, payment)
	if err != nil {
		t.Fatal(err)
	}
	order := result["order"].(map[string]any)
	if order["id"] != json.Number("9007199254740993") || order["amount"] != json.Number("12.3400") {
		t.Fatal("lost numeric precision")
	}
	state := Fields{"merchant_id": "85123", "order_number": "testorder123", "amount": "12.3400", "currency": "RUR"}
	if _, err := client.CheckState(ctx, state); err != nil {
		t.Fatal(err)
	}
	state["order_number"] = "missing"
	_, err = client.CheckState(ctx, state)
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.Status != 200 || apiError.Response["error"] != "Unable to verify the order." {
		t.Fatalf("unexpected missing order response: %v", err)
	}
	quote, err := client.GetPaywayInfo(ctx, Fields{
		"merchant_id": "85123", "amount": "12.34", "currency": "RUR", "opt_payway": "visa", "opt_pc": "USD",
	})
	if err != nil || quote["data"].(map[string]any)["order"] != nil {
		t.Fatalf("expected unavailable quote: %v, %v", quote, err)
	}
	_, err = client.CancelOrder(ctx, "85123", "testorder123")
	if !errors.As(err, &apiError) || apiError.Status != 422 || apiError.Response["order"] == nil {
		t.Fatalf("expected 422 and order: %v", err)
	}
	if _, err := client.GetOrderAddress(ctx, Fields{
		"merchantid": "85123", "amount": "12.34", "number": "testorder123", "cryptoCurrency": "USDTt",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListPayways(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetCurrencyRate(ctx, Fields{
		"merchantid": "85123", "amount": "12.34", "currency": "RUR", "opt_pc": "BTC",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTimeout(t *testing.T) {
	base := os.Getenv("MERCHANT_TEST_URL")
	if base == "" {
		t.Skip("run scripts/test.sh to exercise request timeout")
	}
	client := NewClient(secret, base, 25*time.Millisecond)
	_, err := client.ListPayways(context.Background(), Fields{"delay": "1"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout, got %v", err)
	}
}
