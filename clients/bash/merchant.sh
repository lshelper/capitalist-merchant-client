#!/usr/bin/env bash

merchant_python() {
    python3 - "$@" <<'PY'
import hashlib
import hmac
import json
import os
import sys
import urllib.parse

mode, key, *args = sys.argv[1:]
fields = {}
for arg in args:
    name, sep, value = arg.partition("=")
    if not sep or not name or name in fields:
        sys.exit("Expected unique name=value fields")
    fields[name] = value

def sign(data):
    message = ":".join(data[name] for name in sorted(data) if name != key)
    return hmac.new(os.environ["CAPITALIST_MERCHANT_SECRET"].encode(), message.encode(), hashlib.md5).hexdigest()

if mode in ("verify", "verify-notification"):
    names = ("merchant_id", "order_number", "payment_amount", "payment_currency", "payment_state")
    supplied = fields.get("sign", "")
    if not all(fields.get(name) for name in names) or not supplied.isascii():
        sys.exit(1)
    signed = fields if mode == "verify-notification" else {name: fields[name] for name in names}
    sys.exit(0 if hmac.compare_digest(supplied, sign(signed)) else 1)

if key:
    signed = fields
    if mode == "quote":
        names = ("amount", "currency", "merchant_id", "opt_payway")
        signed = {name: fields[name] for name in names if name in fields}
    fields[key] = sign(signed)
if mode in ("query", "quote"):
    print(urllib.parse.urlencode(fields))
else:
    print(json.dumps(fields, ensure_ascii=False))
PY
}

merchant_request() {
    local method="$1" endpoint="$2" signature="$3" payload response status
    shift 3
    local url="${CAPITALIST_MERCHANT_BASE_URL:-https://api.capitalist.net}"
    url="${url%/}/merchant/payGate/$endpoint"

    local -a args
    args=(-q -sS --max-time "${CAPITALIST_MERCHANT_TIMEOUT:-15}" -H 'Accept: application/json')
    if [[ "$method" == GET ]]; then
        local encoding=query
        if [[ "$endpoint" == paywayinfo ]]; then encoding=quote; fi
        payload="$(merchant_python "$encoding" "$signature" "$@")" || return
        url="$url?$payload"
    else
        payload="$(merchant_python json "$signature" "$@")" || return
        args+=(-H 'Content-Type: application/json' --data-binary @-)
    fi
    response="$(printf '%s' "$payload" | curl "${args[@]}" -w $'\n%{http_code}' "$url")" || return
    status="${response##*$'\n'}"
    payload="${response%$'\n'*}"
    printf '%s\n' "$payload"
    if [[ "$status" != 2[0-9][0-9] ]]; then
        printf 'Merchant API HTTP %s\n' "$status" >&2
        return 22
    fi
    printf '%s' "$payload" | python3 -c 'import json,sys; x=json.load(sys.stdin); sys.exit(0 if isinstance(x,dict) and x.get("success") is not False else 1)'
}

merchant_payment_form() {
    local fields base
    fields="$(merchant_python json sign "$@")" || return
    base="${CAPITALIST_MERCHANT_BASE_URL:-https://api.capitalist.net}"
    printf '%s' "$fields" | python3 -c 'import json,sys; print(json.dumps({"action":sys.argv[1],"method":"POST","fields":json.load(sys.stdin)}))' "${base%/}/merchant/pay?lang=en"
}

merchant_create_order() { merchant_request POST createorder sign "$@"; }
merchant_check_state() { merchant_request GET checkstate sign "$@"; }
merchant_get_payway_info() { merchant_request GET paywayinfo sign "$@"; }
merchant_cancel_order() { merchant_request POST cancel signature "$@"; }
merchant_get_order_address() { merchant_request POST getorderadress '' "$@"; }
merchant_list_payways() {
    if [[ $# -eq 0 ]]; then set -- all=1; fi
    merchant_request GET payways '' "$@"
}
merchant_get_currency_rate() { merchant_request GET currencyrate '' "$@"; }
merchant_verify_return() { merchant_python verify sign "$@"; }
merchant_verify_notification() { merchant_python verify-notification sign "$@"; }

# Reads the already URL-decoded payments_info value from stdin.
merchant_parse_payments_info() {
    python3 -c '
import json
import sys

payments = json.load(sys.stdin, parse_int=str, parse_float=str)
if not isinstance(payments, list) or any(not isinstance(item, dict) for item in payments):
    sys.exit("payments_info must be an array of objects")
print(json.dumps(payments, ensure_ascii=False))
'
}
