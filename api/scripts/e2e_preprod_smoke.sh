#!/usr/bin/env bash

set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"
RUN_SYNC="${RUN_SYNC:-1}"
SYNC_LIMIT="${SYNC_LIMIT:-20}"
WARMUP_LIMIT="${WARMUP_LIMIT:-50}"
SCORE_REGION="${SCORE_REGION:-}"
SCORE_KEYWORDS="${SCORE_KEYWORDS:-}"
EVIDENCE_DIR="${EVIDENCE_DIR:-./artifacts/e2e}"
EVIDENCE_BASENAME="${EVIDENCE_BASENAME:-preprod_smoke}"

mkdir -p "${EVIDENCE_DIR}"
timestamp="$(date -u +"%Y%m%dT%H%M%SZ")"
evidence_file="${EVIDENCE_DIR}/${EVIDENCE_BASENAME}_${timestamp}.json"

live=""
ready=""
profile_payload=""
sync_payload="{}"
list_payload=""
tender_id=""
warmup_payload=""
score_payload=""

write_evidence() {
  local status="$1"
  python3 - \
    "${evidence_file}" \
    "${status}" \
    "${timestamp}" \
    "${API_BASE_URL}" \
    "${RUN_SYNC}" \
    "${SYNC_LIMIT}" \
    "${WARMUP_LIMIT}" \
    "${SCORE_REGION}" \
    "${live}" \
    "${ready}" \
    "${profile_payload}" \
    "${sync_payload}" \
    "${list_payload}" \
    "${warmup_payload}" \
    "${score_payload}" \
    "${tender_id}" <<'PY'
import json
import sys

evidence_file = sys.argv[1]
status = sys.argv[2]
payload = {
    "timestamp_utc": sys.argv[3],
    "api_base_url": sys.argv[4],
    "run_sync": sys.argv[5],
    "sync_limit": sys.argv[6],
    "warmup_limit": sys.argv[7],
    "score_region": sys.argv[8],
    "status": status,
    "health": {
        "live": sys.argv[9],
        "ready": sys.argv[10],
    },
    "responses": {
        "profile": sys.argv[11],
        "sync": sys.argv[12],
        "list": sys.argv[13],
        "warmup": sys.argv[14],
        "score": sys.argv[15],
    },
    "selected_tender_id": sys.argv[16],
}

def decode_maybe_json(raw):
    try:
        return json.loads(raw)
    except Exception:
        return raw

for key in ["profile", "sync", "list", "warmup", "score"]:
    payload["responses"][key] = decode_maybe_json(payload["responses"][key])

with open(evidence_file, "w", encoding="utf-8") as fp:
    json.dump(payload, fp, ensure_ascii=True, indent=2)
PY
}

on_error() {
  write_evidence "failed"
  echo "Smoke E2E failed. Evidence saved at: ${evidence_file}"
}

trap on_error ERR

if [[ -z "${JWT_TOKEN}" ]]; then
  echo "Missing JWT_TOKEN. Export JWT_TOKEN with a valid Bearer token."
  exit 1
fi

request_json() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -sS -X "${method}" "${API_BASE_URL}${path}" \
      -H "Authorization: Bearer ${JWT_TOKEN}" \
      -H "Content-Type: application/json" \
      -d "${body}"
    return
  fi
  curl -sS -X "${method}" "${API_BASE_URL}${path}" \
    -H "Authorization: Bearer ${JWT_TOKEN}"
}

request_public() {
  local path="$1"
  curl -sS "${API_BASE_URL}${path}"
}

extract_json_field() {
  local field="$1"
  python3 - "$field" <<'PY'
import json
import sys
field = sys.argv[1]
payload = json.load(sys.stdin)
value = payload
for part in field.split("."):
    if isinstance(value, dict):
        value = value.get(part)
    else:
        value = None
        break
if value is None:
    print("")
elif isinstance(value, (dict, list)):
    print(json.dumps(value))
else:
    print(value)
PY
}

echo "==> Health checks"
live="$(request_public "/health/live")"
ready="$(request_public "/health/ready")"
echo "live: ${live}"
echo "ready: ${ready}"

echo "==> Profile check"
profile_payload="$(request_json "GET" "/v1/company/profile")"
echo "${profile_payload}" | python3 -m json.tool >/dev/null
echo "profile: ok"

if [[ "${RUN_SYNC}" == "1" ]]; then
  echo "==> Sync tenders"
  sync_payload="$(request_json "GET" "/v1/tenders/sync?limit=${SYNC_LIMIT}")"
  echo "${sync_payload}" | python3 -m json.tool >/dev/null
  echo "sync: ${sync_payload}"
fi

echo "==> List tenders"
list_payload="$(request_json "GET" "/v1/tenders?limit=${SYNC_LIMIT}")"
echo "${list_payload}" | python3 -m json.tool >/dev/null
tender_id="$(
  printf "%s" "${list_payload}" | python3 - <<'PY'
import json, sys
payload = json.load(sys.stdin)
items = payload.get("tenders", [])
if not items:
    print("")
else:
    print(items[0].get("external_id", ""))
PY
)"

if [[ -z "${tender_id}" ]]; then
  echo "No tenders available after list/sync. Cannot continue score validation."
  exit 1
fi

echo "==> Warmup score cache"
warmup_body="{\"limit\":${WARMUP_LIMIT},\"company_region\":\"${SCORE_REGION}\",\"company_keywords\":[${SCORE_KEYWORDS}]}"
if [[ -z "${SCORE_REGION}" ]]; then
  warmup_body="{\"limit\":${WARMUP_LIMIT}}"
fi
warmup_payload="$(request_json "POST" "/v1/tenders/score/warmup" "${warmup_body}")"
echo "${warmup_payload}" | python3 -m json.tool >/dev/null
echo "warmup: ${warmup_payload}"

echo "==> Score check for first tender (${tender_id})"
score_query=""
if [[ -n "${SCORE_REGION}" ]]; then
  score_query="?company_region=${SCORE_REGION}"
fi
score_payload="$(request_json "GET" "/v1/tenders/${tender_id}/score${score_query}")"
echo "${score_payload}" | python3 -m json.tool >/dev/null
echo "score: ${score_payload}"

write_evidence "success"
echo "Evidence saved at: ${evidence_file}"
echo "==> Smoke E2E completed successfully"
