#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/merchant.sh"

export CAPITALIST_MERCHANT_SECRET=mysecret123
payment=(merchantid=85123 number=testorder123 amount=12.34 currency=RUR 'description=My demo payment.')
form="$(merchant_payment_form "${payment[@]}")"
[[ "$form" == *8bb5946e96999f81f8958a8c79d50d4b* ]]
result=(merchant_id=70 order_number=example payment_amount=2.00000000 payment_currency=USD payment_state=success)
signed="$(merchant_python json sign "${result[@]}")"
digest="$(printf '%s' "$signed" | python3 -c 'import json,sys; print(json.load(sys.stdin)["sign"])')"
merchant_verify_notification "${result[@]}" "sign=$digest"
merchant_verify_notification "${result[@]}" client_id= 'payments_info=[]' sign=b970a2012002e3ab3058a9462a3c34d5
if merchant_verify_notification "${result[@]}" client_id= 'payments_info=[{}]' sign=b970a2012002e3ab3058a9462a3c34d5; then exit 1; fi
merchant_verify_return "${result[@]}" client_id= 'payments_info=[]' "sign=$digest"
if merchant_verify_notification "${result[@]}" client_id= 'payments_info=[]' "sign=$digest"; then exit 1; fi
if merchant_verify_notification "${result[@]}" sign=bad; then exit 1; fi

payments="$(printf '%s' '[{"amount":2.00000000,"txid":"a%20b"}]' | merchant_parse_payments_info)"
[[ "$payments" == *'"2.00000000"'* && "$payments" == *'"a%20b"'* ]]

if [[ -n "${MERCHANT_TEST_URL:-}" ]]; then
    export CAPITALIST_MERCHANT_BASE_URL="$MERCHANT_TEST_URL"
    order="$(merchant_create_order "${payment[@]}")"
    [[ "$order" == *12.3400* && "$order" == *9007199254740993* ]]
    merchant_check_state merchant_id=85123 order_number=testorder123 amount=12.3400 currency=RUR >/dev/null
    if merchant_check_state merchant_id=85123 order_number=missing amount=12.3400 currency=RUR >/dev/null; then
        exit 1
    fi
    quote="$(merchant_get_payway_info merchant_id=85123 amount=12.34 currency=RUR opt_payway=visa opt_pc=USD)"
    [[ "$quote" == *'"order":null'* ]]
    status=0
    merchant_cancel_order merchantId=85123 orderNumber=testorder123 >/dev/null || status=$?
    [[ "$status" == 22 ]]
    merchant_get_order_address merchantid=85123 amount=12.34 number=testorder123 cryptoCurrency=USDTt >/dev/null
    merchant_list_payways >/dev/null
    merchant_get_currency_rate merchantid=85123 amount=12.34 currency=RUR opt_pc=BTC >/dev/null
fi

echo 'Bash tests passed'
