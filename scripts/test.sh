#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
tmp="$(mktemp -d)"
python3 tests/mock_api.py "$tmp/url" &
server=$!
cleanup() {
    kill "$server" 2>/dev/null || true
    wait "$server" 2>/dev/null || true
    rm -rf "$tmp"
}
trap cleanup EXIT

for ((i = 0; i < 100; i++)); do
    [[ -s "$tmp/url" ]] && break
    sleep 0.05
done
if [[ ! -s "$tmp/url" ]]; then
    printf 'Mock server did not start\n' >&2
    exit 1
fi
export MERCHANT_TEST_URL
MERCHANT_TEST_URL="$(cat "$tmp/url")"

(cd clients/typescript && npm test)
(cd clients/go && go test ./...)
php -l clients/php/Client.php
php -l clients/php/test.php
php clients/php/test.php
find clients/bash scripts -name '*.sh' -print0 | xargs -0 -n1 bash -n
bash clients/bash/test.sh
