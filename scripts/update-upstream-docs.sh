#!/usr/bin/env bash
set -eu

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOC_URL="https://docs.capitalist.net/api/merchants/merchant-doc.html"
OUT_DIR="${ROOT_DIR}/docs/upstream"
HTML_FILE="${OUT_DIR}/merchant-doc.html"
MD_FILE="${OUT_DIR}/merchant-doc.md"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/capitalist-merchant-docs.XXXXXX")"
TMP_HTML_FILE="${TMP_DIR}/merchant-doc.html"
TMP_MD_FILE="${TMP_DIR}/merchant-doc.md"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

if ! command -v pandoc >/dev/null 2>&1; then
  printf 'pandoc is required to convert the upstream HTML snapshot to Markdown\n' >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

curl --fail --location --compressed --silent --show-error "$DOC_URL" -o "$TMP_HTML_FILE"
perl -0pi -e 's/data-cfemail="[^"]+"/data-cfemail="__cfemail__"/g' "$TMP_HTML_FILE"
pandoc -f html -t gfm --wrap=none "$TMP_HTML_FILE" -o "$TMP_MD_FILE"

mv "$TMP_HTML_FILE" "$HTML_FILE"
mv "$TMP_MD_FILE" "$MD_FILE"

printf 'Updated upstream documentation snapshot from %s\n' "$DOC_URL"
printf 'Review changes with: git diff -- docs/upstream\n'
