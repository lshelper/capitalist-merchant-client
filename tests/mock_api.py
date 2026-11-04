"""Local protocol fixture. Never contacts Capitalist."""
import hashlib
import hmac
import json
import sys
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlsplit

SECRET = b"mysecret123"
PAYMENT = {
    "merchantid": "85123", "number": "testorder123", "amount": "12.34",
    "currency": "RUR", "description": "My demo payment.",
}
STATE = {
    "merchant_id": "85123", "order_number": "testorder123",
    "amount": "12.3400", "currency": "RUR",
}
ROUTES = {
    "/merchant/payGate/createorder": ("POST", PAYMENT, "sign"),
    "/merchant/payGate/checkstate": ("GET", STATE, "sign"),
    "/merchant/payGate/paywayinfo": (
        "GET", {"merchant_id": "85123", "amount": "12.34", "currency": "RUR",
                "opt_payway": "visa"}, "sign",
    ),
    "/merchant/payGate/cancel": (
        "POST", {"merchantId": "85123", "orderNumber": "testorder123"}, "signature",
    ),
    "/merchant/payGate/getorderadress": (
        "POST", {"merchantid": "85123", "amount": "12.34", "number": "testorder123",
                 "cryptoCurrency": "USDTt"}, None,
    ),
    "/merchant/payGate/payways": ("GET", {"all": "1"}, None),
    "/merchant/payGate/currencyrate": (
        "GET", {"merchantid": "85123", "amount": "12.34", "currency": "RUR", "opt_pc": "BTC"}, None,
    ),
}
RESPONSES = {
    "createorder": (200, '{"order":{"amount":12.3400,"id":9007199254740993,"label":"quoted \\"x12\\""}}'),
    "checkstate": (200, '{"success":true,"data":{"order":{"paid":true}}}'),
    "paywayinfo": (200, '{"success":true,"data":{"order":null}}'),
    "cancel": (422, '{"order":{"paid":true}}'),
    "getorderadress": (200, '{"address":"mock-address"}'),
    "payways": (200, '{"success":true,"data":{"payways":[]}}'),
    "currencyrate": (200, '{"data":{"buy":[],"sell":[]}}'),
}


class Handler(BaseHTTPRequestHandler):
    def handle_request(self):
        url = urlsplit(self.path)
        if url.path == "/merchant/payGate/payways" and url.query == "delay=1":
            time.sleep(0.25)
            return self.reply(200, '{"data":{"payways":[]}}')
        route = ROUTES.get(url.path)
        if not route or self.command != route[0]:
            return self.reply(404, '{"error":"Unexpected path or method"}')

        _, fields, signature_field = route
        expected = fields.copy()
        if signature_field:
            signed = fields if not url.path.endswith("/paywayinfo") else {
                name: fields[name] for name in ("amount", "currency", "merchant_id", "opt_payway")
            }
            message = ":".join(signed[name] for name in sorted(signed))
            expected[signature_field] = hmac.new(SECRET, message.encode(), hashlib.md5).hexdigest()

        if self.command == "GET":
            query = parse_qs(url.query, keep_blank_values=True)
            if any(len(values) != 1 for values in query.values()):
                return self.reply(400, '{"error":"Duplicate query field"}')
            actual = {name: values[0] for name, values in query.items()}
        else:
            if self.headers.get_content_type() != "application/json":
                return self.reply(400, '{"error":"Expected JSON"}')
            actual = json.loads(self.rfile.read(int(self.headers["Content-Length"])))

        if url.path.endswith("/checkstate") and actual.get("order_number") == "missing":
            missing = {**fields, "order_number": "missing"}
            message = ":".join(missing[name] for name in sorted(missing))
            missing["sign"] = hmac.new(SECRET, message.encode(), hashlib.md5).hexdigest()
            if actual == missing:
                return self.reply(200, '{"success":false,"error":"Unable to verify the order."}')
        if url.path.endswith("/payways") and actual == {
            "merchantid": "85123", "amount": "12.34", "currency": "RUR",
        }:
            return self.reply(*RESPONSES["payways"])
        if url.path.endswith("/checkstate") and actual.get("sign") != expected["sign"] and {
            key: value for key, value in actual.items() if key != "sign"
        } == fields:
            return self.reply(200, '{"success":false,"error":"Unable to verify the signature."}')
        if url.path.endswith("/paywayinfo") and actual.get("opt_pc") == "USD":
            expected["opt_pc"] = "USD"
        if actual != expected:
            return self.reply(400, '{"error":"Unexpected request fields"}')
        return self.reply(*RESPONSES[url.path.rsplit("/", 1)[-1]])

    do_GET = do_POST = handle_request

    def reply(self, status, text):
        body = text.encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        try:
            self.wfile.write(body)
        except (BrokenPipeError, ConnectionResetError):
            pass

    def log_message(self, *_args):
        pass


with ThreadingHTTPServer(("127.0.0.1", int(sys.argv[2]) if len(sys.argv) > 2 else 0), Handler) as server:
    with open(sys.argv[1], "w", encoding="utf-8") as ready:
        ready.write(f"http://127.0.0.1:{server.server_port}\n")
    server.serve_forever()
